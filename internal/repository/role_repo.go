package repository

import (
	"time"

	"github.com/UtopiaXC/UtopiaDaily/internal/model"
	"gorm.io/gorm"
)

type RoleWithCount struct {
	ID          uint     `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	IsSystem    bool     `json:"is_system"`
	UserCount   int64    `json:"user_count"`
	Permissions []string `json:"permissions"`
}

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) ListPermissions() ([]model.Permission, error) {
	var perms []model.Permission
	err := r.db.Order("group_code ASC, code ASC").Find(&perms).Error
	return perms, err
}

func (r *RoleRepository) ListRolesWithCounts() ([]RoleWithCount, error) {
	var roles []model.Role
	if err := r.db.Preload("Permissions").Order("id ASC").Find(&roles).Error; err != nil {
		return nil, err
	}

	type countRow struct {
		RoleID uint
		Total  int64
	}
	var counts []countRow
	if err := r.db.Model(&model.User{}).Select("role_id, COUNT(*) AS total").Group("role_id").Scan(&counts).Error; err != nil {
		return nil, err
	}
	countMap := make(map[uint]int64, len(counts))
	for _, c := range counts {
		countMap[c.RoleID] = c.Total
	}

	result := make([]RoleWithCount, 0, len(roles))
	for _, role := range roles {
		codes := make([]string, 0, len(role.Permissions))
		for _, p := range role.Permissions {
			codes = append(codes, p.Code)
		}
		result = append(result, RoleWithCount{
			ID:          role.ID,
			Name:        role.Name,
			Description: role.Description,
			IsSystem:    role.IsSystem,
			UserCount:   countMap[role.ID],
			Permissions: codes,
		})
	}
	return result, nil
}

func (r *RoleRepository) FindByID(id uint) (*model.Role, error) {
	var role model.Role
	err := r.db.Preload("Permissions").Where("id = ?", id).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepository) FindByName(name string) (*model.Role, error) {
	var role model.Role
	err := r.db.Preload("Permissions").Where("name = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepository) CountUsersInRole(roleID uint) (int64, error) {
	var cnt int64
	err := r.db.Model(&model.User{}).Where("role_id = ?", roleID).Count(&cnt).Error
	return cnt, err
}

func (r *RoleRepository) FindPermissionsByCodes(codes []string) ([]model.Permission, error) {
	if len(codes) == 0 {
		return nil, nil
	}
	var perms []model.Permission
	err := r.db.Where("code IN ?", codes).Find(&perms).Error
	return perms, err
}

func (r *RoleRepository) CreateRole(name, description string, perms []model.Permission) (*model.Role, error) {
	now := time.Now().UTC()
	role := &model.Role{
		Name:        name,
		Description: description,
		IsSystem:    false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(role).Error; err != nil {
			return err
		}
		if len(perms) > 0 {
			if err := tx.Model(role).Association("Permissions").Replace(perms); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *RoleRepository) UpdateRole(roleID uint, name, description string, perms []model.Permission) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var role model.Role
		if err := tx.Where("id = ?", roleID).First(&role).Error; err != nil {
			return err
		}
		if err := tx.Model(&role).Updates(map[string]any{
			"name":        name,
			"description": description,
			"updated_at":  time.Now().UTC(),
		}).Error; err != nil {
			return err
		}
		return tx.Model(&role).Association("Permissions").Replace(perms)
	})
}

func (r *RoleRepository) DeleteRole(roleID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&model.RolePermission{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Role{}, roleID).Error
	})
}
