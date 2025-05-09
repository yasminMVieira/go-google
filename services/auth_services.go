package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-google/config"
	"go-google/models"
	"go-google/repository"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	// "github.com/google/uuid"
)

// AuthService manipula a lógica de negócio relacionada à autenticação
type AuthService struct {
	config    *config.Config
	userRepo  *repository.UserRepository
	groupRepo *repository.GroupRepository
}

// NewAuthService cria um novo serviço de autenticação
func NewAuthService(config *config.Config, userRepo *repository.UserRepository, groupRepo *repository.GroupRepository) *AuthService {
	return &AuthService{
		config:    config,
		userRepo:  userRepo,
		groupRepo: groupRepo,
	}
}

// GetGoogleAuthURL retorna a URL para iniciar o fluxo de autenticação com Google
func (s *AuthService) GetGoogleAuthURL() string {
	googleConfig := config.GetGoogleOAuthConfig(s.config)
	return googleConfig.AuthCodeURL("state", GetAuthURLOptions()...)
}

// ProcessGoogleCallback processa o callback do Google OAuth
func (s *AuthService) ProcessGoogleCallback(code string) (*models.UserWithToken, error) {
	// Trocar código por token
	googleConfig := config.GetGoogleOAuthConfig(s.config)
	token, err := googleConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("erro ao trocar código por token: %w", err)
	}

	// Buscar informações do usuário do Google
	userInfo, err := s.fetchGoogleUserInfo(token.AccessToken)
	if err != nil {
		return nil, err
	}

	// Verificar se o usuário já existe
	user, err := s.userRepo.FindByGoogleID(userInfo.ID)
	if err != nil {
		return nil, err
	}

	// Carregar papéis padrão
	defaultRoles, err := s.groupRepo.FindOrCreateDefaultRoles()
	if err != nil {
		return nil, err
	}

	// Criar ou atualizar usuário
	if user == nil {
		// Novo usuário
		user = &models.User{
			GoogleID: userInfo.ID,
			Email:    userInfo.Email,
			Name:     userInfo.Name,
			Picture:  userInfo.Picture,
			Roles:    []models.Role{defaultRoles[models.RoleUser]},
		}

		// se for o primeiro usuário, atribui role admin
		existingUsers, _ := s.userRepo.ListAll()
		if len(existingUsers) == 0 {
			user.Roles = []models.Role{defaultRoles[models.RoleAdmin]}
		} else {
			user.Roles = []models.Role{defaultRoles[models.RoleUser]}
		}
		
		if err := s.userRepo.Create(user); err != nil {
			return nil, err
		}
	} else {
		// Atualizar usuário existente
		user.Email = userInfo.Email
		user.Name = userInfo.Name
		user.Picture = userInfo.Picture
		if err := s.userRepo.Update(user); err != nil {
			return nil, err
		}
	}

	// Gerar tokens JWT
	accessToken, refreshToken, expiresIn, err := s.generateTokens(user)
	if err != nil {
		return nil, err
	}

	// Preparar resposta
	userResponse := models.UserResponse{
		ID:       user.ID,
		Email:    user.Email,
		Name:     user.Name,
		Picture:  user.Picture,
		Groups:   []string{},
		Roles:    []string{},
		Permissions: []string{},
	}

	// Adicionar grupos
	for _, group := range user.Groups {
		userResponse.Groups = append(userResponse.Groups, group.Name)
	}

	// Adicionar papéis diretos e papéis dos grupos
	roleMap := make(map[string]bool)
	permMap := make(map[string]bool)

	// Papéis diretos
	for _, role := range user.Roles {
		roleMap[role.Name] = true
		for _, perm := range role.Permissions {
			permMap[perm] = true
		}
	}

	// Papéis dos grupos
	for _, group := range user.Groups {
		for _, role := range group.Roles {
			roleMap[role.Name] = true
			for _, perm := range role.Permissions {
				permMap[perm] = true
			}
		}
	}

	// Converter mapas em slices
	for role := range roleMap {
		userResponse.Roles = append(userResponse.Roles, role)
	}
	for perm := range permMap {
		userResponse.Permissions = append(userResponse.Permissions, perm)
	}

	return &models.UserWithToken{
		User:         userResponse,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

// GoogleUserInfo representa as informações do usuário retornadas pela API do Google
type GoogleUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

// fetchGoogleUserInfo busca informações do usuário do Google
func (s *AuthService) fetchGoogleUserInfo(accessToken string) (*GoogleUserInfo, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar informações do usuário: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("resposta não-OK da API do Google: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta: %w", err)
	}

	var userInfo GoogleUserInfo
	if err := json.Unmarshal(data, &userInfo); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	return &userInfo, nil
}

