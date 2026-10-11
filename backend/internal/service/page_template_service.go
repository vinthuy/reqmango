package service

import (
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// PageTemplateService handles page template business logic.
type PageTemplateService struct {
	db *gorm.DB
}

// NewPageTemplateService creates a new PageTemplateService.
func NewPageTemplateService(db *gorm.DB) *PageTemplateService {
	return &PageTemplateService{db: db}
}

// checkProjectMembership verifies that the caller is an active member of the
// project. Content operations like page templates are member-scoped.
func (s *PageTemplateService) checkProjectMembership(projectID, userID uint64) error {
	var count int64
	s.db.Model(&model.ProjectMember{}).
		Where("project_id = ? AND user_id = ? AND is_active = ?", projectID, userID, true).
		Count(&count)
	if count == 0 {
		return common.Forbidden("You must be a member of the project to manage page templates")
	}
	return nil
}

func (s *PageTemplateService) projectWorkspaceID(projectID uint64) (uint64, error) {
	var p model.Project
	if err := s.db.Select("id, workspace_id").First(&p, projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, common.NotFound("Project not found")
		}
		return 0, common.Internal("Failed to fetch project")
	}
	return p.WorkspaceID, nil
}

// visibleTo limits templates to those a project can see: its own plus the
// workspace-wide ones (project_id IS NULL) of the project's workspace.
func (s *PageTemplateService) visibleTo(projectID uint64) (*gorm.DB, error) {
	workspaceID, err := s.projectWorkspaceID(projectID)
	if err != nil {
		return nil, err
	}
	return s.db.Where("workspace_id = ? AND (project_id = ? OR project_id IS NULL)", workspaceID, projectID), nil
}

func (s *PageTemplateService) findVisible(id, projectID uint64) (*model.PageTemplate, error) {
	q, err := s.visibleTo(projectID)
	if err != nil {
		return nil, err
	}
	var t model.PageTemplate
	if err := q.First(&t, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Page template not found")
		}
		return nil, common.Internal("Failed to fetch page template")
	}
	return &t, nil
}

// List returns templates visible to a project.
func (s *PageTemplateService) List(projectID uint64) ([]response.PageTemplateResponse, error) {
	q, err := s.visibleTo(projectID)
	if err != nil {
		return nil, err
	}
	var templates []model.PageTemplate
	if err := q.Order("is_default DESC, created_at DESC").Find(&templates).Error; err != nil {
		return nil, common.Internal("Failed to fetch page templates")
	}

	resps := make([]response.PageTemplateResponse, len(templates))
	for i, t := range templates {
		resps[i] = toTemplateResponse(&t)
	}
	return resps, nil
}

// Get returns a single template visible to the project.
func (s *PageTemplateService) Get(id, projectID uint64) (*response.PageTemplateResponse, error) {
	t, err := s.findVisible(id, projectID)
	if err != nil {
		return nil, err
	}
	resp := toTemplateResponse(t)
	return &resp, nil
}

// Create creates a new page template in the project's workspace. A nil
// req.ProjectID makes it workspace-wide; any other value must be projectID.
func (s *PageTemplateService) Create(req *request.PageTemplateCreateRequest, projectID, userID uint64) (*response.PageTemplateResponse, error) {
	if err := s.checkProjectMembership(projectID, userID); err != nil {
		return nil, err
	}
	if req.ProjectID != nil && *req.ProjectID != projectID {
		return nil, common.BadRequest("project_id must match the project in the URL")
	}
	workspaceID, err := s.projectWorkspaceID(projectID)
	if err != nil {
		return nil, err
	}
	// If setting as default, unset others first
	if req.IsDefault {
		s.db.Model(&model.PageTemplate{}).
			Where("workspace_id = ? AND is_default = true", workspaceID).
			Update("is_default", false)
	}

	t := &model.PageTemplate{
		Name:        req.Name,
		Description: req.Description,
		Content:     req.Content,
		ContentJSON: req.ContentJSON,
		IsDefault:   req.IsDefault,
		WorkspaceID: workspaceID,
		ProjectID:   req.ProjectID,
	}
	t.CreatedByID = &userID

	if err := s.db.Create(t).Error; err != nil {
		return nil, common.Internal("Failed to create page template")
	}
	resp := toTemplateResponse(t)
	return &resp, nil
}

// Update updates a page template.
func (s *PageTemplateService) Update(id, projectID, userID uint64, req *request.PageTemplateUpdateRequest) (*response.PageTemplateResponse, error) {
	if err := s.checkProjectMembership(projectID, userID); err != nil {
		return nil, err
	}
	t, err := s.findVisible(id, projectID)
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if req.ContentJSON != nil {
		updates["content_json"] = req.ContentJSON
	}
	if req.IsDefault != nil && *req.IsDefault {
		// Unset other defaults
		s.db.Model(&model.PageTemplate{}).
			Where("workspace_id = ? AND is_default = true AND id != ?", t.WorkspaceID, id).
			Update("is_default", false)
		updates["is_default"] = true
	} else if req.IsDefault != nil {
		updates["is_default"] = false
	}
	updates["updated_by_id"] = userID

	if len(updates) > 0 {
		if err := s.db.Model(t).Updates(updates).Error; err != nil {
			return nil, common.Internal("Failed to update page template")
		}
		s.db.First(t, t.ID)
	}
	resp := toTemplateResponse(t)
	return &resp, nil
}

// Delete deletes a page template.
func (s *PageTemplateService) Delete(id, projectID, callerID uint64) error {
	if err := s.checkProjectMembership(projectID, callerID); err != nil {
		return err
	}
	t, err := s.findVisible(id, projectID)
	if err != nil {
		return err
	}
	if err := s.db.Delete(t).Error; err != nil {
		return common.Internal("Failed to delete page template")
	}
	return nil
}

func toTemplateResponse(t *model.PageTemplate) response.PageTemplateResponse {
	return response.PageTemplateResponse{
		ID:          t.ID,
		Name:        t.Name,
		Description: t.Description,
		Content:     t.Content,
		ContentJSON: t.ContentJSON,
		IsDefault:   t.IsDefault,
		WorkspaceID: t.WorkspaceID,
		ProjectID:   t.ProjectID,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}
