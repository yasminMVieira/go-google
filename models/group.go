package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Group representa um grupo de usuários com permissões específicas
type Group struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	Name        string    `gorm:"unique;not null" json:"name"`
	Description string    `json:"description"`
	Users       []User    `gorm:"many2many:user_groups;" json:"-"`
	Roles       []Role    `gorm:"many2many:group_roles;" json:"roles,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BeforeCreate é um hook GORM que gera um UUID antes de criar um grupo
func (g *Group) BeforeCreate(tx *gorm.DB) error {
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	return nil
}

// GroupRequest é um modelo para criar ou atualizar grupos
type GroupRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	RoleIDs     []string `json:"role_ids"`
}