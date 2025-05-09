package handlers

import (
	// "go-google/models"
	"go-google/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthHandler manipula requisições relacionadas à autenticação
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler cria uma nova instância do manipulador de autenticação
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// GoogleLogin inicia o fluxo de login com Google OAuth
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	url := h.authService.GetGoogleAuthURL()
	c.JSON(http.StatusOK, gin.H{"url": url})
}

// GoogleCallback processa o callback do Google OAuth
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Código de autorização ausente"})
		return
	}

	userWithToken, err := h.authService.ProcessGoogleCallback(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Redirecionar para o frontend com token
	redirectURL := h.authService.GetFrontendRedirectURL(userWithToken.AccessToken, userWithToken.RefreshToken)
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// RefreshToken atualiza o token de acesso usando um token de atualização
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token de atualização inválido"})
		return
	}

	userWithToken, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, userWithToken)
}