package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// StateWorkflowService owns the state machine: `workflows` and the transitions
// that constrain issue status changes. It reads nothing from
// `agent_workflows`; orchestration lives in WorkflowService.
type StateWorkflowService struct {
	db *gorm.DB
}

// NewStateWorkflowService creates a new StateWorkflowService.
func NewStateWorkflowService(db *gorm.DB) *StateWorkflowService {
	return &StateWorkflowService{db: db}
}

// StateTransitionResponse represents a state transition in API response.
type StateTransitionResponse struct {
	ID                   uint64  `json:"id"`
	Name                 string  `json:"name"`
	Description          *string `json:"description"`
	WorkflowID           uint64  `json:"workflow_id"`
	SourceStateID        uint64  `json:"source_state_id"`
	TargetStateID        uint64  `json:"target_state_id"`
	IsAuto               bool    `json:"is_auto"`
	RuleType             string  `json:"rule_type"`
	ApproverIDs          *string `json:"approver_ids"`
	RoleAllowed          string  `json:"role_allowed"`
	ApproveTargetStateID *uint64 `json:"approve_target_state_id"`
	RejectTargetStateID  *uint64 `json:"reject_target_state_id"`
	ApprovalMode         string  `json:"approval_mode"`
	ProjectID            *uint64 `json:"project_id"`
	WorkspaceID          uint64  `json:"workspace_id"`
}

// AddTransitionRequest is the payload for creating a state transition.
// Accepts both source/target and from/to field names for UI compatibility.
type AddTransitionRequest struct {
	Name                 string  `json:"name"`
	SourceStateID        uint64  `json:"source_state_id"`
	TargetStateID        uint64  `json:"target_state_id"`
	FromStateID          uint64  `json:"from_state_id"`
	ToStateID            uint64  `json:"to_state_id"`
	Description          string  `json:"description"`
	RuleType             string  `json:"rule_type"` // allow | approval
	ApproverIDs          *string `json:"approver_ids"`
	RoleAllowed          string  `json:"role_allowed"`
	ApproveTargetStateID *uint64 `json:"approve_target_state_id"`
	RejectTargetStateID  *uint64 `json:"reject_target_state_id"`
	ApprovalMode         string  `json:"approval_mode"` // any | all
}

func (r *AddTransitionRequest) normalize() error {
	if r.SourceStateID == 0 {
		r.SourceStateID = r.FromStateID
	}
	if r.TargetStateID == 0 {
		r.TargetStateID = r.ToStateID
	}
	if r.SourceStateID == 0 || r.TargetStateID == 0 {
		return errors.New("from_state_id/source_state_id and to_state_id/target_state_id are required")
	}
	if r.Name == "" {
		r.Name = fmt.Sprintf("%d→%d", r.SourceStateID, r.TargetStateID)
	}
	return nil
}

// CreateStateWorkflowRequest creates a state-machine workflow (not agent orchestration).
type CreateStateWorkflowRequest struct {
	Name         string   `json:"name" binding:"required"`
	Description  string   `json:"description"`
	IssueTypeID  *uint64  `json:"issue_type_id"`
	IssueTypeIDs []uint64 `json:"issue_type_ids"`
}

// UpdateStateWorkflowRequest updates a state-machine workflow.
type UpdateStateWorkflowRequest struct {
	Name         *string   `json:"name"`
	Description  *string   `json:"description"`
	IssueTypeID  *uint64   `json:"issue_type_id"`
	IssueTypeIDs *[]uint64 `json:"issue_type_ids"`
	IsActive     *bool     `json:"is_active"`
}

// StateWorkflowResponse is the UI-facing shape for state-machine workflows.
type StateWorkflowResponse struct {
	ID           uint64                  `json:"id"`
	Name         string                  `json:"name"`
	Description  string                  `json:"description"`
	ProjectID    *uint64                 `json:"project_id"`
	WorkspaceID  uint64                  `json:"workspace_id"`
	IssueTypeID  *uint64                 `json:"issue_type_id"`
	IssueTypeIDs []uint64                `json:"issue_type_ids"`
	IsActive     bool                    `json:"is_active"`
	Transitions  []StateTransitionUIItem `json:"transitions"`
	CreatedAt    string                  `json:"created_at"`
	UpdatedAt    string                  `json:"updated_at"`
}

