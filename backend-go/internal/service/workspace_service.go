package service

import (
	"errors"

	"github.com/reqmanpy/backend-go/internal/common"
	"github.com/reqmanpy/backend-go/internal/dto/request"
	"github.com/reqmanpy/backend-go/internal/dto/response"
	"github.com/reqmanpy/backend-go/internal/model"
	"gorm.io/gorm"
)

type WorkspaceService struct {
	db *gorm.DB
}

func NewWorkspaceService(db *gorm.DB) *WorkspaceService {
	return &WorkspaceService{db: db}
}

// Create creates a new workspace and adds the creator as admin member.
func (s *WorkspaceService) Create(req *request.WorkspaceCreateRequest, ownerID uint64) (*response.WorkspaceResponse, error) {
	// Check slug uniqueness
	var count int64
	s.db.Model(&model.Workspace{}).Where("slug = ?", req.Slug).Count(&count)
	if count > 0 {
		return nil, common.Conflict("Workspace slug already exists")
	}

	timezone := req.Timezone
	if timezone == "" {
		timezone = "UTC"
	}

	workspace := &model.Workspace{
		Name:     req.Name,
		Slug:     req.Slug,
		Timezone: timezone,
		OwnerID:  ownerID,
	}

	if req.OrganizationSize != "" {
		workspace.OrganizationSize = &req.OrganizationSize
	}

	// Use transaction to create workspace + owner membership
	tx := s.db.Begin()
	if err := tx.Create(workspace).Error; err != nil {
		tx.Rollback()
		return nil, common.Internal("Failed to create workspace")
	}

	member := &model.WorkspaceMember{
		WorkspaceID: workspace.ID,
		UserID:      ownerID,
		Role:        common.RoleAdmin,
		IsActive:    true,
	}
	if err := tx.Create(member).Error; err != nil {
		tx.Rollback()
		return nil, common.Internal("Failed to add owner as member")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, common.Internal("Failed to commit transaction")
	}

	return s.buildResponse(workspace.ID)
}

// ListByUser returns all workspaces the user is a member of.
func (s *WorkspaceService) ListByUser(userID uint64) ([]response.WorkspaceLite, error) {
	var memberIDs []uint64
	s.db.Model(&model.WorkspaceMember{}).
		Where("user_id = ? AND is_active = ?", userID, true).
		Pluck("workspace_id", &memberIDs)

	if len(memberIDs) == 0 {
		return []response.WorkspaceLite{}, nil
	}

	var workspaces []model.Workspace
	s.db.Where("id IN ?", memberIDs).Find(&workspaces)

	result := make([]response.WorkspaceLite, len(workspaces))
	for i, ws := range workspaces {
		result[i] = response.WorkspaceLite{
			ID:   ws.ID,
			Name: ws.Name,
			Slug: ws.Slug,
		}
	}
	return result, nil
}

// GetBySlug returns a workspace by its slug.
func (s *WorkspaceService) GetBySlug(slug string) (*response.WorkspaceResponse, error) {
	var workspace model.Workspace
	if err := s.db.Where("slug = ?", slug).First(&workspace).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("Workspace not found")
		}
		return nil, common.Internal("Database error")
	}

	return s.buildResponse(workspace.ID)
}

// GetByID returns a workspace by its ID.
func (s *WorkspaceService) GetByID(workspaceID uint64) (*response.WorkspaceResponse, error) {
	return s.buildResponse(workspaceID)
}

// Update updates a workspace's properties.
func (s *WorkspaceService) Update(workspaceID uint64, req *request.WorkspaceUpdateRequest) (*response.WorkspaceResponse, error) {
	var workspace model.Workspace
	if err := s.db.First(&workspace, workspaceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("Workspace not found")
		}
		return nil, common.Internal("Database error")
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.LogoURL != "" {
		updates["logo_url"] = req.LogoURL
	}
	if req.Timezone != "" {
		updates["timezone"] = req.Timezone
	}

	if len(updates) > 0 {
		if err := s.db.Model(&workspace).Updates(updates).Error; err != nil {
			return nil, common.Internal("Failed to update workspace")
		}
	}

	return s.buildResponse(workspaceID)
}

// Delete performs a soft delete on a workspace.
func (s *WorkspaceService) Delete(workspaceID uint64) error {
	var workspace model.Workspace
	if err := s.db.First(&workspace, workspaceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.NotFound("Workspace not found")
		}
		return common.Internal("Database error")
	}

	if err := s.db.Delete(&workspace).Error; err != nil {
		return common.Internal("Failed to delete workspace")
	}
	return nil
}

