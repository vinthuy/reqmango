package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// SDLCService implements PRD P4-006: 完整 SDLC 流程编排引擎.
//
// The service orchestrates the 11-stage software delivery lifecycle defined
// in PRD §3 (需求分析 → 需求设计 → 分派 Feature → 功能设计 → 拆解 US →
// 迭代排期 → US 开发 → 代码审查 → US 测试 → FE 测试 → 上线).
//
// Responsibilities:
//
//   - Workflow CRUD: create/get/list/cancel/delete an SDLC pipeline run.
//   - Stage tracking: each workflow owns 11 SDLCStage rows (one per canonical
//     stage) created up-front so the UI has a stable progress view.
//   - Async orchestration: runWorkflow walks the stages in order, invoking a
//     pluggable StageExecutor per stage, merging outputs into the workflow's
//     accumulated Artifacts, and broadcasting progress via SSE.
//   - Resume: RetryFromStage re-runs a failed stage and continues the pipeline.
//
// The StageExecutor abstraction decouples the orchestration engine from the
// concrete agent implementations so the workflow remains end-to-end testable
// without external dependencies (mirroring the CICDProvider / CodeGenerator
// patterns used by P4-001/004).
type SDLCService struct {
	db       *gorm.DB
	executor StageExecutor
}

// StageExecutor abstracts the execution of a single SDLC stage so the
// orchestration engine can be unit-tested without real agent dispatch.
//
// Implementations receive the workflow, the stage to run, and the accumulated
// artifacts from prior stages. They return the stage's output (merged into
// the workflow artifacts on success) and a slice of log lines.
type StageExecutor interface {
	Execute(ctx context.Context, wf *model.SDLCWorkflow, stage *model.SDLCStage, prior map[string]interface{}) (output map[string]interface{}, logs []string, err error)
}

// sdlcStageDef describes one canonical SDLC stage (PRD §3.2).
type sdlcStageDef struct {
	Order     int    `json:"order"`
	Key       string `json:"key"`
	Name      string `json:"name"`
	AgentRole string `json:"agent_role"`
}

// canonicalSDLStages is the ordered set of stages that make up a full SDLC
// pipeline (PRD §3.1/§3.3). Stages run strictly in order; the dependency
// chain is linear.
var canonicalSDLStages = []sdlcStageDef{
	{Order: 1, Key: "requirement_analysis", Name: "需求分析", AgentRole: "需求分析师"},
	{Order: 2, Key: "requirement_design", Name: "需求设计", AgentRole: "文档撰写者"},
	{Order: 3, Key: "dispatch_feature", Name: "分派 Feature", AgentRole: "Leader"},
	{Order: 4, Key: "feature_design", Name: "功能设计", AgentRole: "需求分析师"},
	{Order: 5, Key: "breakdown_us", Name: "拆解 US", AgentRole: "需求分析师"},
	{Order: 6, Key: "sprint_planning", Name: "迭代排期", AgentRole: "冲刺规划师"},
	{Order: 7, Key: "development", Name: "US 开发", AgentRole: "Developer"},
	{Order: 8, Key: "code_review", Name: "代码审查", AgentRole: "代码评审员"},
	{Order: 9, Key: "us_testing", Name: "US 测试", AgentRole: "Tester"},
	{Order: 10, Key: "fe_testing", Name: "FE 测试", AgentRole: "Tester"},
	{Order: 11, Key: "deploy", Name: "上线", AgentRole: "Leader"},
}

// NewSDLCService creates a new SDLCService. When executor is nil a stub
// executor is used so the workflow remains end-to-end testable without
// external dependencies.
func NewSDLCService(db *gorm.DB) *SDLCService {
	svc := &SDLCService{db: db}
	svc.executor = &stubSDLCStageExecutor{}
	return svc
}

// SetStageExecutor overrides the default stage executor. Used in tests.
func (s *SDLCService) SetStageExecutor(e StageExecutor) {
	if e != nil {
		s.executor = e
	}
}

// checkWorkspaceAdmin mirrors the guard used by other workspace services.
func (s *SDLCService) checkWorkspaceAdmin(workspaceID, callerID uint64) error {
	var member model.WorkspaceMember
	if err := s.db.Where("workspace_id = ? AND user_id = ? AND is_active = ?", workspaceID, callerID, true).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.Forbidden("You must be a workspace admin to manage SDLC workflows")
		}
		return common.Internal("Database error")
	}
	if member.Role < common.RoleAdmin {
		return common.Forbidden("You must be a workspace admin to manage SDLC workflows")
	}
	return nil
}

