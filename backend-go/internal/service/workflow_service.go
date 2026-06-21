package service

import (
	"github.com/reqmanpy/backend-go/internal/common"
	"github.com/reqmanpy/backend-go/internal/dto/request"
	"github.com/reqmanpy/backend-go/internal/dto/response"
	"github.com/reqmanpy/backend-go/internal/model"
	"gorm.io/gorm"
)

type WorkflowService struct{ db *gorm.DB }
func NewWorkflowService(db *gorm.DB) *WorkflowService { return &WorkflowService{db: db} }
func descStr(d *string) string { if d != nil { return *d }; return "" }

func (s *WorkflowService) Create(pid uint64, req request.WorkflowCreate) (*response.WorkflowResponse, error) {
	w := model.Workflow{Name: req.Name, Description: req.Description, ProjectID: pid, IssueTypeID: req.IssueTypeID}
	if err := s.db.Create(&w).Error; err != nil { return nil, common.Internal("Failed to create workflow") }
	return s.Get(w.ID)
}
func (s *WorkflowService) List(pid uint64) ([]response.WorkflowResponse, error) {
	var ws []model.Workflow
	s.db.Preload("Transitions.SourceState").Preload("Transitions.TargetState").Where("project_id = ?", pid).Find(&ws)
	res := make([]response.WorkflowResponse, len(ws))
	for i, w := range ws {
		res[i] = response.WorkflowResponse{ID: w.ID, Name: w.Name, Description: w.Description, ProjectID: w.ProjectID, IssueTypeID: w.IssueTypeID, IsActive: w.IsActive, CreatedAt: w.CreatedAt, UpdatedAt: w.UpdatedAt, Transitions: make([]response.TransitionResponse, 0)}
		for _, t := range w.Transitions {
			res[i].Transitions = append(res[i].Transitions, response.TransitionResponse{ID: t.ID, WorkflowID: t.WorkflowID, SourceStateID: t.SourceStateID, TargetStateID: t.TargetStateID, Description: descStr(t.Description), RuleType: t.RuleType, ApproverIDs: t.ApproverIDs, RoleAllowed: t.RoleAllowed, SourceName: t.SourceState.Name, TargetName: t.TargetState.Name})
		}
	}
	if res == nil { res = []response.WorkflowResponse{} }; return res, nil
}
func (s *WorkflowService) Get(id uint64) (*response.WorkflowResponse, error) {
	var w model.Workflow
	if err := s.db.Preload("Transitions.SourceState").Preload("Transitions.TargetState").First(&w, id).Error; err != nil { return nil, common.NotFound("Workflow not found") }
	r := &response.WorkflowResponse{ID: w.ID, Name: w.Name, Description: w.Description, ProjectID: w.ProjectID, IssueTypeID: w.IssueTypeID, IsActive: w.IsActive, CreatedAt: w.CreatedAt, UpdatedAt: w.UpdatedAt, Transitions: make([]response.TransitionResponse, 0)}
	for _, t := range w.Transitions {
		r.Transitions = append(r.Transitions, response.TransitionResponse{ID: t.ID, WorkflowID: t.WorkflowID, SourceStateID: t.SourceStateID, TargetStateID: t.TargetStateID, Description: descStr(t.Description), RuleType: t.RuleType, ApproverIDs: t.ApproverIDs, RoleAllowed: t.RoleAllowed, SourceName: t.SourceState.Name, TargetName: t.TargetState.Name})
	}
	return r, nil
}
func (s *WorkflowService) Update(id uint64, req request.WorkflowUpdate) (*response.WorkflowResponse, error) {
	var w model.Workflow
	if err := s.db.First(&w, id).Error; err != nil { return nil, common.NotFound("Workflow not found") }
	if req.Name != nil { w.Name = *req.Name }
	if req.Description != nil { w.Description = *req.Description }
	if req.IssueTypeID != nil { w.IssueTypeID = req.IssueTypeID }
	if req.IsActive != nil { w.IsActive = *req.IsActive }
	s.db.Save(&w); return s.Get(id)
}
func (s *WorkflowService) Delete(id uint64) error {
	s.db.Where("workflow_id = ?", id).Delete(&model.StateTransition{})
	return s.db.Delete(&model.Workflow{}, id).Error
}
func (s *WorkflowService) AddTransition(wid uint64, req request.TransitionCreate) (*response.TransitionResponse, error) {
	var w model.Workflow
	if err := s.db.First(&w, wid).Error; err != nil { return nil, common.NotFound("Workflow not found") }
	t := model.StateTransition{WorkflowID: wid, SourceStateID: req.FromStateID, TargetStateID: req.ToStateID, Description: &req.Description, RuleType: req.RuleType, ApproverIDs: req.ApproverIDs, RoleAllowed: req.RoleAllowed}
	if t.RuleType == "" { t.RuleType = "allow" }
	if err := s.db.Create(&t).Error; err != nil { return nil, common.Internal("Failed to add transition") }
	var fs, ts model.State
	s.db.First(&fs, req.FromStateID); s.db.First(&ts, req.ToStateID)
	return &response.TransitionResponse{ID: t.ID, WorkflowID: wid, SourceStateID: t.SourceStateID, TargetStateID: t.TargetStateID, Description: descStr(t.Description), RuleType: t.RuleType, ApproverIDs: t.ApproverIDs, RoleAllowed: t.RoleAllowed, SourceName: fs.Name, TargetName: ts.Name}, nil
}
func (s *WorkflowService) UpdateTransition(id uint64, req request.TransitionUpdate) (*response.TransitionResponse, error) {
	var t model.StateTransition
	if err := s.db.First(&t, id).Error; err != nil { return nil, common.NotFound("Transition not found") }
	if req.Description != nil { t.Description = req.Description }
	if req.RuleType != nil { t.RuleType = *req.RuleType }
	if req.ApproverIDs != nil { t.ApproverIDs = req.ApproverIDs }
	if req.RoleAllowed != nil { t.RoleAllowed = *req.RoleAllowed }
	s.db.Save(&t)
	return &response.TransitionResponse{ID: t.ID, WorkflowID: t.WorkflowID, SourceStateID: t.SourceStateID, TargetStateID: t.TargetStateID, Description: descStr(t.Description), RuleType: t.RuleType, ApproverIDs: t.ApproverIDs, RoleAllowed: t.RoleAllowed}, nil
}
func (s *WorkflowService) DeleteTransition(id uint64) error { return s.db.Delete(&model.StateTransition{}, id).Error }

