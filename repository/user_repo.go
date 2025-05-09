package repository

import (
	"go-google/models"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository manipula operações de banco de dados relacionadas a usuários
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository cria um novo repositório de usuários
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// FindByGoogleID busca um usuário pelo ID do Google
func (r *UserRepository) FindByGoogleID(googleID string) (*models.User, error) {
	var user models.User
	result := r.db.Where("google_id = ?", googleID).Preload("Groups").Preload("Roles").First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

// FindByID busca um usuário pelo ID
func (r *UserRepository) FindByID(id string) (*models.User, error) {
	var user models.User
	result := r.db.Where("id = ?", id).Preload("Groups").Preload("Groups.Roles").Preload("Roles").First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// Create cria um novo usuário
func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// Update atualiza um usuário existente
func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

// ListAll lista todos os usuários
func (r *UserRepository) ListAll() ([]models.User, error) {
	var users []models.User
	if err := r.db.Preload("Groups").Preload("Roles").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// AssignToGroups atribui um usuário a grupos
func (r *UserRepository) AssignToGroups(userID string, groupIDs []string) error {
	// Converter string IDs para UUIDs
	var groups []models.Group
	for _, id := range groupIDs {
		groupID, err := uuid.Parse(id)
		if err != nil {
			return err
		}
		groups = append(groups, models.Group{ID: groupID})
	}

	// Buscar usuário
	user, err := r.FindByID(userID)
	if err != nil {
		return err
	}

	// Associar usuário a grupos
	return r.db.Model(user).Association("Groups").Replace(groups)
}