// StateTransitionUIItem matches StateWorkflowManager expectations (from_/to_ names).
type StateTransitionUIItem struct {
	ID                   uint64  `json:"id"`
	WorkflowID           uint64  `json:"workflow_id"`
	FromStateID          uint64  `json:"from_state_id"`
	ToStateID            uint64  `json:"to_state_id"`
	SourceStateID        uint64  `json:"source_state_id"`
	TargetStateID        uint64  `json:"target_state_id"`
	Description          string  `json:"description"`
	RuleType             string  `json:"rule_type"`
	ApproverIDs          *string `json:"approver_ids"`
	RoleAllowed          string  `json:"role_allowed"`
	ApproveTargetStateID *uint64 `json:"approve_target_state_id"`
	RejectTargetStateID  *uint64 `json:"reject_target_state_id"`
	ApprovalMode         string  `json:"approval_mode"`
	FromName             string  `json:"from_name,omitempty"`
	ToName               string  `json:"to_name,omitempty"`
	SourceName           string  `json:"source_name,omitempty"`
	TargetName           string  `json:"target_name,omitempty"`
}

// UpdateTransitionRequest is the payload for updating a state transition.
type UpdateTransitionRequest struct {
	Name                 *string `json:"name"`
	SourceStateID        *uint64 `json:"source_state_id"`
	TargetStateID        *uint64 `json:"target_state_id"`
	RuleType             *string `json:"rule_type"`
	ApproverIDs          *string `json:"approver_ids"`
	RoleAllowed          *string `json:"role_allowed"`
	ApproveTargetStateID *uint64 `json:"approve_target_state_id"`
	RejectTargetStateID  *uint64 `json:"reject_target_state_id"`
	ApprovalMode         *string `json:"approval_mode"`
}

// ListTransitions returns all state transitions attached to a state-machine
// workflow. Transition ids and state ids live in different spaces, and
// workflow_id has a single parent (`workflows`), so no kind filter is needed.
func (s *StateWorkflowService) ListTransitions(workflowID uint64) ([]StateTransitionResponse, error) {
	var transitions []model.StateTransition
	if err := s.db.Where("workflow_id = ? AND deleted_at IS NULL", workflowID).
		Order("id ASC").Find(&transitions).Error; err != nil {
		return nil, err
	}

	result := make([]StateTransitionResponse, len(transitions))
	for i, t := range transitions {
		result[i] = s.toTransitionResponse(&t)
	}
	return result, nil
}

// AddTransition creates a new state transition for a state-machine workflow.
func (s *StateWorkflowService) AddTransition(workflowID uint64, req AddTransitionRequest) (*StateTransitionResponse, error) {
	if err := req.normalize(); err != nil {
		return nil, err
	}

	workspaceID, projectID, err := s.resolveWorkflowScope(workflowID)
	if err != nil {
		return nil, err
	}

	ruleType := req.RuleType
	if ruleType == "" {
		ruleType = "allow"
	}
	approvalMode := req.ApprovalMode
	if approvalMode == "" {
		approvalMode = "any"
	}

	var desc *string
	if req.Description != "" {
		d := req.Description
		desc = &d
	}

	transition := model.StateTransition{
		Name:                 req.Name,
		Description:          desc,
		WorkflowID:           workflowID,
		SourceStateID:        req.SourceStateID,
		TargetStateID:        req.TargetStateID,
		RuleType:             ruleType,
		ApproverIDs:          req.ApproverIDs,
		RoleAllowed:          req.RoleAllowed,
		ApproveTargetStateID: req.ApproveTargetStateID,
		RejectTargetStateID:  req.RejectTargetStateID,
		ApprovalMode:         approvalMode,
		ProjectID:            projectID,
		WorkspaceID:          workspaceID,
	}

	if err := s.db.Create(&transition).Error; err != nil {
		return nil, err
	}

	resp := s.toTransitionResponse(&transition)
	return &resp, nil
}

// resolveWorkflowScope finds workspace/project for a state-machine workflow.
func (s *StateWorkflowService) resolveWorkflowScope(workflowID uint64) (uint64, *uint64, error) {
	var sw model.Workflow
	if err := s.db.First(&sw, workflowID).Error; err != nil {
		return 0, nil, errors.New("workflow not found")
	}
	return sw.WorkspaceID, sw.ProjectID, nil
}

// ResolveWorkspaceID resolves a numeric id or slug to a workspace id.
func (s *StateWorkflowService) ResolveWorkspaceID(wsParam string) (uint64, error) {
	if id, err := strconv.ParseUint(wsParam, 10, 64); err == nil {
		return id, nil
	}
	var ws model.Workspace
	if err := s.db.Where("slug = ?", wsParam).First(&ws).Error; err != nil {
		return 0, err
	}
	return ws.ID, nil
}