func (s *WorkflowService) CreateAutomation(pid uint64, req request.AutomationCreate) (*response.AutomationResponse, error) {
	a := model.AutomationRule{Name: req.Name, Description: req.Description, ProjectID: pid, TriggerType: req.TriggerType, Conditions: req.Conditions, Actions: req.Actions, Sequence: req.Sequence}
	if err := s.db.Create(&a).Error; err != nil { return nil, common.Internal("Failed to create automation") }
	return &response.AutomationResponse{ID: a.ID, Name: a.Name, Description: a.Description, ProjectID: a.ProjectID, IsEnabled: a.IsEnabled, Sequence: a.Sequence, TriggerType: a.TriggerType, Conditions: a.Conditions, Actions: a.Actions, CreatedAt: a.CreatedAt, UpdatedAt: a.UpdatedAt}, nil
}
func (s *WorkflowService) ListAutomations(pid uint64) ([]response.AutomationResponse, error) {
	var as []model.AutomationRule
	s.db.Where("project_id = ?", pid).Order("sequence").Find(&as)
	res := make([]response.AutomationResponse, len(as))
	for i, a := range as {
		res[i] = response.AutomationResponse{ID: a.ID, Name: a.Name, Description: a.Description, ProjectID: a.ProjectID, IsEnabled: a.IsEnabled, Sequence: a.Sequence, TriggerType: a.TriggerType, Conditions: a.Conditions, Actions: a.Actions, CreatedAt: a.CreatedAt, UpdatedAt: a.UpdatedAt}
	}
	if res == nil { res = []response.AutomationResponse{} }; return res, nil
}
func (s *WorkflowService) UpdateAutomation(id uint64, req request.AutomationUpdate) (*response.AutomationResponse, error) {
	var a model.AutomationRule
	if err := s.db.First(&a, id).Error; err != nil { return nil, common.NotFound("Automation not found") }
	if req.Name != nil { a.Name = *req.Name }
	if req.Description != nil { a.Description = *req.Description }
	if req.TriggerType != nil { a.TriggerType = *req.TriggerType }
	if req.Conditions != nil { a.Conditions = *req.Conditions }
	if req.Actions != nil { a.Actions = *req.Actions }
	if req.IsEnabled != nil { a.IsEnabled = *req.IsEnabled }
	if req.Sequence != nil { a.Sequence = *req.Sequence }
	s.db.Save(&a)
	return &response.AutomationResponse{ID: a.ID, Name: a.Name, Description: a.Description, ProjectID: a.ProjectID, IsEnabled: a.IsEnabled, Sequence: a.Sequence, TriggerType: a.TriggerType, Conditions: a.Conditions, Actions: a.Actions, CreatedAt: a.CreatedAt, UpdatedAt: a.UpdatedAt}, nil
}
func (s *WorkflowService) DeleteAutomation(id uint64) error { return s.db.Delete(&model.AutomationRule{}, id).Error }