// ======== Request / Response types ========

// SDLCWorkflowCreate captures the inputs for a new SDLC pipeline run.
type SDLCWorkflowCreate struct {
	Title       string          `json:"title" binding:"required"`
	Requirement string          `json:"requirement"`
	ProjectID   *uint64         `json:"project_id"`
	SquadID     *uint64         `json:"squad_id"`
	// Stages optionally restricts which canonical stages run. Stage keys not
	// in this list are marked skipped. Empty = run all stages.
	Stages []string `json:"stages"`
	// FailFast: when true (default) the first stage failure fails the whole
	// workflow; when false the engine continues and marks the workflow
	// partial_failed at the end.
	FailFast *bool           `json:"fail_fast"`
	Config   json.RawMessage `json:"config"`
}

// SDLCWorkflowResponse is the API representation of an SDLCWorkflow.
type SDLCWorkflowResponse struct {
	ID           uint64          `json:"id"`
	WorkspaceID  uint64          `json:"workspace_id"`
	ProjectID    *uint64         `json:"project_id,omitempty"`
	SquadID      *uint64         `json:"squad_id,omitempty"`
	Title        string          `json:"title"`
	Requirement  string          `json:"requirement"`
	Status       string          `json:"status"`
	Progress     int             `json:"progress"`
	CurrentStage *string         `json:"current_stage,omitempty"`
	Config       json.RawMessage `json:"config"`
	Artifacts    json.RawMessage `json:"artifacts"`
	ErrorMessage *string         `json:"error_message,omitempty"`
	StartedAt    *time.Time      `json:"started_at,omitempty"`
	CompletedAt  *time.Time      `json:"completed_at,omitempty"`
	CancelledAt  *time.Time      `json:"cancelled_at,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	// Stages is populated when stages are preloaded (e.g. via Get).
	Stages []SDLCStageResponse `json:"stages,omitempty"`
}

// SDLCStageResponse is the API representation of an SDLCStage.
type SDLCStageResponse struct {
	ID           uint64          `json:"id"`
	WorkflowID   uint64          `json:"workflow_id"`
	WorkspaceID  uint64          `json:"workspace_id"`
	Order        int             `json:"order"`
	Key          string          `json:"key"`
	Name         string          `json:"name"`
	AgentRole    string          `json:"agent_role"`
	Status       string          `json:"status"`
	Progress     int             `json:"progress"`
	Input        json.RawMessage `json:"input"`
	Output       json.RawMessage `json:"output"`
	Logs         json.RawMessage `json:"logs"`
	ErrorMessage *string         `json:"error_message,omitempty"`
	StartedAt    *time.Time      `json:"started_at,omitempty"`
	CompletedAt  *time.Time      `json:"completed_at,omitempty"`
	DurationMs   int64           `json:"duration_ms"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// ======== CRUD ========

// Create starts a new SDLC pipeline run.
//
// The workflow and its 11 stage rows are persisted synchronously (status=
// pending) and the orchestration runs asynchronously so the API can return
// immediately. Callers poll the workflow status or subscribe to SSE events
// (sdlc_workflow.updated / sdlc_stage.updated) for progress.
func (s *SDLCService) Create(wid, callerID uint64, req SDLCWorkflowCreate) (*SDLCWorkflowResponse, error) {
	if err := s.checkWorkspaceAdmin(wid, callerID); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Title) == "" {
		return nil, common.BadRequest("Title is required")
	}

	// Resolve the runtime config: selected stages + fail_fast flag.
	selected := normalizeStageSelection(req.Stages)
	failFast := true
	if req.FailFast != nil {
		failFast = *req.FailFast
	}

	cfgMap := map[string]interface{}{
		"stages":    selected,
		"fail_fast": failFast,
	}
	if len(req.Config) > 0 && string(req.Config) != "null" {
		// Merge caller-supplied config on top (caller values win).
		var extra map[string]interface{}
		if err := json.Unmarshal(req.Config, &extra); err == nil {
			for k, v := range extra {
				cfgMap[k] = v
			}
		}
	}
	cfgJSON, _ := json.Marshal(cfgMap)

	wf := model.SDLCWorkflow{
		WorkspaceID: wid,
		ProjectID:   req.ProjectID,
		SquadID:     req.SquadID,
		Title:       req.Title,
		Requirement: req.Requirement,
		Status:      model.SDLCWorkflowPending,
		Config:      cfgJSON,
		Artifacts:   json.RawMessage("{}"),
	}
	if err := s.db.Create(&wf).Error; err != nil {
		return nil, common.Internal("Failed to create SDLC workflow")
	}

	// Create the 11 canonical stage rows up-front so the UI has a stable
	// progress grid regardless of which stages are selected to run.
	stageRows := make([]model.SDLCStage, 0, len(canonicalSDLStages))
	for _, def := range canonicalSDLStages {
		st := model.SDLCStagePending
		if len(selected) > 0 && !containsString(selected, def.Key) {
			st = model.SDLCStageSkipped
		}
		stageRows = append(stageRows, model.SDLCStage{
			WorkflowID:  wf.ID,
			WorkspaceID: wid,
			Order:       def.Order,
			Key:         def.Key,
			Name:        def.Name,
			AgentRole:   def.AgentRole,
			Status:      st,
			Input:       json.RawMessage("{}"),
			Output:      json.RawMessage("{}"),
			Logs:        json.RawMessage("[]"),
		})
	}
	if err := s.db.Create(&stageRows).Error; err != nil {
		return nil, common.Internal("Failed to create SDLC stages")
	}

	// Spawn the asynchronous orchestration.
	go s.runWorkflow(wf.ID, failFast, selected)

	resp := s.workflowToResponse(&wf)
	s.pushWorkflowEvent("sdlc_workflow.created", resp)
	return resp, nil
}

