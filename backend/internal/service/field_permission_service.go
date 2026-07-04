package service

import (
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

type FieldPermissionService struct{ db *gorm.DB }

func NewFieldPermissionService(db *gorm.DB) *FieldPermissionService {
	return &FieldPermissionService{db: db}
}

// List returns field permissions for a project or workspace.
func (s *FieldPermissionService) List(projectID, workspaceID uint64) ([]model.FieldPermission, error) {
	var perms []model.FieldPermission
	query := s.db.Where("deleted_at IS NULL")
	if projectID > 0 {
		query = query.Where("project_id = ?", projectID)
	} else if workspaceID > 0 {
		query = query.Where("workspace_id = ? AND project_id IS NULL", workspaceID)
	}
	if err := query.Find(&perms).Error; err != nil {
		return nil, common.Internal("Failed to list field permissions")
	}
	return perms, nil
}

// Create adds a new field permission rule.
func (s *FieldPermissionService) Create(perm *model.FieldPermission) error {
	if err := s.db.Create(perm).Error; err != nil {
		return common.Internal("Failed to create field permission")
	}
	return nil
}

// Update modifies an existing field permission rule.
func (s *FieldPermissionService) Update(id uint64, canRead, canWrite *bool) error {
	updates := map[string]interface{}{}
	if canRead != nil {
		updates["can_read"] = *canRead
	}
	if canWrite != nil {
		updates["can_write"] = *canWrite
	}
	if len(updates) == 0 {
		return nil
	}
	r := s.db.Model(&model.FieldPermission{}).Where("id = ?", id).Updates(updates)
	if r.RowsAffected == 0 {
		return common.NotFound("Field permission not found")
	}
	return r.Error
}

// Delete removes a field permission rule.
func (s *FieldPermissionService) Delete(id uint64) error {
	r := s.db.Where("id = ?", id).Delete(&model.FieldPermission{})
	if r.RowsAffected == 0 {
		return common.NotFound("Field permission not found")
	}
	return r.Error
}

// CheckFieldAccess checks if a role has access to a specific field.
func (s *FieldPermissionService) CheckFieldAccess(resource, fieldName string, roleID uint64, projectID uint64) (canRead, canWrite bool) {
	var perm model.FieldPermission
	err := s.db.Where("resource = ? AND field_name = ? AND role_id = ? AND (project_id = ? OR project_id IS NULL) AND deleted_at IS NULL",
		resource, fieldName, roleID, projectID).First(&perm).Error
	if err != nil {
		// Default: allow read, deny write
		return true, false
	}
	return perm.CanRead, perm.CanWrite
}
