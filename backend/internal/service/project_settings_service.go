package service

import (
	"errors"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
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
func (s *ProjectSettingsService) CreateState(req *request.StateCreateRequest, projectID uint64) (*response.StateResponse, error) {
	var project model.Project
	if err := s.db.Select("workspace_id").Where("id = ?", projectID).First(&project).Error; err != nil {
		return nil, common.Internal("Project not found")
	}
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

	projectIDPtr := projectID
	state := &model.State{
		Name:        req.Name,
		Color:       color,
		Group:       group,
		Sequence:    seq,
		IsDefault:   isDefault,
		IsActive:    true,
		ProjectID:   &projectIDPtr,
		WorkspaceID: project.WorkspaceID,
	}

	if err := s.db.Create(state).Error; err != nil {
		return nil, common.Internal("Failed to create state")
	}

	return stateToResponse(state), nil
}

// ListStates returns all states for a project including inherited workspace states.
func (s *ProjectSettingsService) ListStates(projectID uint64, includeInactive bool) ([]response.StateResponse, error) {
	var project model.Project
	if err := s.db.Select("workspace_id").Where("id = ?", projectID).First(&project).Error; err != nil {
		return nil, common.Internal("Project not found")
	}
	query := s.db.Where("(project_id = ? OR (project_id IS NULL AND workspace_id = ?))", projectID, project.WorkspaceID)
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
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	var defaultState model.State
	if err := tx.Where("project_id = ? AND is_default = ?", projectID, true).First(&defaultState).Error; err != nil {
		tx.Rollback()
		return common.NewErrorDetail(common.ErrInternal, "Failed to find default state", err.Error())
	}
	if defaultState.ID == stateID {
		tx.Rollback()
		return common.BadRequest("Cannot delete the default state")
	}
	if err := tx.Model(&model.Issue{}).Where("project_id = ? AND state_id = ?", projectID, stateID).Update("state_id", defaultState.ID).Error; err != nil {
		tx.Rollback()
		return common.NewErrorDetail(common.ErrInternal, "Failed to update issue states", err.Error())
	}
	result := tx.Where("id = ? AND project_id = ?", stateID, projectID).Delete(&model.State{})
	if result.RowsAffected == 0 {
		tx.Rollback()
		return common.NotFound("State not found")
	}
	return tx.Commit().Error
}

// ListWorkspaceStates returns workspace-level states (project_id IS NULL).
func (s *ProjectSettingsService) ListWorkspaceStates(workspaceID uint64) ([]response.StateResponse, error) {
	var states []model.State
	if err := s.db.Where("workspace_id = ? AND project_id IS NULL AND is_active = ?", workspaceID, true).Order("sequence ASC").Find(&states).Error; err != nil {
		return nil, common.Internal("Database error")
	}

	result := make([]response.StateResponse, len(states))
	for i, st := range states {
		result[i] = *stateToResponse(&st)
	}
	return result, nil
}

// CreateWorkspaceState creates a new workspace-level state.
func (s *ProjectSettingsService) CreateWorkspaceState(req *request.StateCreateRequest, workspaceID uint64) (*response.StateResponse, error) {
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
		ProjectID:   nil,
		WorkspaceID: workspaceID,
	}

	if err := s.db.Create(state).Error; err != nil {
		return nil, common.Internal("Failed to create workspace state")
	}

	return stateToResponse(state), nil
}

// GetWorkspaceState returns a single workspace-level state by ID.
func (s *ProjectSettingsService) GetWorkspaceState(workspaceID, stateID uint64) (*response.StateResponse, error) {
	var state model.State
	if err := s.db.Where("id = ? AND workspace_id = ? AND project_id IS NULL", stateID, workspaceID).First(&state).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("Workspace state not found")
		}
		return nil, common.Internal("Database error")
	}
	return stateToResponse(&state), nil
}

// UpdateWorkspaceState updates a workspace-level state's properties.
func (s *ProjectSettingsService) UpdateWorkspaceState(workspaceID, stateID uint64, req *request.StateUpdateRequest) (*response.StateResponse, error) {
	var state model.State
	if err := s.db.Where("id = ? AND workspace_id = ? AND project_id IS NULL", stateID, workspaceID).First(&state).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("Workspace state not found")
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
			return nil, common.Internal("Failed to update workspace state")
		}
	}

	return stateToResponse(&state), nil
}

