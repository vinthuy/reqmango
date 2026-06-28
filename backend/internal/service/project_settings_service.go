package service

import (
	"errors"

	"github.com/reqmanpy/backend/internal/common"
	"github.com/reqmanpy/backend/internal/dto/request"
	"github.com/reqmanpy/backend/internal/dto/response"
	"github.com/reqmanpy/backend/internal/model"
	"gorm.io/gorm"
)

type ProjectSettingsService struct {
	db *gorm.DB
}

func NewProjectSettingsService(db *gorm.DB) *ProjectSettingsService {
	return &ProjectSettingsService{db: db}
}

// ==================== State CRUD ====================

// CreateState creates a new state for a project.
func (s *ProjectSettingsService) CreateState(req *request.StateCreateRequest, projectID, workspaceID uint64) (*response.StateResponse, error) {
	color := req.Color
	if color == "" {
		color = "#6B7280"
	}
	group := req.Group
	if group == "" {
		group = common.StateGroupBacklog
	}
	seq := 1
	if req.Sequence != nil {
		seq = *req.Sequence
	}
	isDefault := false
	if req.IsDefault != nil {
		isDefault = *req.IsDefault
	}

	state := &model.State{
		Name:        req.Name,
		Color:       color,
		Group:       group,
		Sequence:    seq,
		IsDefault:   isDefault,
		IsActive:    true,
		ProjectID:   projectID,
		WorkspaceID: workspaceID,
	}

	if err := s.db.Create(state).Error; err != nil {
		return nil, common.Internal("Failed to create state")
	}

	return stateToResponse(state), nil
}

// ListStates returns all states for a project.
func (s *ProjectSettingsService) ListStates(projectID uint64, includeInactive bool) ([]response.StateResponse, error) {
	query := s.db.Where("project_id = ?", projectID)
	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}

	var states []model.State
	if err := query.Order("sequence ASC").Find(&states).Error; err != nil {
		return nil, common.Internal("Database error")
	}

	result := make([]response.StateResponse, len(states))
	for i, st := range states {
		result[i] = *stateToResponse(&st)
	}
	return result, nil
}

// GetState returns a single state by ID.
func (s *ProjectSettingsService) GetState(projectID, stateID uint64) (*response.StateResponse, error) {
	var state model.State
	if err := s.db.Where("id = ? AND project_id = ?", stateID, projectID).First(&state).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("State not found")
		}
		return nil, common.Internal("Database error")
	}
	return stateToResponse(&state), nil
}

// UpdateState updates a state's properties.
func (s *ProjectSettingsService) UpdateState(projectID, stateID uint64, req *request.StateUpdateRequest) (*response.StateResponse, error) {
	var state model.State
	if err := s.db.Where("id = ? AND project_id = ?", stateID, projectID).First(&state).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("State not found")
		}
		return nil, common.Internal("Database error")
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Color != nil {
		updates["color"] = *req.Color
	}
	if req.Group != nil {
		updates["group"] = *req.Group
	}
	if req.Sequence != nil {
		updates["sequence"] = *req.Sequence
	}
	if req.IsDefault != nil {
		updates["is_default"] = *req.IsDefault
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if len(updates) > 0 {
		if err := s.db.Model(&state).Updates(updates).Error; err != nil {
			return nil, common.Internal("Failed to update state")
		}
	}

	return stateToResponse(&state), nil
}

// DeleteState soft-deletes a state.
func (s *ProjectSettingsService) DeleteState(projectID, stateID uint64) error {
	result := s.db.Where("id = ? AND project_id = ?", stateID, projectID).Delete(&model.State{})
	if result.RowsAffected == 0 {
		return common.NotFound("State not found")
	}
	return nil
}

// CreateDefaultStates creates the 6 default states for a project.
func (s *ProjectSettingsService) CreateDefaultStates(projectID, workspaceID uint64) ([]response.StateResponse, error) {
	var states []model.State
	for _, ds := range common.DefaultStates {
		state := model.State{
			Name:        ds.Name,
			Color:       ds.Color,
			Group:       ds.Group,
			Sequence:    ds.Sequence,
			IsDefault:   ds.IsDefault,
			IsActive:    true,
			ProjectID:   projectID,
			WorkspaceID: workspaceID,
		}
		if err := s.db.Create(&state).Error; err != nil {
			return nil, common.Internal("Failed to create default states")
		}
		states = append(states, state)
	}

	result := make([]response.StateResponse, len(states))
	for i, st := range states {
		result[i] = *stateToResponse(&st)
	}
	return result, nil
}

