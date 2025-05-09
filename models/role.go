package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq" 
	"gorm.io/gorm"
)

// Role representa um papel/função no sistema com permissões específicas
type Role struct {
	ID          uuid.UUID   `gorm:"type:uuid;primary_key" json:"id"`
	Name        string      `gorm:"unique;not null" json:"name"`
	Description string      `json:"description"`
	Permissions pq.StringArray    `gorm:"type:text[]" json:"permissions"`
	Users       []User      `gorm:"many2many:user_roles;" json:"-"`
	Groups      []Group     `gorm:"many2many:group_roles;" json:"-"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// BeforeCreate é um hook GORM que gera um UUID antes de criar um papel
func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// Os papéis padrão do sistema
const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

// GetDefaultRoles retorna os papéis padrão do sistema
func GetDefaultRoles() []Role {
	return []Role{
		{
			Name:        RoleAdmin,
			Description: "Administrador do sistema",
			Permissions: []string{"users:read", "users:write", "groups:read", "groups:write", "roles:read", "roles:write"},
		},
		{
			Name:        RoleUser,
			Description: "Usuário comum",
			Permissions: []string{"profile:read"},
		},
	}
}