// Get returns a single SDLCWorkflow by ID (with its stages preloaded).
func (s *SDLCService) Get(id uint64) (*SDLCWorkflowResponse, error) {
	var wf model.SDLCWorkflow
	if err := s.db.Preload("Stages", func(db *gorm.DB) *gorm.DB {
		return db.Order("\"order\" ASC")
	}).First(&wf, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("SDLC workflow not found")
		}
		return nil, common.Internal("Failed to get SDLC workflow")
	}
	return s.workflowToResponse(&wf), nil
}

// List returns SDLCWorkflows for a workspace, newest first.
// status filter is optional.
func (s *SDLCService) List(wid uint64, status string, limit int) ([]SDLCWorkflowResponse, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	q := s.db.Where("workspace_id = ?", wid)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var wfs []model.SDLCWorkflow
	if err := q.Order("created_at DESC").Limit(limit).Find(&wfs).Error; err != nil {
		return nil, common.Internal("Failed to list SDLC workflows")
	}
	out := make([]SDLCWorkflowResponse, 0, len(wfs))
	for i := range wfs {
		out = append(out, *s.workflowToResponse(&wfs[i]))
	}
	return out, nil
}

// Cancel marks a pending/running workflow as cancelled. The async loop
// honors the transition between stages and stops further execution.
func (s *SDLCService) Cancel(id, callerID uint64) (*SDLCWorkflowResponse, error) {
	var wf model.SDLCWorkflow
	if err := s.db.First(&wf, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("SDLC workflow not found")
		}
		return nil, common.Internal("Failed to get SDLC workflow")
	}
	if err := s.checkWorkspaceAdmin(wf.WorkspaceID, callerID); err != nil {
		return nil, err
	}
	if isTerminalWorkflowStatus(wf.Status) {
		return nil, common.BadRequest("Cannot cancel a terminal workflow")
	}
	now := time.Now()
	wf.Status = model.SDLCWorkflowCancelled
	wf.CancelledAt = &now
	if wf.CompletedAt == nil {
		wf.CompletedAt = &now
	}
	if err := s.db.Save(&wf).Error; err != nil {
		return nil, common.Internal("Failed to cancel SDLC workflow")
	}
	resp := s.workflowToResponse(&wf)
	s.pushWorkflowEvent("sdlc_workflow.cancelled", resp)
	return resp, nil
}

// Delete removes an SDLCWorkflow and its stages (soft-delete via BaseModel).
func (s *SDLCService) Delete(id, callerID uint64) error {
	var wf model.SDLCWorkflow
	if err := s.db.First(&wf, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.NotFound("SDLC workflow not found")
		}
		return common.Internal("Failed to get SDLC workflow")
	}
	if err := s.checkWorkspaceAdmin(wf.WorkspaceID, callerID); err != nil {
		return err
	}
	if err := s.db.Where("workflow_id = ?", id).Delete(&model.SDLCStage{}).Error; err != nil {
		return common.Internal("Failed to delete SDLC stages")
	}
	return s.db.Delete(&wf).Error
}

// ======== Stage endpoints ========

