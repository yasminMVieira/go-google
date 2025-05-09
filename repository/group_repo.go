package repository

import (
	"go-google/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GroupRepository manipula operações de banco de dados relacionadas a grupos
type GroupRepository struct {
	db *gorm.DB
}

// NewGroupRepository cria um novo repositório de grupos
func NewGroupRepository(db *gorm.DB) *GroupRepository {
	return &GroupRepository{
		db: db,
	}
}

// FindByID busca um grupo pelo ID
func (r *GroupRepository) FindByID(id string) (*models.Group, error) {
	var group models.Group
	result := r.db.Where("id = ?", id).Preload("Roles").First(&group)
	if result.Error != nil {
		return nil, result.Error
	}
	return &group, nil
}

// FindByName busca um grupo pelo nome
func (r *GroupRepository) FindByName(name string) (*models.Group, error) {
	var group models.Group
	result := r.db.Where("name = ?", name).Preload("Roles").First(&group)
	if result.Error != nil {
		return nil, result.Error
	}
	return &group, nil
}

// Create cria um novo grupo
func (r *GroupRepository) Create(group *models.Group) error {
	return r.db.Create(group).Error
}

// Update atualiza um grupo existente
func (r *GroupRepository) Update(group *models.Group) error {
	return r.db.Save(group).Error
}

// ListAll lista todos os grupos
func (r *GroupRepository) ListAll() ([]models.Group, error) {
	var groups []models.Group
	if err := r.db.Preload("Roles").Find(&groups).Error; err != nil {
		return nil, err
	}
	return groups, nil
}

// AssignRoles atribui papéis a um grupo
func (r *GroupRepository) AssignRoles(groupID string, roleIDs []string) error {
	// Converter string IDs para UUIDs
	var roles []models.Role
	for _, id := range roleIDs {
		roleID, err := uuid.Parse(id)
		if err != nil {
			continue // Ignorar IDs inválidos
		}
		roles = append(roles, models.Role{ID: roleID})
	}

	// Buscar grupo
	group, err := r.FindByID(groupID)
	if err != nil {
		return err
	}

	// Associar papéis ao grupo
	return r.db.Model(group).Association("Roles").Replace(roles)
}

// FindOrCreateDefaultRoles encontra ou cria os papéis padrão
func (r *GroupRepository) FindOrCreateDefaultRoles() (map[string]models.Role, error) {
	defaultRoles := models.GetDefaultRoles()
	roleMap := make(map[string]models.Role)

	for _, defaultRole := range defaultRoles {
		var role models.Role
		result := r.db.Where("name = ?", defaultRole.Name).First(&role)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				
				// Criar papel se não existir
				if err := r.db.Create(&defaultRole).Error; err != nil {
					return nil, err
				}
				roleMap[defaultRole.Name] = defaultRole
			} else {
				return nil, result.Error
			}
		} else {
			roleMap[role.Name] = role
		}
	}

	return roleMap, nil
}