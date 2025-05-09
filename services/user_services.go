package services

import (
	"fmt"
	"go-google/models"
	"go-google/repository"

	"github.com/google/uuid"
)

// UserService manipula a lógica de negócio relacionada a usuários
type UserService struct {
	userRepo  *repository.UserRepository
	groupRepo *repository.GroupRepository
}

// NewUserService cria um novo serviço de usuário
func NewUserService(userRepo *repository.UserRepository, groupRepo *repository.GroupRepository) *UserService {
	return &UserService{
		userRepo:  userRepo,
		groupRepo: groupRepo,
	}
}

// GetUserProfile retorna o perfil do usuário
func (s *UserService) GetUserProfile(userID string) (*models.UserResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	// Preparar resposta
	userResponse := &models.UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		Name:        user.Name,
		Picture:     user.Picture,
		Groups:      []string{},
		Roles:       []string{},
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

	return userResponse, nil
}

// ListUsers lista todos os usuários
func (s *UserService) ListUsers() ([]models.UserResponse, error) {
	users, err := s.userRepo.ListAll()
	if err != nil {
		return nil, err
	}

	var userResponses []models.UserResponse
	for _, user := range users {
		// Preparar resposta
		userResponse := models.UserResponse{
			ID:          user.ID,
			Email:       user.Email,
			Name:        user.Name,
			Picture:     user.Picture,
			Groups:      []string{},
			Roles:       []string{},
			Permissions: []string{},
		}

		// Adicionar grupos
		for _, group := range user.Groups {
			userResponse.Groups = append(userResponse.Groups, group.Name)
		}

		// Adicionar papéis
		for _, role := range user.Roles {
			userResponse.Roles = append(userResponse.Roles, role.Name)
		}

		userResponses = append(userResponses, userResponse)
	}

	return userResponses, nil
}

// CreateGroup cria um novo grupo
func (s *UserService) CreateGroup(req models.GroupRequest) (*models.Group, error) {
	// Verificar se o grupo já existe
	existingGroup, err := s.groupRepo.FindByName(req.Name)
	if err == nil && existingGroup != nil {
		return nil, fmt.Errorf("grupo com o nome '%s' já existe", req.Name)
	}

	// Criar grupo
	group := &models.Group{
		Name:        req.Name,
		Description: req.Description,
	}

	// Adicionar papéis se especificados
	if len(req.RoleIDs) > 0 {
		var roles []models.Role
		for _, roleID := range req.RoleIDs {
			id, err := uuid.Parse(roleID)
			if err != nil {
				continue // Ignorar IDs inválidos
			}
			roles = append(roles, models.Role{ID: id})
		}
		group.Roles = roles
	}

	// Salvar grupo
	if err := s.groupRepo.Create(group); err != nil {
		return nil, err
	}

	return group, nil
}

// AssignUserToGroups atribui um usuário a grupos
func (s *UserService) AssignUserToGroups(userID string, groupIDs []string) error {
	return s.userRepo.AssignToGroups(userID, groupIDs)
}