func parseIssueTypeIDs(raw json.RawMessage) []uint64 {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var ids []uint64
	if err := json.Unmarshal(raw, &ids); err != nil {
		return nil
	}
	return ids
}

func encodeIssueTypeIDs(ids []uint64) json.RawMessage {
	if len(ids) == 0 {
		return nil
	}
	b, err := json.Marshal(ids)
	if err != nil {
		return nil
	}
	return b
}

func (s *StateWorkflowService) toStateWorkflowResponse(wf *model.Workflow) StateWorkflowResponse {
	resp := StateWorkflowResponse{
		ID:           wf.ID,
		Name:         wf.Name,
		Description:  wf.Description,
		ProjectID:    wf.ProjectID,
		WorkspaceID:  wf.WorkspaceID,
		IssueTypeID:  wf.IssueTypeID,
		IssueTypeIDs: parseIssueTypeIDs(wf.IssueTypeIDs),
		IsActive:     wf.IsActive,
		Transitions:  []StateTransitionUIItem{},
		CreatedAt:    wf.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    wf.UpdatedAt.Format(time.RFC3339),
	}

	var transitions []model.StateTransition
	s.db.Where("workflow_id = ? AND deleted_at IS NULL", wf.ID).
		Order("id ASC").Find(&transitions)
	for _, t := range transitions {
		item := StateTransitionUIItem{
			ID:                   t.ID,
			WorkflowID:           t.WorkflowID,
			FromStateID:          t.SourceStateID,
			ToStateID:            t.TargetStateID,
			SourceStateID:        t.SourceStateID,
			TargetStateID:        t.TargetStateID,
			RuleType:             t.RuleType,
			ApproverIDs:          t.ApproverIDs,
			RoleAllowed:          t.RoleAllowed,
			ApproveTargetStateID: t.ApproveTargetStateID,
			RejectTargetStateID:  t.RejectTargetStateID,
			ApprovalMode:         t.ApprovalMode,
		}
		if t.Description != nil {
			item.Description = *t.Description
		}
		var fromName, toName string
		s.db.Raw("SELECT name FROM states WHERE id = ?", t.SourceStateID).Scan(&fromName)
		s.db.Raw("SELECT name FROM states WHERE id = ?", t.TargetStateID).Scan(&toName)
		item.FromName = fromName
		item.ToName = toName
		item.SourceName = fromName
		item.TargetName = toName
		resp.Transitions = append(resp.Transitions, item)
	}
	return resp
}

// ListStateWorkflowsByWorkspace returns workspace-level state-machine workflows.
func (s *StateWorkflowService) ListStateWorkflowsByWorkspace(workspaceID uint64) ([]StateWorkflowResponse, error) {
	var workflows []model.Workflow
	err := s.db.Where("workspace_id = ? AND project_id IS NULL AND deleted_at IS NULL", workspaceID).
		Order("created_at DESC").Find(&workflows).Error
	if err != nil {
		return nil, err
	}
	result := make([]StateWorkflowResponse, 0, len(workflows))
	for i := range workflows {
		result = append(result, s.toStateWorkflowResponse(&workflows[i]))
	}
	return result, nil
}

// GetStateWorkflow returns one state-machine workflow in a workspace. It reads
// the `workflows` table, never `agent_workflows`: the two have overlapping ids
// and the workspace routes address the former.
func (s *StateWorkflowService) GetStateWorkflow(workspaceID, workflowID uint64) (*StateWorkflowResponse, error) {
	var wf model.Workflow
	if err := s.db.Where("id = ? AND workspace_id = ? AND deleted_at IS NULL", workflowID, workspaceID).First(&wf).Error; err != nil {
		return nil, errors.New("workflow not found")
	}
	resp := s.toStateWorkflowResponse(&wf)
	return &resp, nil
}

// CreateStateWorkflow creates a workspace-level (or project-scoped) state-machine workflow.
func (s *StateWorkflowService) CreateStateWorkflow(workspaceID uint64, projectID *uint64, req CreateStateWorkflowRequest) (*StateWorkflowResponse, error) {
	wf := &model.Workflow{
		Name:         req.Name,
		Description:  req.Description,
		WorkspaceID:  workspaceID,
		ProjectID:    projectID,
		IssueTypeID:  req.IssueTypeID,
		IssueTypeIDs: encodeIssueTypeIDs(req.IssueTypeIDs),
		IsActive:     true,
	}
	if len(req.IssueTypeIDs) == 1 && req.IssueTypeID == nil {
		id := req.IssueTypeIDs[0]
		wf.IssueTypeID = &id
	}
	if err := s.db.Create(wf).Error; err != nil {
		return nil, err
	}
	resp := s.toStateWorkflowResponse(wf)
	return &resp, nil
}

