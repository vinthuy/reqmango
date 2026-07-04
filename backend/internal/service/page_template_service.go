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

// List returns templates for a project (or workspace).
func (s *PageTemplateService) List(workspaceID uint64, projectID *uint64) ([]response.PageTemplateResponse, error) {
	var templates []model.PageTemplate
	q := s.db.Where("workspace_id = ?", workspaceID)
	if projectID != nil {
		q = q.Where("project_id = ? OR project_id IS NULL", *projectID)
	}
	if err := q.Order("is_default DESC, created_at DESC").Find(&templates).Error; err != nil {
		return nil, common.Internal("Failed to fetch page templates")
	}

	resps := make([]response.PageTemplateResponse, len(templates))
	for i, t := range templates {
		resps[i] = toTemplateResponse(&t)
	}
	return resps, nil
}

// Get returns a single template by ID.
func (s *PageTemplateService) Get(id uint64) (*response.PageTemplateResponse, error) {
	var t model.PageTemplate
	if err := s.db.First(&t, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Page template not found")
		}
		return nil, common.Internal("Failed to fetch page template")
	}
	resp := toTemplateResponse(&t)
	return &resp, nil
}

// Create creates a new page template.
func (s *PageTemplateService) Create(req *request.PageTemplateCreateRequest, workspaceID, userID uint64) (*response.PageTemplateResponse, error) {
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
func (s *PageTemplateService) Update(id, userID uint64, req *request.PageTemplateUpdateRequest) (*response.PageTemplateResponse, error) {
	var t model.PageTemplate
	if err := s.db.First(&t, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Page template not found")
		}
		return nil, common.Internal("Failed to fetch page template")
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
		if err := s.db.Model(&t).Updates(updates).Error; err != nil {
			return nil, common.Internal("Failed to update page template")
		}
		s.db.First(&t, t.ID)
	}
	resp := toTemplateResponse(&t)
	return &resp, nil
}

// Delete deletes a page template.
func (s *PageTemplateService) Delete(id uint64) error {
	result := s.db.Delete(&model.PageTemplate{}, id)
	if result.Error != nil {
		return common.Internal("Failed to delete page template")
	}
	if result.RowsAffected == 0 {
		return common.NotFound("Page template not found")
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
