package service

import (
	"errors"
	"time"

	"github.com/reqmanpy/backend-go/internal/common"
	"github.com/reqmanpy/backend-go/internal/dto/request"
	"github.com/reqmanpy/backend-go/internal/dto/response"
	"github.com/reqmanpy/backend-go/internal/model"
	"gorm.io/gorm"
)

type ProjectService struct {
	db *gorm.DB
}

func NewProjectService(db *gorm.DB) *ProjectService {
	return &ProjectService{db: db}
}

// Create creates a new project and adds the creator as admin member.
func (s *ProjectService) Create(req *request.ProjectCreateRequest, workspaceID, userID uint64) (*response.ProjectResponse, error) {
	// Verify workspace exists
	var workspace model.Workspace
	if err := s.db.First(&workspace, workspaceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("Workspace not found")
		}
		return nil, common.Internal("Database error")
	}

	// Check identifier uniqueness within workspace
	var count int64
	s.db.Model(&model.Project{}).Where("workspace_id = ? AND identifier = ?", workspaceID, req.Identifier).Count(&count)
	if count > 0 {
		return nil, common.Conflict("Project identifier already exists in this workspace")
	}

	timezone := req.Timezone
	if timezone == "" {
		timezone = "UTC"
	}

	isPublic := false
	if req.IsPublic != nil {
		isPublic = *req.IsPublic
	}

	project := &model.Project{
		Name:              req.Name,
		Identifier:        req.Identifier,
		Description:       req.Description,
		IsPublic:          isPublic,
		Timezone:          timezone,
		WorkspaceID:       workspaceID,
		DefaultAssigneeID: req.DefaultAssigneeID,
	}

	tx := s.db.Begin()

	if err := tx.Create(project).Error; err != nil {
		tx.Rollback()
		return nil, common.Internal("Failed to create project")
	}

	// Add creator as project admin member
	member := &model.ProjectMember{
		ProjectID: project.ID,
		UserID:    userID,
		Role:      common.RoleAdmin,
		IsActive:  true,
	}
	if err := tx.Create(member).Error; err != nil {
		tx.Rollback()
		return nil, common.Internal("Failed to add project member")
	}

	// Create default states
	for _, ds := range common.DefaultStates {
		state := &model.State{
			Name:        ds.Name,
			Color:       ds.Color,
			Group:       ds.Group,
			Sequence:    ds.Sequence,
			IsDefault:   ds.IsDefault,
			IsActive:    true,
			ProjectID:   project.ID,
			WorkspaceID: workspaceID,
		}
		if err := tx.Create(state).Error; err != nil {
			tx.Rollback()
			return nil, common.Internal("Failed to create default states")
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, common.Internal("Failed to commit transaction")
	}

	return s.buildResponse(project.ID)
}

// GetByID returns a project by its ID.
func (s *ProjectService) GetByID(projectID uint64) (*response.ProjectResponse, error) {
	return s.buildResponse(projectID)
}

// ListByWorkspace returns all projects in a workspace.
func (s *ProjectService) ListByWorkspace(workspaceID uint64, includeArchived bool, limit, offset int) ([]response.ProjectResponse, error) {
	query := s.db.Where("workspace_id = ?", workspaceID)
	if !includeArchived {
		query = query.Where("archived_at IS NULL")
	}

	var projects []model.Project
	if err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&projects).Error; err != nil {
		return nil, common.Internal("Database error")
	}

	result := make([]response.ProjectResponse, len(projects))
	for i, p := range projects {
		resp, err := s.buildProjectResponse(&p)
		if err != nil {
			return nil, err
		}
		result[i] = *resp
	}
	return result, nil
}

// Update updates a project's properties.
func (s *ProjectService) Update(projectID uint64, req *request.ProjectUpdateRequest) (*response.ProjectResponse, error) {
	var project model.Project
	if err := s.db.First(&project, projectID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("Project not found")
		}
		return nil, common.Internal("Database error")
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.IsPublic != nil {
		updates["is_public"] = *req.IsPublic
	}

	if len(updates) > 0 {
		if err := s.db.Model(&project).Updates(updates).Error; err != nil {
			return nil, common.Internal("Failed to update project")
		}
	}

	return s.buildResponse(projectID)
}

// Delete performs a soft delete on a project.
func (s *ProjectService) Delete(projectID uint64) error {
	var project model.Project
	if err := s.db.First(&project, projectID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.NotFound("Project not found")
		}
		return common.Internal("Database error")
	}
	return s.db.Delete(&project).Error
}

// Archive sets the project's archived_at to now.
func (s *ProjectService) Archive(projectID uint64) (*response.ProjectResponse, error) {
	var project model.Project
	if err := s.db.First(&project, projectID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("Project not found")
		}
		return nil, common.Internal("Database error")
	}

	now := time.Now()
	project.ArchivedAt = &now
	s.db.Save(&project)
	return s.buildResponse(projectID)
}