// ListStages returns the stages of a workflow in canonical order.
func (s *SDLCService) ListStages(wfID uint64) ([]SDLCStageResponse, error) {
	var stages []model.SDLCStage
	if err := s.db.Where("workflow_id = ?", wfID).Order("\"order\" ASC").Find(&stages).Error; err != nil {
		return nil, common.Internal("Failed to list SDLC stages")
	}
	out := make([]SDLCStageResponse, 0, len(stages))
	for i := range stages {
		out = append(out, *s.stageToResponse(&stages[i]))
	}
	return out, nil
}

// GetStage returns a single SDLCStage by ID.
func (s *SDLCService) GetStage(wfID, stageID uint64) (*SDLCStageResponse, error) {
	var stage model.SDLCStage
	if err := s.db.Where("workflow_id = ? AND id = ?", wfID, stageID).First(&stage).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("SDLC stage not found")
		}
		return nil, common.Internal("Failed to get SDLC stage")
	}
	return s.stageToResponse(&stage), nil
}

// RetryFromStage resumes a failed/partial workflow by re-running a single
// stage and continuing the pipeline from that point onward (PRD §17.3
// 任务恢复执行). Prior successful stages keep their artifacts; the target
// stage and all subsequent non-skipped stages are re-run.
func (s *SDLCService) RetryFromStage(wfID, stageID, callerID uint64) (*SDLCWorkflowResponse, error) {
	var wf model.SDLCWorkflow
	if err := s.db.First(&wf, wfID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("SDLC workflow not found")
		}
		return nil, common.Internal("Failed to get SDLC workflow")
	}
	if err := s.checkWorkspaceAdmin(wf.WorkspaceID, callerID); err != nil {
		return nil, err
	}
	var stage model.SDLCStage
	if err := s.db.Where("workflow_id = ? AND id = ?", wfID, stageID).First(&stage).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("SDLC stage not found")
		}
		return nil, common.Internal("Failed to get SDLC stage")
	}
	if wf.Status != model.SDLCWorkflowFailed && wf.Status != model.SDLCWorkflowPartial {
		return nil, common.BadRequest("Only a failed or partial-failed workflow can be retried")
	}

	// Reset the target stage and every subsequent non-skipped stage to pending.
	if err := s.db.Model(&model.SDLCStage{}).
		Where("workflow_id = ? AND \"order\" >= ?", wfID, stage.Order).
		Updates(map[string]interface{}{
			"status":        model.SDLCStagePending,
			"progress":      0,
			"error_message": nil,
			"started_at":    nil,
			"completed_at":  nil,
			"duration_ms":   0,
		}).Error; err != nil {
		return nil, common.Internal("Failed to reset SDLC stages")
	}

	// Clear the workflow error and flip back to running.
	wf.Status = model.SDLCWorkflowRunning
	wf.ErrorMessage = nil
	stageName := stage.Name
	wf.CurrentStage = &stageName
	if err := s.db.Save(&wf).Error; err != nil {
		return nil, common.Internal("Failed to resume SDLC workflow")
	}

	// Decode the runtime config to honor fail_fast + selected stages.
	cfg := decodeWorkflowConfig(wf.Config)
	go s.runWorkflowFrom(wf.ID, stage.Order, cfg.FailFast, cfg.Stages)

	resp := s.workflowToResponse(&wf)
	s.pushWorkflowEvent("sdlc_workflow.resumed", resp)
	return resp, nil
}

// ======== Async orchestration ========

// runWorkflow executes the full pipeline from stage 1.
func (s *SDLCService) runWorkflow(wfID uint64, failFast bool, selected []string) {
	s.runWorkflowFrom(wfID, 1, failFast, selected)
}