// UpdateStateWorkflow updates a state-machine workflow owned by the workspace.
func (s *StateWorkflowService) UpdateStateWorkflow(workspaceID, workflowID uint64, req UpdateStateWorkflowRequest) (*StateWorkflowResponse, error) {
	var wf model.Workflow
	if err := s.db.Where("id = ? AND workspace_id = ?", workflowID, workspaceID).First(&wf).Error; err != nil {
		return nil, errors.New("workflow not found")
	}
	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.IssueTypeID != nil {
		updates["issue_type_id"] = *req.IssueTypeID
	}
	if req.IssueTypeIDs != nil {
		updates["issue_type_ids"] = encodeIssueTypeIDs(*req.IssueTypeIDs)
	}
	if len(updates) > 0 {
		if err := s.db.Model(&wf).Updates(updates).Error; err != nil {
			return nil, err
		}
		if err := s.db.First(&wf, workflowID).Error; err != nil {
			return nil, err
		}
	}
	resp := s.toStateWorkflowResponse(&wf)
	return &resp, nil
}

// DeleteStateWorkflow soft-deletes a state-machine workflow and its transitions.
func (s *StateWorkflowService) DeleteStateWorkflow(workspaceID, workflowID uint64) error {
	var wf model.Workflow
	if err := s.db.Where("id = ? AND workspace_id = ?", workflowID, workspaceID).First(&wf).Error; err != nil {
		return errors.New("workflow not found")
	}
	if err := s.db.Where("workflow_id = ?", workflowID).
		Delete(&model.StateTransition{}).Error; err != nil {
		return err
	}
	result := s.db.Delete(&wf)
	if result.RowsAffected == 0 {
		return errors.New("workflow not found")
	}
	return result.Error
}

// UpdateTransition updates a state transition.
func (s *StateWorkflowService) UpdateTransition(transitionID uint64, req UpdateTransitionRequest) error {
	updates := map[string]interface{}{}

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.SourceStateID != nil {
		updates["source_state_id"] = *req.SourceStateID
	}
	if req.TargetStateID != nil {
		updates["target_state_id"] = *req.TargetStateID
	}
	if req.RuleType != nil {
		updates["rule_type"] = *req.RuleType
	}
	if req.ApproverIDs != nil {
		updates["approver_ids"] = *req.ApproverIDs
	}
	if req.RoleAllowed != nil {
		updates["role_allowed"] = *req.RoleAllowed
	}
	if req.ApproveTargetStateID != nil {
		updates["approve_target_state_id"] = *req.ApproveTargetStateID
	}
	if req.RejectTargetStateID != nil {
		updates["reject_target_state_id"] = *req.RejectTargetStateID
	}
	if req.ApprovalMode != nil {
		updates["approval_mode"] = *req.ApprovalMode
	}

	if len(updates) == 0 {
		return nil
	}

	result := s.db.Model(&model.StateTransition{}).
		Where("id = ?", transitionID).Updates(updates)
	if result.RowsAffected == 0 {
		return errors.New("transition not found")
	}
	return result.Error
}

// DeleteTransition deletes a state transition.
func (s *StateWorkflowService) DeleteTransition(transitionID uint64) error {
	result := s.db.Where("id = ?", transitionID).Delete(&model.StateTransition{})
	if result.RowsAffected == 0 {
		return errors.New("transition not found")
	}
	return result.Error
}

func (s *StateWorkflowService) toTransitionResponse(t *model.StateTransition) StateTransitionResponse {
	return StateTransitionResponse{
		ID:                   t.ID,
		Name:                 t.Name,
		Description:          t.Description,
		WorkflowID:           t.WorkflowID,
		SourceStateID:        t.SourceStateID,
		TargetStateID:        t.TargetStateID,
		IsAuto:               t.IsAuto,
		RuleType:             t.RuleType,
		ApproverIDs:          t.ApproverIDs,
		RoleAllowed:          t.RoleAllowed,
		ApproveTargetStateID: t.ApproveTargetStateID,
		RejectTargetStateID:  t.RejectTargetStateID,
		ApprovalMode:         t.ApprovalMode,
		ProjectID:            t.ProjectID,
		WorkspaceID:          t.WorkspaceID,
	}
}
