package service

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// AutopilotService implements PRD P4-008: 自动化任务调度与触发.
//
// The service manages AutopilotTask rows (workspace-scoped, optionally
// project-scoped) and their execution records. Tasks can be triggered three
// ways: manually via the API, on a cron schedule, or via a public webhook
// URL. Each execution runs asynchronously and broadcasts progress via SSE as
// autopilot_execution.* events so the UI can render live status without
// polling.
//
// The execution logic is abstracted behind AutopilotExecutor so the pipeline
// remains end-to-end testable without external agent dependencies (mirroring
// the CICDProvider / StageExecutor patterns used by P4-004/006). When no
// executor is configured a stub executor produces realistic per-task-type
// output.
type AutopilotService struct {
	db       *gorm.DB
	executor AutopilotExecutor
}

// AutopilotExecutor abstracts the body of an autopilot run so the service can
// be unit-tested without dispatching a real agent task. Implementations
// receive the task and the execution record and return the output payload,
// log lines, and an error (which marks the execution failed).
type AutopilotExecutor interface {
	Execute(task *model.AutopilotTask, exec *model.AutopilotExecution) (output map[string]interface{}, logs []string, err error)
}

// NewAutopilotService creates a new AutopilotService. A stub executor is used
// by default so the workflow remains observable end-to-end without external
// dependencies.
func NewAutopilotService(db *gorm.DB) *AutopilotService {
	svc := &AutopilotService{db: db}
	svc.executor = &stubAutopilotExecutor{}
	return svc
}

// SetExecutor overrides the default executor. Used in tests.
func (s *AutopilotService) SetExecutor(e AutopilotExecutor) {
	if e != nil {
		s.executor = e
	}
}

// Autopilot task/execution status constants.
const (
	autopilotTaskActive   = "active"
	autopilotTaskPaused   = "paused"
	autopilotExecPending  = "pending"
	autopilotExecRunning  = "running"
	autopilotExecComplete = "completed"
	autopilotExecFailed   = "failed"
)

// Valid trigger types for an autopilot task.
var autopilotTriggerTypes = map[string]bool{
	"cron": true, "webhook": true, "manual": true,
}

// checkWorkspaceAdmin mirrors the guard used by other workspace services.
func (s *AutopilotService) checkWorkspaceAdmin(workspaceID, callerID uint64) error {
	var member model.WorkspaceMember
	if err := s.db.Where("workspace_id = ? AND user_id = ? AND is_active = ?", workspaceID, callerID, true).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.Forbidden("You must be a workspace admin to manage autopilot tasks")
		}
		return common.Internal("Database error")
	}
	if member.Role < common.RoleAdmin {
		return common.Forbidden("You must be a workspace admin to manage autopilot tasks")
	}
	return nil
}

// ======== CRUD ========

// CreateTask persists a new AutopilotTask. For webhook-triggered tasks a
// unique trigger URL is generated; for cron tasks the next run time is
// computed from the cron expression.
func (s *AutopilotService) CreateTask(wid, callerID uint64, req request.AutopilotTaskCreate) (*response.AutopilotTaskResponse, error) {
	if err := s.checkWorkspaceAdmin(wid, callerID); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, common.BadRequest("Name is required")
	}
	if !autopilotTriggerTypes[req.TriggerType] {
		return nil, common.BadRequest("Unsupported trigger_type: " + req.TriggerType)
	}
	if strings.TrimSpace(req.TaskType) == "" {
		return nil, common.BadRequest("task_type is required")
	}
	if req.TriggerType == "cron" {
		if strings.TrimSpace(req.CronExpression) == "" {
			return nil, common.BadRequest("cron_expression is required for cron trigger type")
		}
		if _, err := parseCron(req.CronExpression); err != nil {
			return nil, common.BadRequest("Invalid cron_expression: " + err.Error())
		}
	}

	task := &model.AutopilotTask{
		WorkspaceID:    wid,
		ProjectID:      req.ProjectID,
		Name:           req.Name,
		Description:    req.Description,
		TriggerType:    req.TriggerType,
		CronExpression: req.CronExpression,
		TaskType:       req.TaskType,
		AgentTemplateID: req.AgentTemplateID,
		AgentConfigID:   req.AgentConfigID,
		InputData:       normalizeAutopilotJSON(req.InputData, "{}"),
		Config:          normalizeAutopilotJSON(req.Config, "{}"),
		NotificationConfig: normalizeAutopilotJSON(req.NotificationConfig, "{}"),
		TimeoutSeconds:  req.TimeoutSeconds,
		RetryCount:      req.RetryCount,
		Status:          autopilotTaskActive,
		Enabled:         req.Enabled,
	}
	if req.TriggerType == "webhook" {
		task.TriggerURL = fmt.Sprintf("/api/v1/autopilot/webhook/%s", generateWebhookToken())
	}
	if req.TriggerType == "cron" && req.CronExpression != "" {
		if next, err := calculateNextRun(req.CronExpression, time.Now()); err == nil {
			task.NextRunAt = &next
		}
	}

	if err := s.db.Create(task).Error; err != nil {
		return nil, common.Internal("Failed to create autopilot task")
	}

	resp := s.buildTaskResponse(task)
	s.pushTaskEvent("autopilot_task.created", resp)
	return resp, nil
}