// Restore clears the project's archived_at.
func (s *ProjectService) Restore(projectID uint64) (*response.ProjectResponse, error) {
	var project model.Project
	if err := s.db.First(&project, projectID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("Project not found")
		}
		return nil, common.Internal("Database error")
	}

	project.ArchivedAt = nil
	s.db.Save(&project)
	return s.buildResponse(projectID)
}

// ListMembers returns all project members.
func (s *ProjectService) ListMembers(projectID uint64, onlyActive bool) ([]response.ProjectMemberResponse, error) {
	var project model.Project
	if err := s.db.First(&project, projectID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("Project not found")
		}
		return nil, common.Internal("Database error")
	}

	query := s.db.Where("project_id = ?", projectID).Preload("User")
	if onlyActive {
		query = query.Where("is_active = ?", true)
	}

	var members []model.ProjectMember
	if err := query.Find(&members).Error; err != nil {
		return nil, common.Internal("Database error")
	}

	result := make([]response.ProjectMemberResponse, len(members))
	for i, m := range members {
		result[i] = response.ProjectMemberResponse{
			ID:        m.ID,
			ProjectID: m.ProjectID,
			UserID:    m.UserID,
			Role:      m.Role,
			IsActive:  m.IsActive,
			User: &response.UserLite{
				ID:          m.User.ID,
				DisplayName: m.User.DisplayName,
				Email:       m.User.Email,
			},
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		}
	}
	return result, nil
}

// AddMember adds a user as a project member.
func (s *ProjectService) AddMember(projectID, userID uint64, role int, addedBy uint64) (*response.ProjectMemberResponse, error) {
	// Verify user exists
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("User not found")
		}
		return nil, common.Internal("Database error")
	}

	// Check if already a member
	var existing model.ProjectMember
	err := s.db.Where("project_id = ? AND user_id = ?", projectID, userID).First(&existing).Error
	if err == nil {
		return nil, common.Conflict("User is already a project member")
	}

	member := &model.ProjectMember{
		ProjectID: projectID,
		UserID:    userID,
		Role:      role,
		IsActive:  true,
		BaseModel: model.BaseModel{CreatedByID: &addedBy},
	}
	if err := s.db.Create(member).Error; err != nil {
		return nil, common.Internal("Failed to add project member")
	}

	// Reload with user
	s.db.Preload("User").First(member, member.ID)

	return &response.ProjectMemberResponse{
		ID:        member.ID,
		ProjectID: member.ProjectID,
		UserID:    member.UserID,
		Role:      member.Role,
		IsActive:  member.IsActive,
		User: &response.UserLite{
			ID:          member.User.ID,
			DisplayName: member.User.DisplayName,
			Email:       member.User.Email,
		},
		CreatedAt: member.CreatedAt,
		UpdatedAt: member.UpdatedAt,
	}, nil
}

// UpdateMember updates a member's role.
func (s *ProjectService) UpdateMember(projectID, userID uint64, role int) (*response.ProjectMemberResponse, error) {
	var member model.ProjectMember
	if err := s.db.Where("project_id = ? AND user_id = ?", projectID, userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("Project member not found")
		}
		return nil, common.Internal("Database error")
	}

	member.Role = role
	s.db.Save(&member)

	s.db.Preload("User").First(&member, member.ID)

	return &response.ProjectMemberResponse{
		ID:        member.ID,
		ProjectID: member.ProjectID,
		UserID:    member.UserID,
		Role:      member.Role,
		IsActive:  member.IsActive,
		User: &response.UserLite{
			ID:          member.User.ID,
			DisplayName: member.User.DisplayName,
			Email:       member.User.Email,
		},
		CreatedAt: member.CreatedAt,
		UpdatedAt: member.UpdatedAt,
	}, nil
}

// RemoveMember removes a user from a project.
func (s *ProjectService) RemoveMember(projectID, userID uint64) error {
	result := s.db.Where("project_id = ? AND user_id = ?", projectID, userID).Delete(&model.ProjectMember{})
	if result.RowsAffected == 0 {
		return common.NotFound("Project member not found")
	}
	return nil
}

