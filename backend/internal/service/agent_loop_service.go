package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	agentloop "github.com/reqmango/backend/internal/agent/loop"
	agentmodel "github.com/reqmango/backend/internal/agent/model"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// LoopService manages Loop CRUD and execution.
type LoopService struct {
	db       *gorm.DB
	agentSvc *AgentService
}

func NewLoopService(db *gorm.DB, agentSvc *AgentService) *LoopService {
	return &LoopService{db: db, agentSvc: agentSvc}
}

// CreateLoop persists a new Loop definition.
func (s *LoopService) CreateLoop(workspaceID, userID uint64, name, description string, loopDef json.RawMessage) (*agentmodel.Loop, error) {
	loop := &agentmodel.Loop{
		WorkspaceID: workspaceID,
		Name:        name,
		LoopDef:     loopDef,
		CreatedByID: &userID,
		Status:      "active",
	}
	if description != "" {
		loop.Description = &description
	}
	if err := s.db.Create(loop).Error; err != nil {
		return nil, common.Internal(fmt.Sprintf("failed to create loop: %v", err))
	}
	return loop, nil
}

// ListLoops returns all active Loops for a workspace.
func (s *LoopService) ListLoops(workspaceID uint64) ([]agentmodel.Loop, error) {
	var loops []agentmodel.Loop
	if err := s.db.Where("workspace_id = ? AND status != ?", workspaceID, "archived").
		Order("created_at DESC").Find(&loops).Error; err != nil {
		return nil, common.Internal(fmt.Sprintf("failed to list loops: %v", err))
	}
	return loops, nil
}

// GetLoop retrieves a single Loop by ID.
func (s *LoopService) GetLoop(workspaceID, loopID uint64) (*agentmodel.Loop, error) {
	var loop agentmodel.Loop
	if err := s.db.Where("id = ? AND workspace_id = ?", loopID, workspaceID).First(&loop).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("loop not found")
		}
		return nil, common.Internal(fmt.Sprintf("failed to get loop: %v", err))
	}
	return &loop, nil
}

// UpdateLoop updates a Loop definition.
func (s *LoopService) UpdateLoop(workspaceID, loopID uint64, name, description *string, loopDef json.RawMessage, status *string) (*agentmodel.Loop, error) {
	loop, err := s.GetLoop(workspaceID, loopID)
	if err != nil {
		return nil, err
	}
	if name != nil {
		loop.Name = *name
	}
	if description != nil {
		loop.Description = description
	}
	if loopDef != nil {
		loop.LoopDef = loopDef
	}
	if status != nil {
		loop.Status = *status
	}
	if err := s.db.Save(loop).Error; err != nil {
		return nil, common.Internal(fmt.Sprintf("failed to update loop: %v", err))
	}
	return loop, nil
}

// DeleteLoop soft-deletes a Loop.
func (s *LoopService) DeleteLoop(workspaceID, loopID uint64) error {
	result := s.db.Where("id = ? AND workspace_id = ?", loopID, workspaceID).Delete(&agentmodel.Loop{})
	if result.RowsAffected == 0 {
		return common.NotFound("loop not found")
	}
	return result.Error
}