// GetTask returns a single AutopilotTask by ID.
func (s *AutopilotService) GetTask(id uint64) (*response.AutopilotTaskResponse, error) {
	var task model.AutopilotTask
	if err := s.db.First(&task, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("Autopilot task not found")
		}
		return nil, common.Internal("Failed to get autopilot task")
	}
	return s.buildTaskResponse(&task), nil
}

// ListTasks returns AutopilotTasks for a workspace, newest first.
// Optional filters: projectID, status.
func (s *AutopilotService) ListTasks(wid uint64, projectID *uint64, status string) ([]*response.AutopilotTaskResponse, error) {
	q := s.db.Where("workspace_id = ?", wid)
	if projectID != nil && *projectID != 0 {
		q = q.Where("project_id = ?", projectID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var tasks []model.AutopilotTask
	if err := q.Order("created_at DESC").Find(&tasks).Error; err != nil {
		return nil, common.Internal("Failed to list autopilot tasks")
	}
	out := make([]*response.AutopilotTaskResponse, 0, len(tasks))
	for i := range tasks {
		out = append(out, s.buildTaskResponse(&tasks[i]))
	}
	return out, nil
}

// UpdateTask applies a partial update to an AutopilotTask.
func (s *AutopilotService) UpdateTask(id, callerID uint64, req request.AutopilotTaskUpdate) (*response.AutopilotTaskResponse, error) {
	var task model.AutopilotTask
	if err := s.db.First(&task, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("Autopilot task not found")
		}
		return nil, common.Internal("Failed to get autopilot task")
	}
	if err := s.checkWorkspaceAdmin(task.WorkspaceID, callerID); err != nil {
		return nil, err
	}
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return nil, common.BadRequest("Name cannot be empty")
		}
		task.Name = *req.Name
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.CronExpression != nil {
		if strings.TrimSpace(*req.CronExpression) == "" {
			return nil, common.BadRequest("cron_expression cannot be empty")
		}
		if _, err := parseCron(*req.CronExpression); err != nil {
			return nil, common.BadRequest("Invalid cron_expression: " + err.Error())
		}
		task.CronExpression = *req.CronExpression
		if next, err := calculateNextRun(task.CronExpression, time.Now()); err == nil {
			task.NextRunAt = &next
		}
	}
	if req.Enabled != nil {
		task.Enabled = *req.Enabled
	}
	if req.InputData != nil {
		task.InputData = normalizeAutopilotJSON(req.InputData, "{}")
	}
	if req.Config != nil {
		task.Config = normalizeAutopilotJSON(req.Config, "{}")
	}

	if err := s.db.Save(&task).Error; err != nil {
		return nil, common.Internal("Failed to update autopilot task")
	}

	resp := s.buildTaskResponse(&task)
	s.pushTaskEvent("autopilot_task.updated", resp)
	return resp, nil
}