// DeleteWorkspaceState soft-deletes a workspace-level state.
func (s *ProjectSettingsService) DeleteWorkspaceState(workspaceID, stateID uint64) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var state model.State
	if err := tx.Where("id = ? AND workspace_id = ? AND project_id IS NULL", stateID, workspaceID).First(&state).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.NotFound("Workspace state not found")
		}
		return common.Internal("Database error")
	}

	if state.IsDefault {
		tx.Rollback()
		return common.BadRequest("Cannot delete the default state")
	}

	result := tx.Where("id = ? AND workspace_id = ? AND project_id IS NULL", stateID, workspaceID).Delete(&model.State{})
	if result.RowsAffected == 0 {
		tx.Rollback()
		return common.NotFound("Workspace state not found")
	}
	return tx.Commit().Error
}

// CreateDefaultStates creates the 6 default states for a project.
func (s *ProjectSettingsService) CreateDefaultStates(projectID uint64) ([]response.StateResponse, error) {
	var project model.Project
	if err := s.db.Select("workspace_id").Where("id = ?", projectID).First(&project).Error; err != nil {
		return nil, common.Internal("Project not found")
	}
	var states []model.State
	for _, ds := range common.DefaultStates {
		projectIDPtr := projectID
		state := model.State{
			Name:        ds.Name,
			Color:       ds.Color,
			Group:       ds.Group,
			Sequence:    ds.Sequence,
			IsDefault:   ds.IsDefault,
			IsActive:    true,
			ProjectID:   &projectIDPtr,
			WorkspaceID: project.WorkspaceID,
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
	var project model.Project
	if err := s.db.Select("workspace_id").Where("id = ?", projectID).First(&project).Error; err != nil {
		return nil, common.Internal("Project not found")
	}
	color := req.Color
	if color == "" {
		color = "#6B7280"
	}

	label := &model.Label{
		Name:        req.Name,
		Color:       color,
		Description: req.Description,
		ProjectID:   projectID,
		WorkspaceID: project.WorkspaceID,
	}

	if err := s.db.Create(label).Error; err != nil {
		if common.IsUniqueViolation(err) {
			return nil, common.Conflict("Label name already exists in this project")
		}
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

// SearchLabels returns labels matching the query.
func (s *ProjectSettingsService) SearchLabels(projectID uint64, query string) ([]response.LabelResponse, error) {
	var labels []model.Label
	if err := s.db.Where("project_id = ? AND name ILIKE ?", projectID, "%"+query+"%").Order("created_at ASC").Find(&labels).Error; err != nil {
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
			if common.IsUniqueViolation(err) {
				return nil, common.Conflict("Label name already exists in this project")
			}
			return nil, common.Internal("Failed to update label")
		}
	}

	return labelToResponse(&label), nil
}

// DeleteLabel soft-deletes a label.
func (s *ProjectSettingsService) DeleteLabel(projectID, labelID uint64) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if err := tx.Where("label_id = ?", labelID).Delete(&model.IssueLabel{}).Error; err != nil {
		tx.Rollback()
		return common.Internal("Failed to delete issue labels")
	}
	result := tx.Where("id = ? AND project_id = ?", labelID, projectID).Delete(&model.Label{})
	if result.RowsAffected == 0 {
		tx.Rollback()
		return common.NotFound("Label not found")
	}
	return tx.Commit().Error
}

// stateToResponse converts a State model to a StateResponse.
func stateToResponse(state *model.State) *response.StateResponse {
	id := state.ID
	resp := &response.StateResponse{
		ID:          &id,
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
		IsInherited: state.ProjectID == nil,
	}

	if state.DeletedAt.Valid {
		resp.DeletedAt = &state.DeletedAt.Time
		resp.IsDeleted = true
	}

	return resp
}

// labelToResponse converts a Label model to a LabelResponse.
func labelToResponse(label *model.Label) *response.LabelResponse {
	id := label.ID
	resp := &response.LabelResponse{
		ID:          &id,
		Name:        label.Name,
		Color:       label.Color,
		Description: label.Description,
		ProjectID:   label.ProjectID,
		WorkspaceID: label.WorkspaceID,
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