// StartLoop creates a LoopRun and starts execution in a background goroutine.
func (s *LoopService) StartLoop(workspaceID, loopID uint64, userID uint64) (*agentmodel.LoopRun, error) {
	loop, err := s.GetLoop(workspaceID, loopID)
	if err != nil {
		return nil, err
	}

	var def struct {
		Goal           string  `json:"goal"`
		MaxIterations  int     `json:"max_iterations"`
		MaxTokens      int     `json:"max_tokens"`
		MaxCost        float64 `json:"max_cost"`
		MaxDurationSec int     `json:"max_duration_sec"`
	}
	if err := json.Unmarshal(loop.LoopDef, &def); err != nil {
		return nil, common.Validation("invalid loop definition")
	}

	if def.MaxIterations == 0 {
		def.MaxIterations = 10
	}
	if def.MaxTokens == 0 {
		def.MaxTokens = 50000
	}
	if def.MaxDurationSec == 0 {
		def.MaxDurationSec = 3600
	}

	run := &agentmodel.LoopRun{
		LoopID:        loopID,
		Status:        "running",
		MaxIterations: def.MaxIterations,
		Goal:          def.Goal,
		WorkingMemory: json.RawMessage("{}"),
		StartedAt:     time.Now(),
	}
	if err := s.db.Create(run).Error; err != nil {
		return nil, common.Internal(fmt.Sprintf("failed to create loop run: %v", err))
	}

	executor := &loopAgentExecutor{
		agentSvc:    s.agentSvc,
		workspaceID: workspaceID,
		userID:      userID,
	}
	collector := &loopMetricsCollector{db: s.db, workspaceID: workspaceID}
	evaluator := &loopGoalEvaluator{}

	config := agentloop.LoopConfig{
		Goal:          def.Goal,
		MaxIterations: def.MaxIterations,
		MaxTokens:     def.MaxTokens,
		MaxCost:       def.MaxCost,
		MaxDuration:   time.Duration(def.MaxDurationSec) * time.Second,
	}

	runner := agentloop.NewLoopRunner(s.db, executor, collector, evaluator, config)

	go func() {
		ctx := context.Background()
		iterations, stopReason, runErr := runner.Run(ctx)

		now := time.Now()
		run.CompletedAt = &now
		run.StoppedReason = &stopReason
		run.TokensUsed = runner.Budget().UsedTokens
		run.CostUSD = runner.Budget().UsedCost

		if runErr != nil || (stopReason != "goal_achieved" && stopReason != "cancelled") {
			run.Status = "failed"
		} else {
			run.Status = "completed"
		}
		s.db.Save(run)

		for _, iter := range iterations {
			actionJSON, _ := json.Marshal(map[string]string{"task": iter.Action, "result": iter.Result})
			resultJSON, _ := json.Marshal(iter.Metrics)
			reasoning := iter.Reasoning
			dur := iter.DurationMs
			iteration := &agentmodel.LoopIteration{
				LoopRunID:      run.ID,
				IterationNum:   iter.Num,
				ActionTaken:    actionJSON,
				ResultObserved: resultJSON,
				Reasoning:      &reasoning,
				Decision:       string(iter.Decision),
				TokensUsed:     iter.TokensUsed,
				DurationMs:     &dur,
			}
			s.db.Create(iteration)
		}

		s.recordSession(workspaceID, runner.SessionID(), stopReason, runner.Budget().UsedTokens, runner.Budget().UsedCost, runErr)
	}()

	return run, nil
}

// StopLoop marks a running LoopRun as stopped.
func (s *LoopService) StopLoop(workspaceID, runID uint64) error {
	reason := "manual"
	now := time.Now()
	result := s.db.Model(&agentmodel.LoopRun{}).
		Where("id = ? AND status = ?", runID, "running").
		Updates(map[string]interface{}{
			"status":         "stopped",
			"stopped_reason": &reason,
			"completed_at":   &now,
		})
	if result.RowsAffected == 0 {
		return common.NotFound("loop run not found or not running")
	}
	return result.Error
}

// GetLoopRuns returns run history for a Loop.
func (s *LoopService) GetLoopRuns(workspaceID, loopID uint64, limit int) ([]agentmodel.LoopRun, error) {
	if limit == 0 {
		limit = 20
	}
	var runs []agentmodel.LoopRun
	if err := s.db.Where("loop_id = ?", loopID).
		Order("started_at DESC").Limit(limit).Find(&runs).Error; err != nil {
		return nil, common.Internal(fmt.Sprintf("failed to get loop runs: %v", err))
	}
	return runs, nil
}