// DeleteTask removes an AutopilotTask (soft-delete via BaseModel).
func (s *AutopilotService) DeleteTask(id, callerID uint64) error {
	var task model.AutopilotTask
	if err := s.db.First(&task, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.NotFound("Autopilot task not found")
		}
		return common.Internal("Failed to get autopilot task")
	}
	if err := s.checkWorkspaceAdmin(task.WorkspaceID, callerID); err != nil {
		return err
	}
	if err := s.db.Delete(&task).Error; err != nil {
		return common.Internal("Failed to delete autopilot task")
	}
	s.pushTaskEvent("autopilot_task.deleted", s.buildTaskResponse(&task))
	return nil
}

// ToggleTask flips the Enabled flag on an AutopilotTask.
func (s *AutopilotService) ToggleTask(id, callerID uint64) (*response.AutopilotTaskResponse, error) {
	var task model.AutopilotTask
	if err := s.db.First(&task, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("Autopilot task not found")
		}
		return nil, common.Internal("Failed to get autopilot task")
	}
	if err := s.checkWorkspaceAdmin(task.WorkspaceID, callerID); err != nil {
		return nil, err
	}
	task.Enabled = !task.Enabled
	if err := s.db.Save(&task).Error; err != nil {
		return nil, common.Internal("Failed to toggle autopilot task")
	}
	resp := s.buildTaskResponse(&task)
	s.pushTaskEvent("autopilot_task.toggled", resp)
	return resp, nil
}

// ======== Execution ========

// ExecuteTask starts a new execution for the given task. The execution record
// is persisted synchronously (status=running) and the body runs
// asynchronously so the API can return immediately. Callers poll the
// execution status or subscribe to autopilot_execution.* SSE events for
// completion.
func (s *AutopilotService) ExecuteTask(id, callerID uint64, triggerType string) (*response.AutopilotExecutionResponse, error) {
	var task model.AutopilotTask
	if err := s.db.First(&task, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("Autopilot task not found")
		}
		return nil, common.Internal("Failed to get autopilot task")
	}
	if err := s.checkWorkspaceAdmin(task.WorkspaceID, callerID); err != nil {
		return nil, err
	}
	if !task.Enabled {
		return nil, common.BadRequest("Task is disabled")
	}

	now := time.Now()
	exec := &model.AutopilotExecution{
		TaskID:      id,
		Status:      autopilotExecRunning,
		TriggerType: triggerType,
		InputData:   normalizeAutopilotJSON(nil, "{}"),
		Logs:        json.RawMessage("[]"),
		StartedAt:   &now,
	}
	if len(task.InputData) > 0 {
		exec.InputData = task.InputData
	}

	if err := s.db.Create(exec).Error; err != nil {
		return nil, common.Internal("Failed to create autopilot execution")
	}

	// Update task last run time + next run for cron tasks.
	task.LastRunAt = &now
	if task.TriggerType == "cron" && task.CronExpression != "" {
		if next, err := calculateNextRun(task.CronExpression, now); err == nil {
			task.NextRunAt = &next
		}
	}
	s.db.Save(&task)

	go s.executeTaskLogic(exec.ID, task.ID)

	resp := s.buildExecutionResponse(exec)
	s.pushExecutionEvent("autopilot_execution.started", resp)
	return resp, nil
}

// TriggerWebhook locates a task by its webhook token and executes it. The
// token is the trailing path segment of the task's TriggerURL. This endpoint
// is public (no auth) so external systems can invoke it.
func (s *AutopilotService) TriggerWebhook(token string) (*response.AutopilotExecutionResponse, error) {
	if strings.TrimSpace(token) == "" {
		return nil, common.BadRequest("Webhook token is required")
	}
	var task model.AutopilotTask
	// Match by the token suffix of trigger_url to avoid coupling to the
	// exact URL prefix.
	if err := s.db.Where("trigger_url LIKE ?", "%/"+token).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("No autopilot task for this webhook token")
		}
		return nil, common.Internal("Failed to lookup autopilot task")
	}
	if task.TriggerType != "webhook" {
		return nil, common.BadRequest("Task is not webhook-triggered")
	}
	if !task.Enabled {
		return nil, common.BadRequest("Task is disabled")
	}

	now := time.Now()
	exec := &model.AutopilotExecution{
		TaskID:      task.ID,
		Status:      autopilotExecRunning,
		TriggerType: "webhook",
		InputData:   normalizeAutopilotJSON(nil, "{}"),
		Logs:        json.RawMessage("[]"),
		StartedAt:   &now,
	}
	if len(task.InputData) > 0 {
		exec.InputData = task.InputData
	}
	if err := s.db.Create(exec).Error; err != nil {
		return nil, common.Internal("Failed to create autopilot execution")
	}
	task.LastRunAt = &now
	s.db.Save(&task)

	go s.executeTaskLogic(exec.ID, task.ID)

	resp := s.buildExecutionResponse(exec)
	s.pushExecutionEvent("autopilot_execution.started", resp)
	return resp, nil
}