// buildResponse constructs a full WorkspaceResponse with counts.
func (s *WorkspaceService) buildResponse(workspaceID uint64) (*response.WorkspaceResponse, error) {
	var workspace model.Workspace
	if err := s.db.Preload("Owner").First(&workspace, workspaceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("Workspace not found")
		}
		return nil, common.Internal("Database error")
	}

	var memberCount int64
	s.db.Model(&model.WorkspaceMember{}).
		Where("workspace_id = ? AND is_active = ?", workspaceID, true).
		Count(&memberCount)

	var projectCount int64
	s.db.Model(&model.Project{}).
		Where("workspace_id = ?", workspaceID).
		Count(&projectCount)

	owner := userToResponse(&workspace.Owner)

	resp := &response.WorkspaceResponse{
		ID:               workspace.ID,
		Name:             workspace.Name,
		Slug:             workspace.Slug,
		LogoURL:          workspace.LogoURL,
		OrganizationSize: workspace.OrganizationSize,
		Timezone:         workspace.Timezone,
		OwnerID:          workspace.OwnerID,
		Owner:            *owner,
		TotalMembers:     memberCount,
		TotalProjects:    projectCount,
		CreatedAt:        workspace.CreatedAt,
		UpdatedAt:        workspace.UpdatedAt,
		CreatedByID:      workspace.CreatedByID,
		UpdatedByID:      workspace.UpdatedByID,
	}

	if workspace.DeletedAt.Valid {
		resp.DeletedAt = &workspace.DeletedAt.Time
		resp.IsDeleted = true
	}

	return resp, nil
}

// ListMembers returns all active members of a workspace
func (s *WorkspaceService) ListMembers(workspaceID uint64) ([]model.WorkspaceMember, error) {
	var members []model.WorkspaceMember
	err := s.db.Preload("User").Where("workspace_id = ? AND is_active = ?", workspaceID, true).
		Find(&members).Error
	if err != nil {
		return nil, err
	}
	if members == nil {
		members = []model.WorkspaceMember{}
	}
	return members, nil
}

// AddMember adds a user to a workspace.
func (s *WorkspaceService) AddMember(workspaceID uint64, req *request.WorkspaceAddMemberRequest, addedBy uint64) (*response.WorkspaceMemberResponse, error) {
	// Verify user exists
	var user model.User
	if err := s.db.First(&user, req.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("User not found")
		}
		return nil, common.Internal("Database error")
	}

	// Check if already a workspace member
	var existing model.WorkspaceMember
	err := s.db.Where("workspace_id = ? AND user_id = ?", workspaceID, req.UserID).First(&existing).Error
	if err == nil {
		return nil, common.Conflict("User is already a workspace member")
	}

	role := req.Role
	if role == 0 {
		role = common.RoleMember
	}

	member := &model.WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      req.UserID,
		Role:        role,
		IsActive:    true,
	}
	if err := s.db.Create(member).Error; err != nil {
		return nil, common.Internal("Failed to add workspace member")
	}

	// Reload with user
	s.db.Preload("User").First(member, member.ID)

	return &response.WorkspaceMemberResponse{
		ID:          member.ID,
		WorkspaceID: member.WorkspaceID,
		UserID:      member.UserID,
		Role:        member.Role,
		IsActive:    member.IsActive,
		User: response.UserLite{
			ID:          member.User.ID,
			DisplayName: member.User.DisplayName,
			Email:       member.User.Email,
		},
		CreatedAt: member.CreatedAt,
		UpdatedAt: member.UpdatedAt,
	}, nil
}

// UpdateMember updates a workspace member's role.
func (s *WorkspaceService) UpdateMember(workspaceID uint64, userID uint64, role int) (*response.WorkspaceMemberResponse, error) {
	var member model.WorkspaceMember
	if err := s.db.Where("workspace_id = ? AND user_id = ?", workspaceID, userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("Workspace member not found")
		}
		return nil, common.Internal("Database error")
	}

	member.Role = role
	s.db.Save(&member)

	s.db.Preload("User").First(&member, member.ID)

	return &response.WorkspaceMemberResponse{
		ID:          member.ID,
		WorkspaceID: member.WorkspaceID,
		UserID:      member.UserID,
		Role:        member.Role,
		IsActive:    member.IsActive,
		User: response.UserLite{
			ID:          member.User.ID,
			DisplayName: member.User.DisplayName,
			Email:       member.User.Email,
		},
		CreatedAt: member.CreatedAt,
		UpdatedAt: member.UpdatedAt,
	}, nil
}

// RemoveMember performs a soft remove on a workspace member (sets is_active=false).
func (s *WorkspaceService) RemoveMember(workspaceID uint64, userID uint64) error {
	result := s.db.Model(&model.WorkspaceMember{}).
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		Update("is_active", false)
	if result.RowsAffected == 0 {
		return common.NotFound("Workspace member not found")
	}
	return nil
}