// GetLoopRun retrieves a single run with its iterations.
func (s *LoopService) GetLoopRun(workspaceID, runID uint64) (*agentmodel.LoopRun, []agentmodel.LoopIteration, error) {
	var run agentmodel.LoopRun
	if err := s.db.Where("id = ?", runID).First(&run).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, common.NotFound("loop run not found")
		}
		return nil, nil, err
	}
	var iterations []agentmodel.LoopIteration
	s.db.Where("loop_run_id = ?", runID).Order("iteration_num ASC").Find(&iterations)
	return &run, iterations, nil
}

func (s *LoopService) recordSession(workspaceID uint64, sessionID, reason string, tokens int, cost float64, runErr error) {
	status := "completed"
	var errMsg *string
	if runErr != nil || (reason != "goal_achieved" && reason != "cancelled") {
		status = "failed"
		msg := reason
		errMsg = &msg
	}
	now := time.Now()
	summary := fmt.Sprintf("Loop run completed: %s", reason)
	session := &agentmodel.AgentSession{
		ID:            sessionID,
		WorkspaceID:   workspaceID,
		AgentType:     "loop_iteration",
		Status:        status,
		InputSummary:  &summary,
		OutputSummary: &reason,
		TokensInput:   tokens / 2,
		TokensOutput:  tokens / 2,
		CostUSD:       cost,
		ErrorMessage:  errMsg,
		StartedAt:     time.Now().Add(-1 * time.Minute),
		CompletedAt:   &now,
	}
	s.db.Create(session)
}

// --- Internal adapters ---

type loopAgentExecutor struct {
	agentSvc    *AgentService
	workspaceID uint64
	userID      uint64
}

func (e *loopAgentExecutor) Execute(ctx context.Context, task string, context map[string]interface{}) (string, int, float64, error) {
	agents, err := e.agentSvc.ListByWorkspace(e.workspaceID)
	if err != nil || len(agents) == 0 {
		return "", 0, 0, fmt.Errorf("no agents available: %w", err)
	}

	var runner *model.Agent
	for i := range agents {
		if agents[i].Status == "active" {
			runner = &agents[i]
			break
		}
	}
	if runner == nil {
		return "", 0, 0, fmt.Errorf("no active agent found")
	}

	var issueIDPtr *uint64
	if issueID, ok := context["issue_id"].(uint64); ok {
		issueIDPtr = &issueID
	}
	var projectIDPtr *uint64
	if projectID, ok := context["project_id"].(uint64); ok {
		projectIDPtr = &projectID
	}

	dispCtx := &DispatchContext{
		IssueID:     issueIDPtr,
		ProjectID:   projectIDPtr,
		WorkspaceID: e.workspaceID,
		TriggeredBy: "loop",
	}

	activity, err := e.agentSvc.DispatchAgent(runner.ID, e.userID, task, dispCtx)
	if err != nil {
		return "", 0, 0, err
	}

	result := activity.ResultSummary
	tokens := 500
	cost := float64(tokens) * 0.000002

	return result, tokens, cost, nil
}

type loopMetricsCollector struct {
	db          *gorm.DB
	workspaceID uint64
}

func (c *loopMetricsCollector) Collect(ctx context.Context, context map[string]interface{}) (map[string]float64, error) {
	metrics := make(map[string]float64)
	if cycleID, ok := context["cycle_id"]; ok {
		var total, completed int64
		c.db.Table("issues").Where("cycle_id = ? AND deleted_at IS NULL", cycleID).Count(&total)
		c.db.Table("issues").Where("cycle_id = ? AND state_group = 'completed' AND deleted_at IS NULL", cycleID).Count(&completed)
		if total > 0 {
			metrics["progress"] = float64(completed) / float64(total)
		} else {
			metrics["progress"] = 0
		}
	}
	return metrics, nil
}

type loopGoalEvaluator struct{}

func (e *loopGoalEvaluator) Evaluate(goal string, metrics map[string]float64) (bool, string) {
	if progress, ok := metrics["progress"]; ok {
		if progress >= 0.9 {
			return true, fmt.Sprintf("progress %.1f%% meets 90%% target", progress*100)
		}
		return false, fmt.Sprintf("progress %.1f%% below 90%% target", progress*100)
	}
	return false, "no progress metric found"
}
