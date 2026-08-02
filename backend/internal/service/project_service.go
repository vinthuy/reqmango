package service

import (
	"errors"
	"time"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

type ProjectService struct {
	db *gorm.DB
}

func NewProjectService(db *gorm.DB) *ProjectService {
	return &ProjectService{db: db}
}

// DB returns the database instance for use in handlers (for security checks).
func (s *ProjectService) DB() *gorm.DB {
	return s.db
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
	if req.TemplateID != nil {
		project.TemplateID = req.TemplateID
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

	// Create default states (skip if using template — template provides them via Apply)
	if req.TemplateID == nil {
		var workspaceStates []model.State
		tx.Where("workspace_id = ? AND project_id IS NULL AND is_active = ?", workspaceID, true).Order("sequence").Find(&workspaceStates)

		if len(workspaceStates) > 0 {
			for _, ws := range workspaceStates {
				projectIDPtr := project.ID
				state := &model.State{
					Name:        ws.Name,
					Color:       ws.Color,
					Group:       ws.Group,
					Sequence:    ws.Sequence,
					IsDefault:   ws.IsDefault,
					IsActive:    true,
					ProjectID:   &projectIDPtr,
					WorkspaceID: workspaceID,
				}
				if err := tx.Create(state).Error; err != nil {
					tx.Rollback()
					return nil, common.Internal("Failed to create workspace states")
				}
			}
		} else {
			for _, ds := range common.DefaultStates {
				projectIDPtr := project.ID
				state := &model.State{
					Name:        ds.Name,
					Color:       ds.Color,
					Group:       ds.Group,
					Sequence:    ds.Sequence,
					IsDefault:   ds.IsDefault,
					IsActive:    true,
					ProjectID:   &projectIDPtr,
					WorkspaceID: workspaceID,
				}
				if err := tx.Create(state).Error; err != nil {
					tx.Rollback()
					return nil, common.Internal("Failed to create default states")
				}
			}
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
	if req.DefaultAssigneeID != nil {
		updates["default_assignee_id"] = *req.DefaultAssigneeID
	}
	if req.ArchivedAt != nil {
		updates["archived_at"] = *req.ArchivedAt
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

// checkProjectAdmin verifies that the caller is an active admin-level member
// of the project. This guards membership mutations against privilege
// escalation by non-admin members.
func (s *ProjectService) checkProjectAdmin(projectID, callerID uint64) error {
	var member model.ProjectMember
	if err := s.db.Where("project_id = ? AND user_id = ? AND is_active = ?", projectID, callerID, true).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.Forbidden("You must be a project admin to manage members")
		}
		return common.Internal("Database error")
	}
	if member.Role < common.RoleAdmin {
		return common.Forbidden("You must be a project admin to manage members")
	}
	return nil
}

// AddMember adds a user as a project member.
func (s *ProjectService) AddMember(projectID, userID uint64, role int, addedBy uint64) (*response.ProjectMemberResponse, error) {
	if err := s.checkProjectAdmin(projectID, addedBy); err != nil {
		return nil, err
	}

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
func (s *ProjectService) UpdateMember(projectID, callerID, userID uint64, role int) (*response.ProjectMemberResponse, error) {
	if err := s.checkProjectAdmin(projectID, callerID); err != nil {
		return nil, err
	}

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
func (s *ProjectService) RemoveMember(projectID, callerID, userID uint64) error {
	if err := s.checkProjectAdmin(projectID, callerID); err != nil {
		return err
	}

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
	if err := s.db.Preload("Workspace").Preload("DefaultAssignee").Preload("ProjectLead").First(&project, projectID).Error; err != nil {
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

	resp.ProjectLeadID = project.ProjectLeadID
	if project.ProjectLead != nil && project.ProjectLead.ID != 0 {
		resp.ProjectLead = &response.UserLite{
			ID:          project.ProjectLead.ID,
			DisplayName: project.ProjectLead.DisplayName,
			Email:       project.ProjectLead.Email,
		}
	}

	if project.DeletedAt.Valid {
		resp.DeletedAt = &project.DeletedAt.Time
		resp.IsDeleted = true
	}

	return resp, nil
}

// UpdateProjectLead sets or clears the project lead.
func (s *ProjectService) UpdateProjectLead(projectID uint64, userID *uint64) (*response.ProjectResponse, error) {
	var project model.Project
	if err := s.db.First(&project, projectID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("Project not found")
		}
		return nil, common.Internal("Database error")
	}

	project.ProjectLeadID = userID
	if err := s.db.Save(&project).Error; err != nil {
		return nil, common.Internal("Failed to update project lead")
	}

	return s.buildResponse(projectID)
}

// ListSubscribers returns all subscribers for a project.
func (s *ProjectService) ListSubscribers(projectID uint64) ([]response.ProjectSubscriberResponse, error) {
	var project model.Project
	if err := s.db.First(&project, projectID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("Project not found")
		}
		return nil, common.Internal("Database error")
	}

	var subs []model.ProjectSubscriber
	if err := s.db.Where("project_id = ?", projectID).Preload("User").Find(&subs).Error; err != nil {
		return nil, common.Internal("Database error")
	}

	result := make([]response.ProjectSubscriberResponse, len(subs))
	for i, s := range subs {
		result[i] = response.ProjectSubscriberResponse{
			ID:        s.ID,
			ProjectID: s.ProjectID,
			UserID:    s.UserID,
			User: &response.UserLite{
				ID:          s.User.ID,
				DisplayName: s.User.DisplayName,
				Email:       s.User.Email,
			},
			CreatedAt: s.CreatedAt,
			UpdatedAt: s.UpdatedAt,
		}
	}
	return result, nil
}

// AddSubscriber adds a user as a project subscriber.
func (s *ProjectService) AddSubscriber(projectID, userID uint64) (*response.ProjectSubscriberResponse, error) {
	// Verify user exists
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("User not found")
		}
		return nil, common.Internal("Database error")
	}

	// Check if already subscribed
	var existing model.ProjectSubscriber
	err := s.db.Where("project_id = ? AND user_id = ?", projectID, userID).First(&existing).Error
	if err == nil {
		return nil, common.Conflict("User is already a subscriber")
	}

	sub := &model.ProjectSubscriber{
		ProjectID: projectID,
		UserID:    userID,
	}
	if err := s.db.Create(sub).Error; err != nil {
		return nil, common.Internal("Failed to add subscriber")
	}

	s.db.Preload("User").First(sub, sub.ID)

	return &response.ProjectSubscriberResponse{
		ID:        sub.ID,
		ProjectID: sub.ProjectID,
		UserID:    sub.UserID,
		User: &response.UserLite{
			ID:          sub.User.ID,
			DisplayName: sub.User.DisplayName,
			Email:       sub.User.Email,
		},
		CreatedAt: sub.CreatedAt,
		UpdatedAt: sub.UpdatedAt,
	}, nil
}

// RemoveSubscriber removes a user from project subscribers.
func (s *ProjectService) RemoveSubscriber(projectID, userID uint64) error {
	result := s.db.Where("project_id = ? AND user_id = ?", projectID, userID).Delete(&model.ProjectSubscriber{})
	if result.RowsAffected == 0 {
		return common.NotFound("Subscriber not found")
	}
	return nil
}
