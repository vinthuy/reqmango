package service

import (
	"github.com/reqmanpy/backend-go/internal/common"
	"github.com/reqmanpy/backend-go/internal/dto/request"
	"github.com/reqmanpy/backend-go/internal/dto/response"
	"github.com/reqmanpy/backend-go/internal/model"
	"gorm.io/gorm"
)

type RoleService struct {
	db *gorm.DB
}

func NewRoleService(db *gorm.DB) *RoleService {
	return &RoleService{db: db}
}

// List returns all roles for a workspace or project.
func (s *RoleService) List(scope string, workspaceID, projectID *uint64) ([]response.RoleResponse, error) {
	var roles []model.Role
	query := s.db.Preload("Permissions")
	if workspaceID != nil {
		query = query.Where("workspace_id = ? OR workspace_id IS NULL", *workspaceID)
	} else {
		query = query.Where("workspace_id IS NULL")
	}
	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}
	if err := query.Order("sort_order ASC, level DESC").Find(&roles).Error; err != nil {
		return nil, common.Internal("Failed to list roles")
	}
	result := make([]response.RoleResponse, len(roles))
	for i, r := range roles {
		result[i] = response.ToRoleResponse(&r)
	}
	return result, nil
}

// Create creates a new custom role.
func (s *RoleService) Create(req *request.CreateRoleRequest) (*response.RoleResponse, error) {
	role := model.Role{
		Name:        req.Name,
		Description: req.Description,
		Scope:       req.Scope,
		WorkspaceID: req.WorkspaceID,
		ProjectID:   req.ProjectID,
		Level:       req.Level,
		IsSystem:    false,
	}
	if role.Level == 0 {
		role.Level = 15
	}
	if err := s.db.Create(&role).Error; err != nil {
		return nil, common.Internal("Failed to create role")
	}
	// Assign permissions
	if len(req.Permissions) > 0 {
		var perms []model.Permission
		s.db.Where("id IN ?", req.Permissions).Find(&perms)
		s.db.Model(&role).Association("Permissions").Replace(perms)
	}
	s.db.Preload("Permissions").First(&role, role.ID)
	resp := response.ToRoleResponse(&role)
	return &resp, nil
}

// Update updates a role.
func (s *RoleService) Update(id uint64, req *request.UpdateRoleRequest) (*response.RoleResponse, error) {
	var role model.Role
	if err := s.db.First(&role, id).Error; err != nil {
		return nil, common.NotFound("Role not found")
	}
	if role.IsSystem {
		return nil, common.Forbidden("Cannot modify system roles")
	}
	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Level != nil {
		updates["level"] = *req.Level
	}
	if len(updates) > 0 {
		s.db.Model(&role).Updates(updates)
	}
	if req.Permissions != nil {
		var perms []model.Permission
		if len(req.Permissions) > 0 {
			s.db.Where("id IN ?", req.Permissions).Find(&perms)
		}
		s.db.Model(&role).Association("Permissions").Replace(perms)
	}
	s.db.Preload("Permissions").First(&role, id)
	resp := response.ToRoleResponse(&role)
	return &resp, nil
}

// Delete removes a custom role.
func (s *RoleService) Delete(id uint64) error {
	var role model.Role
	if err := s.db.First(&role, id).Error; err != nil {
		return common.NotFound("Role not found")
	}
	if role.IsSystem {
		return common.Forbidden("Cannot delete system roles")
	}
	// Remove all permissions first
	s.db.Model(&role).Association("Permissions").Clear()
	return s.db.Delete(&role).Error
}

// ListPermissions returns all available permissions.
func (s *RoleService) ListPermissions() ([]response.PermissionResponse, error) {
	var perms []model.Permission
	if err := s.db.Order("scope ASC, resource ASC, action ASC").Find(&perms).Error; err != nil {
		return nil, common.Internal("Failed to list permissions")
	}
	result := make([]response.PermissionResponse, len(perms))
	for i, p := range perms {
		result[i] = response.ToPermissionResponse(&p)
	}
	return result, nil
}

// GetUserPermissions returns all permission codes for a user in a workspace/project context.
func (s *RoleService) GetUserPermissions(userID, workspaceID, projectID uint64) ([]string, error) {
	// Get workspace-level role
	var wsMember model.WorkspaceMember
	if err := s.db.Where("user_id = ? AND workspace_id = ?", userID, workspaceID).First(&wsMember).Error; err != nil {
		return nil, nil // not a member
	}
	// Find default role by level matching
	var wsRole model.Role
	if err := s.db.Preload("Permissions").Where("workspace_id IS NULL AND level = ? AND scope = 'workspace'", wsMember.Role).First(&wsRole).Error; err != nil {
		return nil, nil
	}

	permSet := make(map[string]bool)
	for _, p := range wsRole.Permissions {
		permSet[p.Code] = true
	}

	// If project context, also get project-level role
	if projectID > 0 {
		var prjMember model.ProjectMember
		if err := s.db.Where("user_id = ? AND project_id = ?", userID, projectID).First(&prjMember).Error; err == nil {
			var prjRole model.Role
			if err := s.db.Preload("Permissions").Where("workspace_id IS NULL AND level = ? AND scope = 'project'", prjMember.Role).First(&prjRole).Error; err == nil {
				for _, p := range prjRole.Permissions {
					permSet[p.Code] = true
				}
			}
		}
	}

	// Check for custom role overrides
	var customRoles []model.Role
	s.db.Preload("Permissions").Where(
		"(workspace_id = ? AND scope = 'workspace') OR (project_id = ? AND scope = 'project')",
		workspaceID, projectID,
	).Find(&customRoles)
	// todo: check if user has custom role assignment

	perms := make([]string, 0, len(permSet))
	for code := range permSet {
		perms = append(perms, code)
	}
	return perms, nil
}

// GetUserRoleLevel returns the maximum role level for a user in a workspace/project context.
func (s *RoleService) GetUserRoleLevel(userID, workspaceID, projectID uint64) int {
	maxLevel := 0
	var wsMember model.WorkspaceMember
	if err := s.db.Where("user_id = ? AND workspace_id = ?", userID, workspaceID).First(&wsMember).Error; err == nil {
		if wsMember.Role > maxLevel {
			maxLevel = wsMember.Role
		}
	}
	if projectID > 0 {
		var prjMember model.ProjectMember
		if err := s.db.Where("user_id = ? AND project_id = ?", userID, projectID).First(&prjMember).Error; err == nil {
			if prjMember.Role > maxLevel {
				maxLevel = prjMember.Role
			}
		}
	}
	return maxLevel
}

// HasPermission checks if a user has a specific permission.
func (s *RoleService) HasPermission(userID, workspaceID, projectID uint64, requiredPerm string) bool {
	perms, _ := s.GetUserPermissions(userID, workspaceID, projectID)
	for _, p := range perms {
		if p == requiredPerm {
			return true
		}
	}
	return false
}
