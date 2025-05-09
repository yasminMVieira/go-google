package handlers

import (
	"go-google/models"
	"go-google/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// UserHandler manipula requisições relacionadas a usuários
type UserHandler struct {
	userService *services.UserService
}

// NewUserHandler cria uma nova instância do manipulador de usuários
func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetProfile retorna o perfil do usuário autenticado
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

	profile, err := h.userService.GetUserProfile(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// ListUsers lista todos os usuários (apenas para administradores)
func (h *UserHandler) ListUsers(c *gin.Context) {
	users, err := h.userService.ListUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

// CreateGroup cria um novo grupo
func (h *UserHandler) CreateGroup(c *gin.Context) {
	var req models.GroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group, err := h.userService.CreateGroup(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, group)
}

// AssignUserToGroup atribui um usuário a um grupo
func (h *UserHandler) AssignUserToGroup(c *gin.Context) {
	userID := c.Param("id")
	var req struct {
		GroupIDs []string `json:"group_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.AssignUserToGroups(userID, req.GroupIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Usuário atribuído aos grupos com sucesso"})
}