// executeTaskLogic runs the task body asynchronously: invoke the executor,
// persist output/logs, and transition the execution to a terminal status.
// Errors mark the execution failed with the error message captured.
func (s *AutopilotService) executeTaskLogic(executionID, taskID uint64) {
	var task model.AutopilotTask
	if err := s.db.First(&task, taskID).Error; err != nil {
		s.failExecution(executionID, fmt.Sprintf("Task not found: %v", err))
		return
	}
	var exec model.AutopilotExecution
	if err := s.db.First(&exec, executionID).Error; err != nil {
		return
	}

	output, logs, execErr := s.executor.Execute(&task, &exec)

	completedAt := time.Now()
	durationMs := int64(0)
	if exec.StartedAt != nil {
		durationMs = completedAt.Sub(*exec.StartedAt).Milliseconds()
	}
	exec.CompletedAt = &completedAt
	exec.DurationMs = durationMs
	exec.Logs = marshalAutopilotLogs(logs)

	if execErr != nil {
		msg := execErr.Error()
		exec.Status = autopilotExecFailed
		exec.FailedAt = &completedAt
		exec.ErrorInfo = msg
		exec.OutputData = marshalAutopilotOutput(output)
		s.db.Save(&exec)
		s.pushExecutionEvent("autopilot_execution.failed", s.buildExecutionResponse(&exec))
		return
	}

	exec.Status = autopilotExecComplete
	exec.OutputData = marshalAutopilotOutput(output)
	s.db.Save(&exec)
	s.pushExecutionEvent("autopilot_execution.completed", s.buildExecutionResponse(&exec))
}

// failExecution marks an execution as failed when the task cannot be loaded.
func (s *AutopilotService) failExecution(executionID uint64, message string) {
	var exec model.AutopilotExecution
	if err := s.db.First(&exec, executionID).Error; err != nil {
		return
	}
	now := time.Now()
	exec.Status = autopilotExecFailed
	exec.FailedAt = &now
	exec.ErrorInfo = message
	if exec.StartedAt != nil {
		exec.DurationMs = now.Sub(*exec.StartedAt).Milliseconds()
	}
	s.db.Save(&exec)
	s.pushExecutionEvent("autopilot_execution.failed", s.buildExecutionResponse(&exec))
}

// GetExecution returns a single AutopilotExecution by ID.
func (s *AutopilotService) GetExecution(id uint64) (*response.AutopilotExecutionResponse, error) {
	var exec model.AutopilotExecution
	if err := s.db.First(&exec, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NotFound("Autopilot execution not found")
		}
		return nil, common.Internal("Failed to get autopilot execution")
	}
	return s.buildExecutionResponse(&exec), nil
}

// ListExecutions returns executions for a task, newest first.
func (s *AutopilotService) ListExecutions(taskID uint64) ([]*response.AutopilotExecutionResponse, error) {
	var execs []model.AutopilotExecution
	if err := s.db.Where("task_id = ?", taskID).Order("created_at DESC").Find(&execs).Error; err != nil {
		return nil, common.Internal("Failed to list autopilot executions")
	}
	out := make([]*response.AutopilotExecutionResponse, 0, len(execs))
	for i := range execs {
		out = append(out, s.buildExecutionResponse(&execs[i]))
	}
	return out, nil
}

// ======== Response builders ========

