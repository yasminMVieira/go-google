package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// User representa um usuário no sistema
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	Email        string    `gorm:"unique;not null" json:"email"`
	Name         string    `gorm:"not null" json:"name"`
	Picture      string    `json:"picture"`
	GoogleID     string    `gorm:"unique;not null" json:"google_id"`
	RefreshToken string    `gorm:"-" json:"-"`
	Groups       []Group   `gorm:"many2many:user_groups;" json:"groups,omitempty"`
	Roles        []Role    `gorm:"many2many:user_roles;" json:"roles,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// BeforeCreate é um hook GORM que gera um UUID antes de criar um usuário
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// UserResponse é um modelo para resposta de API com informações do usuário
type UserResponse struct {
	ID         uuid.UUID `json:"id"`
	Email      string    `json:"email"`
	Name       string    `json:"name"`
	Picture    string    `json:"picture"`
	Groups     []string  `json:"groups"`
	Roles      []string  `json:"roles"`
	Permissions pq.StringArray `json:"permissions"`
}

// UserWithToken representa um usuário com tokens JWT
type UserWithToken struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
}