// runWorkflowFrom executes the pipeline starting from a given stage order.
// It walks the canonical stages in order, skipping deselected/skipped stages,
// invoking the StageExecutor per stage, and merging artifacts. Between
// stages it re-reads the workflow to honor in-flight cancellation.
func (s *SDLCService) runWorkflowFrom(wfID uint64, fromOrder int, failFast bool, selected []string) {
	var wf model.SDLCWorkflow
	if err := s.db.First(&wf, wfID).Error; err != nil {
		return
	}
	if wf.Status == model.SDLCWorkflowCancelled {
		return
	}

	// Flip to running and mark the start.
	now := time.Now()
	wf.Status = model.SDLCWorkflowRunning
	if wf.StartedAt == nil {
		wf.StartedAt = &now
	}
	wf.ErrorMessage = nil
	s.db.Save(&wf)
	s.pushWorkflowEvent("sdlc_workflow.updated", s.workflowToResponse(&wf))

	// Load accumulated artifacts so resumed runs keep prior outputs.
	artifacts := decodeArtifacts(wf.Artifacts)

	failedCount := 0
	totalRunnable := 0
	for _, def := range canonicalSDLStages {
		if def.Order < fromOrder {
			continue
		}
		if len(selected) > 0 && !containsString(selected, def.Key) {
			continue
		}
		totalRunnable++
	}

	for _, def := range canonicalSDLStages {
		if def.Order < fromOrder {
			continue
		}
		// Honor cancellation between stages.
		if err := s.db.First(&wf, wfID).Error; err != nil {
			return
		}
		if wf.Status == model.SDLCWorkflowCancelled {
			return
		}

		// Skipped stages stay skipped.
		if len(selected) > 0 && !containsString(selected, def.Key) {
			continue
		}

		var stage model.SDLCStage
		if err := s.db.Where("workflow_id = ? AND \"order\" = ?", wfID, def.Order).First(&stage).Error; err != nil {
			continue
		}

		// Skip stages that already succeeded during a prior partial run
		// (only relevant when resuming from an earlier order).
		if stage.Status == model.SDLCStageSuccess {
			continue
		}

		stageErr := s.runStage(&wf, &stage, artifacts)
		if stageErr != nil {
			failedCount++
			if failFast {
				s.failWorkflow(&wf, fmt.Sprintf("Stage %q failed: %v", def.Name, stageErr))
				return
			}
			continue
		}
	}

	// All runnable stages attempted — finalize.
	s.finalizeWorkflow(wfID, failedCount, totalRunnable)
}

// runStage executes a single stage end-to-end: mark running, snapshot input,
// invoke executor, persist output/logs, merge artifacts, broadcast SSE.
// Returns the stage error (already persisted on the stage row) so the caller
// can apply fail_fast policy.
func (s *SDLCService) runStage(wf *model.SDLCWorkflow, stage *model.SDLCStage, artifacts map[string]interface{}) error {
	startedAt := time.Now()
	stage.Status = model.SDLCStageRunning
	stage.Progress = 10
	stage.StartedAt = &startedAt
	stage.ErrorMessage = nil
	// Snapshot the inputs handed to the stage (requirement + prior artifacts).
	inputSnapshot := map[string]interface{}{
		"title":       wf.Title,
		"requirement": wf.Requirement,
		"artifacts":   artifacts,
	}
	inputJSON, _ := json.Marshal(inputSnapshot)
	stage.Input = inputJSON
	s.db.Save(stage)
	s.pushStageEvent("sdlc_stage.updated", stage)

	// Build the executor view of the stage (decouple model pointer aliasing).
	execStage := *stage
	execStage.WorkflowID = wf.ID
	execWf := *wf

	output, logs, execErr := s.executor.Execute(context.Background(), &execWf, &execStage, artifacts)

	completedAt := time.Now()
	duration := completedAt.Sub(startedAt).Milliseconds()
	stage.DurationMs = duration
	stage.StartedAt = &startedAt
	stage.CompletedAt = &completedAt
	stage.Logs = marshalLogs(logs)

	if execErr != nil {
		msg := execErr.Error()
		stage.Status = model.SDLCStageFailed
		stage.Progress = 0
		stage.ErrorMessage = &msg
		stage.Output = marshalOutput(output)
		s.db.Save(stage)
		s.pushStageEvent("sdlc_stage.failed", stage)
		// Reflect the current stage on the workflow for visibility.
		wf.CurrentStage = &stage.Name
		wf.Progress = workflowProgress(stage.Order, len(canonicalSDLStages))
		s.db.Save(wf)
		s.pushWorkflowEvent("sdlc_workflow.updated", s.workflowToResponse(wf))
		return execErr
	}

	// Success — persist output and merge into accumulated artifacts.
	stage.Status = model.SDLCStageSuccess
	stage.Progress = 100
	stage.Output = marshalOutput(output)
	s.db.Save(stage)
	s.pushStageEvent("sdlc_stage.completed", stage)

	for k, v := range output {
		artifacts[k] = v
	}
	artifactsJSON, _ := json.Marshal(artifacts)
	wf.Artifacts = artifactsJSON
	wf.CurrentStage = &stage.Name
	wf.Progress = workflowProgress(stage.Order, len(canonicalSDLStages))
	s.db.Save(wf)
	s.pushWorkflowEvent("sdlc_workflow.updated", s.workflowToResponse(wf))
	return nil
}