func (s *AutopilotService) buildTaskResponse(task *model.AutopilotTask) *response.AutopilotTaskResponse {
	return &response.AutopilotTaskResponse{
		ID:                 task.ID,
		WorkspaceID:        task.WorkspaceID,
		ProjectID:          task.ProjectID,
		Name:               task.Name,
		Description:        task.Description,
		TriggerType:        task.TriggerType,
		CronExpression:     task.CronExpression,
		TriggerURL:         task.TriggerURL,
		TaskType:           task.TaskType,
		AgentTemplateID:    task.AgentTemplateID,
		AgentConfigID:      task.AgentConfigID,
		InputData:          normalizeAutopilotJSONRaw(task.InputData, "{}"),
		Status:             task.Status,
		LastRunAt:          task.LastRunAt,
		NextRunAt:          task.NextRunAt,
		Config:             normalizeAutopilotJSONRaw(task.Config, "{}"),
		NotificationConfig: normalizeAutopilotJSONRaw(task.NotificationConfig, "{}"),
		TimeoutSeconds:     task.TimeoutSeconds,
		RetryCount:         task.RetryCount,
		Enabled:            task.Enabled,
		CreatedAt:          task.CreatedAt,
		UpdatedAt:          task.UpdatedAt,
	}
}

func (s *AutopilotService) buildExecutionResponse(exec *model.AutopilotExecution) *response.AutopilotExecutionResponse {
	return &response.AutopilotExecutionResponse{
		ID:          exec.ID,
		TaskID:      exec.TaskID,
		Status:      exec.Status,
		TriggerType: exec.TriggerType,
		InputData:   normalizeAutopilotJSONRaw(exec.InputData, "{}"),
		OutputData:  normalizeAutopilotJSONRaw(exec.OutputData, "{}"),
		ErrorInfo:   exec.ErrorInfo,
		Logs:        normalizeAutopilotJSONRaw(exec.Logs, "[]"),
		StartedAt:   exec.StartedAt,
		CompletedAt: exec.CompletedAt,
		FailedAt:    exec.FailedAt,
		DurationMs:  exec.DurationMs,
		RetryCount:  exec.RetryCount,
		CreatedAt:   exec.CreatedAt,
	}
}

// pushTaskEvent broadcasts an autopilot_task.* SSE event.
func (s *AutopilotService) pushTaskEvent(event string, resp *response.AutopilotTaskResponse) {
	data, _ := json.Marshal(resp)
	SSE.BroadcastEvent(event, json.RawMessage(data))
}

// pushExecutionEvent broadcasts an autopilot_execution.* SSE event.
func (s *AutopilotService) pushExecutionEvent(event string, resp *response.AutopilotExecutionResponse) {
	data, _ := json.Marshal(resp)
	SSE.BroadcastEvent(event, json.RawMessage(data))
}

// ======== Helpers ========

// generateWebhookToken returns a unique, unguessable token for a webhook URL.
// It uses crypto/rand so the token cannot be predicted and does not collide
// across rapid successive calls (time.Now() has limited resolution on some
// platforms, e.g. Windows).
func generateWebhookToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback to a timestamp-derived token if crypto/rand fails.
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// normalizeAutopilotJSON marshals a map to JSONB, defaulting to fallback when
// nil/empty so the column never stores SQL NULL.
func normalizeAutopilotJSON(in map[string]interface{}, fallback string) json.RawMessage {
	if in == nil || len(in) == 0 {
		return json.RawMessage(fallback)
	}
	b, _ := json.Marshal(in)
	return b
}

// normalizeAutopilotJSONRaw ensures a stored JSONB column is never empty;
// empty/nil values are replaced with the fallback.
func normalizeAutopilotJSONRaw(raw json.RawMessage, fallback string) json.RawMessage {
	if len(raw) == 0 || string(raw) == "null" {
		return json.RawMessage(fallback)
	}
	return raw
}

// marshalAutopilotOutput serializes an executor output map to JSONB.
func marshalAutopilotOutput(out map[string]interface{}) json.RawMessage {
	if out == nil {
		return json.RawMessage("{}")
	}
	b, _ := json.Marshal(out)
	return b
}

