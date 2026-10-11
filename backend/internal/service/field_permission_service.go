package service

import (
	"errors"
	"strconv"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

type FieldPermissionService struct{ db *gorm.DB }

func NewFieldPermissionService(db *gorm.DB) *FieldPermissionService {
	return &FieldPermissionService{db: db}
}

// checkWorkspaceAdmin verifies that the caller is an active admin-level member
// of the workspace. Guards mutations against privilege escalation.
func (s *FieldPermissionService) checkWorkspaceAdmin(workspaceID, callerID uint64) error {
	var member model.WorkspaceMember
	if err := s.db.Where("workspace_id = ? AND user_id = ? AND is_active = ?", workspaceID, callerID, true).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.Forbidden("You must be a workspace admin to manage field permissions")
		}
		return common.Internal("Database error")
	}
	if member.Role < common.RoleAdmin {
		return common.Forbidden("You must be a workspace admin to manage field permissions")
	}
	return nil
}

// checkProjectAdmin verifies that the caller is an active admin-level member
// of the project. Guards mutations against privilege escalation.
func (s *FieldPermissionService) checkProjectAdmin(projectID, callerID uint64) error {
	var member model.ProjectMember
	if err := s.db.Where("project_id = ? AND user_id = ? AND is_active = ?", projectID, callerID, true).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.Forbidden("You must be a project admin to manage field permissions")
		}
		return common.Internal("Database error")
	}
	if member.Role < common.RoleAdmin {
		return common.Forbidden("You must be a project admin to manage field permissions")
	}
	return nil
}

// assertCallerCanManagePerm authorizes the caller against the scope of the perm.
func (s *FieldPermissionService) assertCallerCanManagePerm(perm *model.FieldPermission, callerID uint64) error {
	if perm.WorkspaceID > 0 {
		return s.checkWorkspaceAdmin(perm.WorkspaceID, callerID)
	}
	if perm.ProjectID > 0 {
		return s.checkProjectAdmin(perm.ProjectID, callerID)
	}
	return common.BadRequest("Field permission must specify workspace_id or project_id")
}

// List returns field permissions for a project or workspace.
// Both IDs zero is rejected so callers cannot dump every tenant's rows.
func (s *FieldPermissionService) List(projectID, workspaceID uint64) ([]model.FieldPermission, error) {
	if projectID == 0 && workspaceID == 0 {
		return nil, common.BadRequest("workspace_id or project_id is required")
	}
	var perms []model.FieldPermission
	query := s.db.Where("deleted_at IS NULL")
	if projectID > 0 {
		query = query.Where("project_id = ?", projectID)
	} else {
		query = query.Where("workspace_id = ?", workspaceID)
	}
	if err := query.Find(&perms).Error; err != nil {
		return nil, common.Internal("Failed to list field permissions")
	}
	return perms, nil
}

// ResolveWorkspaceID resolves a numeric id or slug to a workspace id.
func (s *FieldPermissionService) ResolveWorkspaceID(wsParam string) (uint64, error) {
	if id, err := strconv.ParseUint(wsParam, 10, 64); err == nil {
		return id, nil
	}
	var ws model.Workspace
	if err := s.db.Select("id").Where("slug = ?", wsParam).First(&ws).Error; err != nil {
		return 0, common.NotFound("Workspace not found")
	}
	return ws.ID, nil
}

// Create adds a new field permission rule.
func (s *FieldPermissionService) Create(perm *model.FieldPermission, callerID uint64) error {
	if err := s.assertCallerCanManagePerm(perm, callerID); err != nil {
		return err
	}
	if err := s.db.Create(perm).Error; err != nil {
		return common.Internal("Failed to create field permission")
	}
	return nil
}

// Update modifies an existing field permission rule.
func (s *FieldPermissionService) Update(id uint64, callerID uint64, canRead, canWrite *bool) error {
	var perm model.FieldPermission
	if err := s.db.First(&perm, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.NotFound("Field permission not found")
		}
		return common.Internal("Database error")
	}
	if err := s.assertCallerCanManagePerm(&perm, callerID); err != nil {
		return err
	}
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
func (s *FieldPermissionService) Delete(id uint64, callerID uint64) error {
	var perm model.FieldPermission
	if err := s.db.First(&perm, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.NotFound("Field permission not found")
		}
		return common.Internal("Database error")
	}
	if err := s.assertCallerCanManagePerm(&perm, callerID); err != nil {
		return err
	}
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