// finalizeWorkflow sets the terminal status based on failure count.
func (s *SDLCService) finalizeWorkflow(wfID uint64, failedCount, totalRunnable int) {
	var wf model.SDLCWorkflow
	if err := s.db.First(&wf, wfID).Error; err != nil {
		return
	}
	if wf.Status == model.SDLCWorkflowCancelled {
		return
	}
	now := time.Now()
	wf.CompletedAt = &now
	wf.Progress = 100
	completedStage := "completed"
	switch {
	case failedCount == 0:
		wf.Status = model.SDLCWorkflowCompleted
	case failedCount >= totalRunnable:
		wf.Status = model.SDLCWorkflowFailed
		msg := fmt.Sprintf("%d of %d stage(s) failed", failedCount, totalRunnable)
		wf.ErrorMessage = &msg
		wf.CurrentStage = nil
	default:
		wf.Status = model.SDLCWorkflowPartial
		msg := fmt.Sprintf("%d of %d stage(s) failed", failedCount, totalRunnable)
		wf.ErrorMessage = &msg
		wf.CurrentStage = nil
	}
	if wf.Status == model.SDLCWorkflowCompleted {
		wf.CurrentStage = &completedStage
	}
	s.db.Save(&wf)
	event := "sdlc_workflow.updated"
	switch wf.Status {
	case model.SDLCWorkflowCompleted:
		event = "sdlc_workflow.completed"
	case model.SDLCWorkflowFailed:
		event = "sdlc_workflow.failed"
	}
	s.pushWorkflowEvent(event, s.workflowToResponse(&wf))
}

// failWorkflow marks the workflow as failed and emits an SSE event.
func (s *SDLCService) failWorkflow(wf *model.SDLCWorkflow, message string) {
	now := time.Now()
	wf.Status = model.SDLCWorkflowFailed
	wf.ErrorMessage = &message
	wf.CompletedAt = &now
	stage := "failed"
	wf.CurrentStage = &stage
	s.db.Save(wf)
	s.pushWorkflowEvent("sdlc_workflow.failed", s.workflowToResponse(wf))
}

// ======== Response builders ========

func (s *SDLCService) workflowToResponse(wf *model.SDLCWorkflow) *SDLCWorkflowResponse {
	cfg := wf.Config
	if len(cfg) == 0 {
		cfg = json.RawMessage("{}")
	}
	artifacts := wf.Artifacts
	if len(artifacts) == 0 {
		artifacts = json.RawMessage("{}")
	}
	resp := &SDLCWorkflowResponse{
		ID:           wf.ID,
		WorkspaceID:  wf.WorkspaceID,
		ProjectID:    wf.ProjectID,
		SquadID:      wf.SquadID,
		Title:        wf.Title,
		Requirement:  wf.Requirement,
		Status:       string(wf.Status),
		Progress:     wf.Progress,
		CurrentStage: wf.CurrentStage,
		Config:       cfg,
		Artifacts:    artifacts,
		ErrorMessage: wf.ErrorMessage,
		StartedAt:    wf.StartedAt,
		CompletedAt:  wf.CompletedAt,
		CancelledAt:  wf.CancelledAt,
		CreatedAt:    wf.CreatedAt,
		UpdatedAt:    wf.UpdatedAt,
	}
	if len(wf.Stages) > 0 {
		resp.Stages = make([]SDLCStageResponse, 0, len(wf.Stages))
		// Stages may be preloaded in arbitrary order; sort by Order for a stable view.
		sorted := append([]model.SDLCStage(nil), wf.Stages...)
		// Simple insertion sort by Order (small, fixed N=11).
		for i := 1; i < len(sorted); i++ {
			for j := i; j > 0 && sorted[j].Order < sorted[j-1].Order; j-- {
				sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
			}
		}
		for i := range sorted {
			resp.Stages = append(resp.Stages, *s.stageToResponse(&sorted[i]))
		}
	}
	return resp
}