// marshalAutopilotLogs serializes a log slice to JSONB.
func marshalAutopilotLogs(logs []string) json.RawMessage {
	if logs == nil {
		return json.RawMessage("[]")
	}
	b, _ := json.Marshal(logs)
	return b
}

// ======== Cron parsing ========
//
// A minimal but correct 5-field cron parser (minute hour day-of-month month
// day-of-week). Supports: * , - and */step. Next-run is computed by scanning
// forward minute-by-minute up to 5 years from the reference time, which is
// more than enough for any practical schedule and keeps the implementation
// small and dependency-free.

// cronSchedule is the parsed representation of a 5-field cron expression.
type cronSchedule struct {
	minute, hour, dom, month, dow []int
}

// parseCron parses a 5-field cron expression into a cronSchedule.
func parseCron(expr string) (*cronSchedule, error) {
	fields := strings.Fields(strings.TrimSpace(expr))
	if len(fields) != 5 {
		return nil, fmt.Errorf("expected 5 fields, got %d", len(fields))
	}
	s := &cronSchedule{}
	var err error
	if s.minute, err = parseCronField(fields[0], 0, 59); err != nil {
		return nil, fmt.Errorf("minute: %w", err)
	}
	if s.hour, err = parseCronField(fields[1], 0, 23); err != nil {
		return nil, fmt.Errorf("hour: %w", err)
	}
	if s.dom, err = parseCronField(fields[2], 1, 31); err != nil {
		return nil, fmt.Errorf("day-of-month: %w", err)
	}
	if s.month, err = parseCronField(fields[3], 1, 12); err != nil {
		return nil, fmt.Errorf("month: %w", err)
	}
	if s.dow, err = parseCronField(fields[4], 0, 6); err != nil {
		return nil, fmt.Errorf("day-of-week: %w", err)
	}
	return s, nil
}

// parseCronField parses one cron field into the sorted set of allowed values.
// Supports: * , - and */step, plus a-b/step.
func parseCronField(field string, min, max int) ([]int, error) {
	field = strings.TrimSpace(field)
	if field == "" {
		return nil, errors.New("empty field")
	}
	set := make(map[int]struct{})
	for _, part := range strings.Split(field, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		step := 1
		if idx := strings.Index(part, "/"); idx >= 0 {
			stepStr := part[idx+1:]
			if _, err := fmt.Sscanf(stepStr, "%d", &step); err != nil || step < 1 {
				return nil, fmt.Errorf("invalid step %q", stepStr)
			}
			part = part[:idx]
		}
		lo, hi := min, max
		switch {
		case part == "*" || part == "":
			lo, hi = min, max
		default:
			if dash := strings.Index(part, "-"); dash >= 0 {
				if _, err := fmt.Sscanf(part[:dash], "%d", &lo); err != nil {
					return nil, fmt.Errorf("invalid range start %q", part[:dash])
				}
				if _, err := fmt.Sscanf(part[dash+1:], "%d", &hi); err != nil {
					return nil, fmt.Errorf("invalid range end %q", part[dash+1:])
				}
			} else {
				if _, err := fmt.Sscanf(part, "%d", &lo); err != nil {
					return nil, fmt.Errorf("invalid value %q", part)
				}
				hi = lo
			}
		}
		if lo < min || hi > max {
			return nil, fmt.Errorf("value out of range [%d,%d]", min, max)
		}
		for v := lo; v <= hi; v += step {
			set[v] = struct{}{}
		}
	}
	out := make([]int, 0, len(set))
	for v := range set {
		out = append(out, v)
	}
	// Sort for deterministic iteration.
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j] < out[j-1]; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	if len(out) == 0 {
		return nil, errors.New("no values matched")
	}
	return out, nil
}

// calculateNextRun returns the next time at or after `from` that satisfies the
// cron expression. Returns an error if the expression is invalid. The scan is
// capped at 5 years to guarantee termination for impossible expressions
// (e.g. "0 0 30 2 *").
func calculateNextRun(expr string, from time.Time) (time.Time, error) {
	s, err := parseCron(expr)
	if err != nil {
		return time.Time{}, err
	}
	// Start at the top of the next minute after `from`.
	t := from.Truncate(time.Minute).Add(time.Minute)
	deadline := from.AddDate(5, 0, 0)
	for t.Before(deadline) {
		if s.match(t) {
			return t, nil
		}
		t = t.Add(time.Minute)
	}
	return time.Time{}, errors.New("no matching time within 5 years")
}

