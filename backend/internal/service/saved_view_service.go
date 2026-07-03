package service

import (
	"encoding/json"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// SavedViewService handles saved view business logic.
type SavedViewService struct {
	db *gorm.DB
}

// NewSavedViewService creates a new SavedViewService.
func NewSavedViewService(db *gorm.DB) *SavedViewService {
	return &SavedViewService{db: db}
}

// List returns all saved views for a project accessible to a user.
func (s *SavedViewService) List(projectID, userID uint64) ([]response.SavedViewResponse, error) {
	var views []model.SavedView
	if err := s.db.Where("project_id = ? AND (owner_id = ? OR is_shared = ?)", projectID, userID, true).
		Order("is_default DESC, created_at ASC").Find(&views).Error; err != nil {
		return nil, common.Internal("Failed to fetch saved views")
	}

	resps := make([]response.SavedViewResponse, len(views))
	for i, v := range views {
		resps[i] = viewToResponse(&v)
	}
	return resps, nil
}

// Get returns a single saved view.
func (s *SavedViewService) Get(id, projectID, userID uint64) (*response.SavedViewResponse, error) {
	var v model.SavedView
	if err := s.db.Where("id = ? AND project_id = ? AND (owner_id = ? OR is_shared = ?)", id, projectID, userID, true).
		First(&v).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Saved view not found")
		}
		return nil, common.Internal("Failed to fetch saved view")
	}
	resp := viewToResponse(&v)
	return &resp, nil
}

// Create creates a new saved view.
func (s *SavedViewService) Create(req *request.SavedViewCreateRequest, projectID, userID uint64) (*response.SavedViewResponse, error) {
	v := &model.SavedView{
		Name:        req.Name,
		Description: req.Description,
		ViewType:    req.ViewType,
		Filters:     normalizeJSON(req.Filters),
		RQL:         req.RQL,
		SortConfig:  normalizeJSON(req.SortConfig),
		Columns:     normalizeJSON(req.Columns),
		GroupBy:     req.GroupBy,
		IsShared:    req.IsShared,
		OwnerID:     userID,
		ProjectID:   projectID,
	}
	if v.ViewType == "" {
		v.ViewType = "list"
	}

	if err := s.db.Create(v).Error; err != nil {
		return nil, common.Internal("Failed to create saved view")
	}
	resp := viewToResponse(v)
	return &resp, nil
}

// Update updates a saved view.
func (s *SavedViewService) Update(id, projectID, userID uint64, req *request.SavedViewUpdateRequest) (*response.SavedViewResponse, error) {
	var v model.SavedView
	if err := s.db.Where("id = ? AND project_id = ? AND owner_id = ?", id, projectID, userID).First(&v).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Saved view not found or access denied")
		}
		return nil, common.Internal("Failed to fetch saved view")
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = req.Description
	}
	if req.ViewType != nil {
		updates["view_type"] = *req.ViewType
	}
	if req.Filters != nil {
		updates["filters"] = normalizeJSON(*req.Filters)
	}
	if req.RQL != nil {
		updates["rql"] = *req.RQL
	}
	if req.SortConfig != nil {
		updates["sort_config"] = normalizeJSON(*req.SortConfig)
	}
	if req.Columns != nil {
		updates["columns"] = normalizeJSON(*req.Columns)
	}
	if req.GroupBy != nil {
		updates["group_by"] = req.GroupBy
	}
	if req.IsShared != nil {
		updates["is_shared"] = *req.IsShared
	}

	if len(updates) > 0 {
		if err := s.db.Model(&v).Updates(updates).Error; err != nil {
			return nil, common.Internal("Failed to update saved view")
		}
		// Reload to get updated values
		s.db.First(&v, v.ID)
	}
	resp := viewToResponse(&v)
	return &resp, nil
}

// Delete deletes a saved view.
func (s *SavedViewService) Delete(id, projectID, userID uint64) error {
	result := s.db.Where("id = ? AND project_id = ? AND owner_id = ?", id, projectID, userID).Delete(&model.SavedView{})
	if result.Error != nil {
		return common.Internal("Failed to delete saved view")
	}
	if result.RowsAffected == 0 {
		return common.NotFound("Saved view not found or access denied")
	}
	return nil
}

// SetDefault sets a view as the default for the user in this project.
func (s *SavedViewService) SetDefault(id, projectID, userID uint64) (*response.SavedViewResponse, error) {
	var v model.SavedView
	if err := s.db.Where("id = ? AND project_id = ? AND owner_id = ?", id, projectID, userID).First(&v).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Saved view not found or access denied")
		}
		return nil, common.Internal("Failed to fetch saved view")
	}

	// Unset previous defaults for this user+project
	s.db.Model(&model.SavedView{}).
		Where("project_id = ? AND owner_id = ? AND is_default = ?", projectID, userID, true).
		Update("is_default", false)

	// Set new default
	s.db.Model(&v).Update("is_default", true)
	v.IsDefault = true
	resp := viewToResponse(&v)
	return &resp, nil
}

// Duplicate duplicates a saved view.
func (s *SavedViewService) Duplicate(id, projectID, userID uint64) (*response.SavedViewResponse, error) {
	var src model.SavedView
	if err := s.db.Where("id = ? AND project_id = ? AND (owner_id = ? OR is_shared = ?)", id, projectID, userID, true).
		First(&src).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Saved view not found")
		}
		return nil, common.Internal("Failed to fetch saved view")
	}

	clone := model.SavedView{
		Name:        src.Name + " (Copy)",
		Description: src.Description,
		ViewType:    src.ViewType,
		Filters:     src.Filters,
		RQL:         src.RQL,
		SortConfig:  src.SortConfig,
		Columns:     src.Columns,
		GroupBy:     src.GroupBy,
		IsDefault:   false,
		IsShared:    false,
		OwnerID:     userID,
		ProjectID:   projectID,
	}
	if err := s.db.Create(&clone).Error; err != nil {
		return nil, common.Internal("Failed to duplicate saved view")
	}
	resp := viewToResponse(&clone)
	return &resp, nil
}

// ==================== Helpers ====================

func normalizeJSON(raw json.RawMessage) json.RawMessage {
	if raw == nil || string(raw) == "" || string(raw) == "null" {
		return json.RawMessage("{}")
	}
	return raw
}

func viewToResponse(v *model.SavedView) response.SavedViewResponse {
	return response.SavedViewResponse{
		ID:          v.ID,
		Name:        v.Name,
		Description: v.Description,
		ViewType:    v.ViewType,
		Filters:     v.Filters,
		RQL:         v.RQL,
		SortConfig:  v.SortConfig,
		Columns:     v.Columns,
		GroupBy:     v.GroupBy,
		IsDefault:   v.IsDefault,
		IsShared:    v.IsShared,
		OwnerID:     v.OwnerID,
		ProjectID:   v.ProjectID,
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
	}
}