// stageToResponse converts a model.SDLCStage to its API representation.
func (s *SDLCService) stageToResponse(stage *model.SDLCStage) *SDLCStageResponse {
	input := stage.Input
	if len(input) == 0 {
		input = json.RawMessage("{}")
	}
	output := stage.Output
	if len(output) == 0 {
		output = json.RawMessage("{}")
	}
	logs := stage.Logs
	if len(logs) == 0 {
		logs = json.RawMessage("[]")
	}
	return &SDLCStageResponse{
		ID:           stage.ID,
		WorkflowID:   stage.WorkflowID,
		WorkspaceID:  stage.WorkspaceID,
		Order:        stage.Order,
		Key:          stage.Key,
		Name:         stage.Name,
		AgentRole:    stage.AgentRole,
		Status:       string(stage.Status),
		Progress:     stage.Progress,
		Input:        input,
		Output:       output,
		Logs:         logs,
		ErrorMessage: stage.ErrorMessage,
		StartedAt:    stage.StartedAt,
		CompletedAt:  stage.CompletedAt,
		DurationMs:   stage.DurationMs,
		CreatedAt:    stage.CreatedAt,
		UpdatedAt:    stage.UpdatedAt,
	}
}

// pushWorkflowEvent broadcasts an sdlc_workflow.* SSE event.
func (s *SDLCService) pushWorkflowEvent(event string, resp *SDLCWorkflowResponse) {
	data, _ := json.Marshal(resp)
	SSE.BroadcastEvent(event, json.RawMessage(data))
}

// pushStageEvent broadcasts an sdlc_stage.* SSE event.
func (s *SDLCService) pushStageEvent(event string, stage *model.SDLCStage) {
	data, _ := json.Marshal(s.stageToResponse(stage))
	SSE.BroadcastEvent(event, json.RawMessage(data))
}

// ======== Helpers ========

// workflowConfig is the decoded form of SDLCWorkflow.Config.
type workflowConfig struct {
	Stages   []string `json:"stages"`
	FailFast bool     `json:"fail_fast"`
}

func decodeWorkflowConfig(raw json.RawMessage) workflowConfig {
	cfg := workflowConfig{FailFast: true, Stages: []string{}}
	if len(raw) == 0 {
		return cfg
	}
	_ = json.Unmarshal(raw, &cfg)
	if cfg.Stages == nil {
		cfg.Stages = []string{}
	}
	return cfg
}

func decodeArtifacts(raw json.RawMessage) map[string]interface{} {
	out := map[string]interface{}{}
	if len(raw) == 0 {
		return out
	}
	_ = json.Unmarshal(raw, &out)
	return out
}