// GetStatistics returns project issue statistics.
func (s *ProjectService) GetStatistics(projectID uint64) (*response.ProjectStatistics, error) {
	var project model.Project
	if err := s.db.First(&project, projectID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("Project not found")
		}
		return nil, common.Internal("Database error")
	}

	var totalIssues int64
	s.db.Model(&model.Issue{}).Where("project_id = ?", projectID).Count(&totalIssues)

	var completedIssues int64
	s.db.Model(&model.Issue{}).
		Joins("JOIN states ON states.id = issues.state_id").
		Where("issues.project_id = ? AND states.group = ?", projectID, common.StateGroupCompleted).
		Count(&completedIssues)

	var activeMembers int64
	s.db.Model(&model.ProjectMember{}).
		Where("project_id = ? AND is_active = ?", projectID, true).
		Count(&activeMembers)

	// State breakdown
	stateCounts := make(map[string]int)
	var states []model.State
	s.db.Where("project_id = ? AND is_active = ?", projectID, true).Find(&states)
	for _, state := range states {
		var count int64
		s.db.Model(&model.Issue{}).Where("project_id = ? AND state_id = ?", projectID, state.ID).Count(&count)
		stateCounts[state.Name] = int(count)
	}

	// Priority breakdown
	priorityCounts := make(map[string]int)
	for _, p := range []string{"urgent", "high", "medium", "low", "none"} {
		var count int64
		s.db.Model(&model.Issue{}).Where("project_id = ? AND priority = ?", projectID, p).Count(&count)
		priorityCounts[p] = int(count)
	}

	return &response.ProjectStatistics{
		ProjectID:       project.ID,
		ProjectName:     project.Name,
		TotalIssues:     totalIssues,
		CompletedIssues: completedIssues,
		ActiveMembers:   activeMembers,
		States:          stateCounts,
		Priorities:      priorityCounts,
	}, nil
}

// GetIssuesSummary returns issue counts by state group.
func (s *ProjectService) GetIssuesSummary(projectID uint64) (*response.IssuesSummary, error) {
	var project model.Project
	if err := s.db.First(&project, projectID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("Project not found")
		}
		return nil, common.Internal("Database error")
	}

	summary := &response.IssuesSummary{
		ProjectID:   project.ID,
		ProjectName: project.Name,
		Issues:      make(map[string]int),
	}

	groups := []string{common.StateGroupUnstarted, common.StateGroupStarted, common.StateGroupCompleted, common.StateGroupCancelled}
	for _, group := range groups {
		var count int64
		s.db.Model(&model.Issue{}).
			Joins("JOIN states ON states.id = issues.state_id").
			Where("issues.project_id = ? AND states.group = ?", projectID, group).
			Count(&count)
		summary.Issues[group] = int(count)
	}

	// Map to frontend-expected keys
	summary.Issues["todo"] = summary.Issues[common.StateGroupUnstarted]
	summary.Issues["in_progress"] = summary.Issues[common.StateGroupStarted]
	summary.Issues["done"] = summary.Issues[common.StateGroupCompleted]
	summary.Issues["cancelled"] = summary.Issues[common.StateGroupCancelled]

	return summary, nil
}

// buildResponse is a wrapper that fetches the full model and builds the response.
func (s *ProjectService) buildResponse(projectID uint64) (*response.ProjectResponse, error) {
	var project model.Project
	if err := s.db.Preload("Workspace").Preload("DefaultAssignee").First(&project, projectID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("Project not found")
		}
		return nil, common.Internal("Database error")
	}
	return s.buildProjectResponse(&project)
}

// buildProjectResponse constructs a ProjectResponse from a Project model.
func (s *ProjectService) buildProjectResponse(project *model.Project) (*response.ProjectResponse, error) {
	var issueCount, memberCount int64
	s.db.Model(&model.Issue{}).Where("project_id = ?", project.ID).Count(&issueCount)
	s.db.Model(&model.ProjectMember{}).Where("project_id = ? AND is_active = ?", project.ID, true).Count(&memberCount)

	resp := &response.ProjectResponse{
		ID:                project.ID,
		Name:              project.Name,
		Identifier:        project.Identifier,
		Description:       project.Description,
		IsPublic:          project.IsPublic,
		Timezone:          project.Timezone,
		ArchivedAt:        project.ArchivedAt,
		WorkspaceID:       project.WorkspaceID,
		DefaultAssigneeID: project.DefaultAssigneeID,
		TotalIssues:       issueCount,
		TotalMembers:      memberCount,
		CreatedAt:         project.CreatedAt,
		UpdatedAt:         project.UpdatedAt,
		CreatedByID:       project.CreatedByID,
		UpdatedByID:       project.UpdatedByID,
	}

	if project.Workspace.ID != 0 {
		resp.Workspace = &response.WorkspaceLite{
			ID:   project.Workspace.ID,
			Name: project.Workspace.Name,
			Slug: project.Workspace.Slug,
		}
	}

	if project.DefaultAssignee != nil && project.DefaultAssignee.ID != 0 {
		resp.DefaultAssignee = &response.UserLite{
			ID:          project.DefaultAssignee.ID,
			DisplayName: project.DefaultAssignee.DisplayName,
			Email:       project.DefaultAssignee.Email,
		}
	}

	if project.DeletedAt.Valid {
		resp.DeletedAt = &project.DeletedAt.Time
		resp.IsDeleted = true
	}

	return resp, nil
}