// match reports whether a given time satisfies the cron schedule.
func (s *cronSchedule) match(t time.Time) bool {
	return containsInt(s.minute, t.Minute()) &&
		containsInt(s.hour, t.Hour()) &&
		containsInt(s.dom, t.Day()) &&
		containsInt(s.month, int(t.Month())) &&
		// cron treats Sunday as 0; Go's Weekday() does too.
		containsInt(s.dow, int(t.Weekday()))
}

// containsInt reports whether a sorted slice contains v.
func containsInt(list []int, v int) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

// ======== Stub executor ========
//
// stubAutopilotExecutor simulates autopilot task bodies without dispatching a
// real agent task. Each task_type produces a realistic output payload and a
// short log trail so the pipeline is observable end-to-end without LLM or
// agent dependencies (mirroring stubCICDProvider / stubSDLCStageExecutor).

type stubAutopilotExecutor struct{}

func (e *stubAutopilotExecutor) Execute(task *model.AutopilotTask, exec *model.AutopilotExecution) (map[string]interface{}, []string, error) {
	ts := time.Now().Format("15:04:05")
	logs := []string{
		fmt.Sprintf("[%s] task %d (%s) started via %s", ts, task.ID, task.TaskType, exec.TriggerType),
	}
	// Brief delay so SSE progress is observable; keeps tests fast.
	time.Sleep(500 * time.Millisecond)

	out := stubAutopilotTaskOutput(task)
	ts2 := time.Now().Format("15:04:05")
	logs = append(logs, fmt.Sprintf("[%s] produced: %s", ts2, summarizeAutopilotOutput(out)))
	logs = append(logs, fmt.Sprintf("[%s] task %d completed", ts2, task.ID))
	return out, logs, nil
}

// stubAutopilotTaskOutput returns a deterministic output payload keyed off
// the task_type so different autopilot tasks produce distinguishable results.
func stubAutopilotTaskOutput(task *model.AutopilotTask) map[string]interface{} {
	name := task.Name
	if name == "" {
		name = "未命名任务"
	}
	switch task.TaskType {
	case "report", "build_report":
		return map[string]interface{}{
			"report_url": fmt.Sprintf("/reports/autopilot-%d.html", task.ID),
			"summary":    fmt.Sprintf("自动生成报告：%s（共 12 项指标）", name),
			"metrics":    map[string]interface{}{"issues": 12, "resolved": 8},
		}
	case "scan", "security_scan":
		return map[string]interface{}{
			"scan_id":    fmt.Sprintf("scan-%d", task.ID),
			"findings":   3,
			"severity":   "medium",
			"scanned_at": time.Now().Format(time.RFC3339),
		}
	case "sync", "sync_issues":
		return map[string]interface{}{
			"synced":     42,
			"created":    5,
			"updated":    37,
			"source":     "external-tracker",
		}
	case "backup":
		return map[string]interface{}{
			"backup_path": fmt.Sprintf("/backups/autopilot-%d.tar.gz", task.ID),
			"bytes":       1048576,
		}
	case "notify", "notification":
		return map[string]interface{}{
			"notified":  8,
			"channels":  []string{"email", "slack"},
			"message":   fmt.Sprintf("通知：%s", name),
		}
	default:
		return map[string]interface{}{
			"task_type": task.TaskType,
			"message":   fmt.Sprintf("任务 %s 已执行完成", name),
			"task_id":   task.ID,
		}
	}
}

// summarizeAutopilotOutput returns a short digest of an executor output for logs.
func summarizeAutopilotOutput(out map[string]interface{}) string {
	if len(out) == 0 {
		return "(empty)"
	}
	keys := make([]string, 0, len(out))
	for k := range out {
		keys = append(keys, k)
	}
	return strings.Join(keys, ", ")
}

// Compile-time interface check.
var _ AutopilotExecutor = (*stubAutopilotExecutor)(nil)