// normalizeStageSelection returns nil (=> run all) when empty, otherwise the
// validated subset of canonical stage keys. Unknown keys are dropped.
func normalizeStageSelection(stages []string) []string {
	if len(stages) == 0 {
		return nil
	}
	valid := make(map[string]bool, len(canonicalSDLStages))
	for _, def := range canonicalSDLStages {
		valid[def.Key] = true
	}
	out := make([]string, 0, len(stages))
	for _, s := range stages {
		s = strings.TrimSpace(s)
		if valid[s] && !containsString(out, s) {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// isTerminalWorkflowStatus returns true for statuses that cannot transition further.
func isTerminalWorkflowStatus(s model.SDLCWorkflowStatus) bool {
	switch s {
	case model.SDLCWorkflowCompleted, model.SDLCWorkflowFailed, model.SDLCWorkflowPartial, model.SDLCWorkflowCancelled:
		return true
	}
	return false
}

// workflowProgress maps a completed stage order to a 0-100 percentage of the
// full pipeline (stage N of total → N/total*100).
func workflowProgress(order, total int) int {
	if total <= 0 {
		return 0
	}
	p := order * 100 / total
	if p > 100 {
		p = 100
	}
	return p
}

// marshalOutput serializes a stage output map to JSONB, defaulting to "{}".
func marshalOutput(out map[string]interface{}) json.RawMessage {
	if out == nil {
		return json.RawMessage("{}")
	}
	b, _ := json.Marshal(out)
	return b
}

// marshalLogs serializes a log slice to JSONB, defaulting to "[]".
func marshalLogs(logs []string) json.RawMessage {
	if logs == nil {
		return json.RawMessage("[]")
	}
	b, _ := json.Marshal(logs)
	return b
}

// ======== Stub stage executor ========

// stubSDLCStageExecutor simulates the 11 SDLC stages without external calls.
// Each stage produces a realistic artifact keyed off the workflow title/
// requirement so the pipeline is observable end-to-end without LLM, GitHub,
// or CI/CD dependencies (mirroring stubCICDProvider / stubGenerate).
//
// The stub sleeps briefly per stage so SSE progress is visible; tests that
// need determinism supply their own StageExecutor via SetStageExecutor.
type stubSDLCStageExecutor struct{}

func (e *stubSDLCStageExecutor) Execute(ctx context.Context, wf *model.SDLCWorkflow, stage *model.SDLCStage, prior map[string]interface{}) (map[string]interface{}, []string, error) {
	ts := time.Now().Format("15:04:05")
	logs := []string{
		fmt.Sprintf("[%s] %s (%s) started", ts, stage.Name, stage.AgentRole),
	}
	// Brief, cancellable delay so progress is observable without slowing tests
	// that supply their own executor.
	select {
	case <-time.After(300 * time.Millisecond):
	case <-ctx.Done():
		return nil, logs, ctx.Err()
	}

	out := stubStageOutput(stage, wf, prior)
	ts2 := time.Now().Format("15:04:05")
	logs = append(logs, fmt.Sprintf("[%s] %s produced: %s", ts2, stage.Name, summarizeOutput(out)))
	return out, logs, nil
}

// stubStageOutput returns the deterministic artifact map for a given stage.
func stubStageOutput(stage *model.SDLCStage, wf *model.SDLCWorkflow, prior map[string]interface{}) map[string]interface{} {
	_ = prior
	title := wf.Title
	if title == "" {
		title = "未命名需求"
	}
	switch stage.Key {
	case "requirement_analysis":
		return map[string]interface{}{
			"analysis_report": fmt.Sprintf("需求分析报告：%s\n- 功能范围：待定义\n- 优先级：高\n- 技术可行性：可行", title),
			"priority":        "high",
		}
	case "requirement_design":
		return map[string]interface{}{
			"prd_doc":      fmt.Sprintf("# %s PRD\n## 目标\n实现 %s 的核心能力。", title, title),
			"tech_design":  fmt.Sprintf("技术方案：%s 采用分层架构。", title),
		}
	case "dispatch_feature":
		return map[string]interface{}{
			"feature_issue_id": fakeID("feature"),
			"feature_title":    title,
		}
	case "feature_design":
		return map[string]interface{}{
			"feature_design_doc": fmt.Sprintf("功能设计：%s\n- 接口：/api/v1/...\n- 数据模型：见下文", title),
			"api_spec":           "GET /api/v1/items, POST /api/v1/items",
		}
	case "breakdown_us":
		return map[string]interface{}{
			"user_stories": []string{
				fmt.Sprintf("US-1: 作为用户，我可以使用 %s", title),
				"US-2: 作为管理员，我可以管理配置",
			},
		}
	case "sprint_planning":
		return map[string]interface{}{
			"sprint_name":  fmt.Sprintf("Sprint %s", fakeID("sprint")),
			"capacity":     20,
			"story_points": 8,
		}
	case "development":
		return map[string]interface{}{
			"branch":      fmt.Sprintf("feature/%s", slugify(title)),
			"commit_sha":  fakeID("commit"),
			"pr_url":      fmt.Sprintf("https://github.com/example/repo/pull/%s", fakeID("pr")),
			"pr_number":   42,
		}
	case "code_review":
		return map[string]interface{}{
			"review_result": "approved",
			"comments":      "代码质量良好，符合规范。",
		}
	case "us_testing":
		return map[string]interface{}{
			"test_report": "全部测试通过 (12/12)",
			"passed":      12,
			"failed":      0,
			"bugs":        []interface{}{},
		}
	case "fe_testing":
		return map[string]interface{}{
			"e2e_report": "E2E 测试通过 (5/5)",
			"passed":     5,
			"failed":     0,
		}
	case "deploy":
		return map[string]interface{}{
			"deploy_url":       fmt.Sprintf("https://app.example.com/%s", slugify(title)),
			"release_version":  fmt.Sprintf("v1.0.0-%s", fakeID("release")),
			"deploy_status":    "success",
		}
	}
	return map[string]interface{}{
		"note": fmt.Sprintf("stage %s produced no artifact", stage.Key),
	}
}

// summarizeOutput returns a short human-readable digest of a stage output for logs.
func summarizeOutput(out map[string]interface{}) string {
	if len(out) == 0 {
		return "(empty)"
	}
	keys := make([]string, 0, len(out))
	for k := range out {
		keys = append(keys, k)
	}
	return strings.Join(keys, ", ")
}

// fakeID returns a pseudo identifier derived from a prefix + the current
// nanosecond timestamp so artifacts look realistic without being constant.
func fakeID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano()%100000)
}

// slugify converts a title into a git-safe slug for the stub.
func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		out = "feature"
	}
	if len(out) > 40 {
		out = out[:40]
	}
	return out
}

// Compile-time interface check.
var _ StageExecutor = (*stubSDLCStageExecutor)(nil)