// RefreshToken atualiza o token de acesso usando um token de atualização
func (s *AuthService) RefreshToken(refreshToken string) (*models.UserWithToken, error) {
	// Verificar token de atualização
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("método de assinatura inválido")
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("token de atualização inválido")
	}

	// Extrair claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("falha ao extrair claims")
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return nil, errors.New("ID de usuário inválido no token")
	}

	// Verificar tipo de token
	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "refresh" {
		return nil, errors.New("tipo de token inválido")
	}

	// Buscar usuário
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	// Gerar novos tokens
	accessToken, newRefreshToken, expiresIn, err := s.generateTokens(user)
	if err != nil {
		return nil, err
	}

	// Preparar resposta
	userResponse := models.UserResponse{
		ID:       user.ID,
		Email:    user.Email,
		Name:     user.Name,
		Picture:  user.Picture,
		Groups:   []string{},
		Roles:    []string{},
		Permissions: []string{},
	}

	// Adicionar grupos
	for _, group := range user.Groups {
		userResponse.Groups = append(userResponse.Groups, group.Name)
	}

	// Adicionar papéis diretos e papéis dos grupos
	roleMap := make(map[string]bool)
	permMap := make(map[string]bool)

	// Papéis diretos
	for _, role := range user.Roles {
		roleMap[role.Name] = true
		for _, perm := range role.Permissions {
			permMap[perm] = true
		}
	}

	// Papéis dos grupos
	for _, group := range user.Groups {
		for _, role := range group.Roles {
			roleMap[role.Name] = true
			for _, perm := range role.Permissions {
				permMap[perm] = true
			}
		}
	}

	// Converter mapas em slices
	for role := range roleMap {
		userResponse.Roles = append(userResponse.Roles, role)
	}
	for perm := range permMap {
		userResponse.Permissions = append(userResponse.Permissions, perm)
	}

	return &models.UserWithToken{
		User:         userResponse,
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

// generateTokens gera tokens JWT para o usuário
func (s *AuthService) generateTokens(user *models.User) (accessToken string, refreshToken string, expiresIn int64, err error) {
	// Calcular duração dos tokens
	accessTokenExpiry := time.Now().Add(15 * time.Minute)
	refreshTokenExpiry := time.Now().Add(7 * 24 * time.Hour)
	expiresIn = int64(accessTokenExpiry.Sub(time.Now()).Seconds())

	// Coletar papéis e permissões
	roles := []string{}
	permissions := []string{}

	// Papéis diretos
	for _, role := range user.Roles {
		roles = append(roles, role.Name)
		permissions = append(permissions, role.Permissions...)
	}

	// Papéis dos grupos
	for _, group := range user.Groups {
		for _, role := range group.Roles {
			roles = append(roles, role.Name)
			permissions = append(permissions, role.Permissions...)
		}
	}

	// Remover duplicatas
	uniqueRoles := make(map[string]bool)
	uniquePermissions := make(map[string]bool)

	var finalRoles []string
	var finalPermissions []string

	for _, role := range roles {
		if !uniqueRoles[role] {
			uniqueRoles[role] = true
			finalRoles = append(finalRoles, role)
		}
	}

	for _, perm := range permissions {
		if !uniquePermissions[perm] {
			uniquePermissions[perm] = true
			finalPermissions = append(finalPermissions, perm)
		}
	}

	// Gerar token de acesso
	accessClaims := jwt.MapClaims{
		"sub":         user.ID.String(),
		"email":       user.Email,
		"name":        user.Name,
		"roles":       finalRoles,
		"permissions": finalPermissions,
		"exp":         accessTokenExpiry.Unix(),
		"iat":         time.Now().Unix(),
		"type":        "access",
	}

	accessJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = accessJWT.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return "", "", 0, err
	}

	// Gerar token de atualização
	refreshClaims := jwt.MapClaims{
		"sub":  user.ID.String(),
		"exp":  refreshTokenExpiry.Unix(),
		"iat":  time.Now().Unix(),
		"type": "refresh",
	}

	refreshJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refreshJWT.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return "", "", 0, err
	}

	return accessToken, refreshToken, expiresIn, nil
}

// GetFrontendRedirectURL gera a URL para redirecionar para o frontend com tokens
func (s *AuthService) GetFrontendRedirectURL(accessToken, refreshToken string) string {
	baseURL := s.config.FrontendURL
	params := url.Values{}
	params.Add("access_token", accessToken)
	params.Add("refresh_token", refreshToken)
	
	if strings.Contains(baseURL, "?") {
		return baseURL + "&" + params.Encode()
	}
	return baseURL + "?" + params.Encode()
}

// GetAuthURLOptions retorna opções para a URL de autenticação do Google
func GetAuthURLOptions() []oauth2.AuthCodeOption {
	return []oauth2.AuthCodeOption{
		oauth2.AccessTypeOffline,
		oauth2.ApprovalForce,
	}
}