// ==================== Label CRUD ====================

// CreateLabel creates a new label for a project.
func (s *ProjectSettingsService) CreateLabel(req *request.LabelCreateRequest, projectID uint64) (*response.LabelResponse, error) {
	color := req.Color
	if color == "" {
		color = "#6B7280"
	}

	label := &model.Label{
		Name:        req.Name,
		Color:       color,
		Description: req.Description,
		ProjectID:   projectID,
	}

	if err := s.db.Create(label).Error; err != nil {
		return nil, common.Internal("Failed to create label")
	}

	return labelToResponse(label), nil
}

// ListLabels returns all labels for a project.
func (s *ProjectSettingsService) ListLabels(projectID uint64) ([]response.LabelResponse, error) {
	var labels []model.Label
	if err := s.db.Where("project_id = ?", projectID).Order("created_at ASC").Find(&labels).Error; err != nil {
		return nil, common.Internal("Database error")
	}

	result := make([]response.LabelResponse, len(labels))
	for i, l := range labels {
		result[i] = *labelToResponse(&l)
	}
	return result, nil
}

// GetLabel returns a single label by ID.
func (s *ProjectSettingsService) GetLabel(projectID, labelID uint64) (*response.LabelResponse, error) {
	var label model.Label
	if err := s.db.Where("id = ? AND project_id = ?", labelID, projectID).First(&label).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("Label not found")
		}
		return nil, common.Internal("Database error")
	}
	return labelToResponse(&label), nil
}

// UpdateLabel updates a label's properties.
func (s *ProjectSettingsService) UpdateLabel(projectID, labelID uint64, req *request.LabelUpdateRequest) (*response.LabelResponse, error) {
	var label model.Label
	if err := s.db.Where("id = ? AND project_id = ?", labelID, projectID).First(&label).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("Label not found")
		}
		return nil, common.Internal("Database error")
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Color != nil {
		updates["color"] = *req.Color
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}

	if len(updates) > 0 {
		if err := s.db.Model(&label).Updates(updates).Error; err != nil {
			return nil, common.Internal("Failed to update label")
		}
	}

	return labelToResponse(&label), nil
}

// DeleteLabel soft-deletes a label.
func (s *ProjectSettingsService) DeleteLabel(projectID, labelID uint64) error {
	result := s.db.Where("id = ? AND project_id = ?", labelID, projectID).Delete(&model.Label{})
	if result.RowsAffected == 0 {
		return common.NotFound("Label not found")
	}
	return nil
}

// stateToResponse converts a State model to a StateResponse.
func stateToResponse(state *model.State) *response.StateResponse {
	resp := &response.StateResponse{
		ID:          state.ID,
		Name:        state.Name,
		Color:       state.Color,
		Group:       state.Group,
		Sequence:    state.Sequence,
		IsDefault:   state.IsDefault,
		IsActive:    state.IsActive,
		ProjectID:   state.ProjectID,
		WorkspaceID: state.WorkspaceID,
		CreatedAt:   state.CreatedAt,
		UpdatedAt:   state.UpdatedAt,
		CreatedByID: state.CreatedByID,
		UpdatedByID: state.UpdatedByID,
	}

	if state.DeletedAt.Valid {
		resp.DeletedAt = &state.DeletedAt.Time
		resp.IsDeleted = true
	}

	return resp
}

// labelToResponse converts a Label model to a LabelResponse.
func labelToResponse(label *model.Label) *response.LabelResponse {
	resp := &response.LabelResponse{
		ID:          label.ID,
		Name:        label.Name,
		Color:       label.Color,
		Description: label.Description,
		ProjectID:   label.ProjectID,
		CreatedAt:   label.CreatedAt,
		UpdatedAt:   label.UpdatedAt,
		CreatedByID: label.CreatedByID,
		UpdatedByID: label.UpdatedByID,
	}

	if label.DeletedAt.Valid {
		resp.DeletedAt = &label.DeletedAt.Time
		resp.IsDeleted = true
	}

	return resp
}
