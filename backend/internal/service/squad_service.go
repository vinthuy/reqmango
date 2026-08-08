package service

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// AgentExecutorInterface defines the interface for agent execution.
// Used to avoid circular dependency between internal/service and internal/ai/service.
type AgentExecutorInterface interface {
	DispatchAgent(agentID uint64, userID uint64, task string, ctx *AgentDispatchContext) (*AgentDispatchResult, error)
}

// AgentDispatchContext is a simplified context struct for agent dispatch.
type AgentDispatchContext struct {
	IssueID     *uint64
	ProjectID   *uint64
	WorkspaceID uint64
	TriggeredBy string
}

// AgentDispatchResult is a simplified result struct for agent dispatch.
type AgentDispatchResult struct {
	ResultSummary string
}

type SquadService struct {
	db         *gorm.DB
	agentSvc   AgentExecutorInterface
	cancelStore sync.Map // executionID -> context.CancelFunc
}

func NewSquadService(db *gorm.DB) *SquadService {
	return &SquadService{db: db}
}

// SetAgentExecutor sets the agent executor for squad execution.
func (s *SquadService) SetAgentExecutor(agentSvc AgentExecutorInterface) {
	s.agentSvc = agentSvc
}

func (s *SquadService) Create(wid uint64, req request.SquadCreate) (*response.SquadResponse, error) {
	squad := &model.Squad{
		WorkspaceID:   wid,
		Name:          req.Name,
		Description:   req.Description,
		LeaderAgentID: req.LeaderAgentID,
		Goal:          req.Goal,
		Status:        "active",
	}
	if req.Config != nil {
		b, _ := json.Marshal(req.Config)
		squad.Config = model.FromRawMessage(b)
	}

	if err := s.db.Create(squad).Error; err != nil {
		return nil, err
	}

	// Add members
	for _, m := range req.Members {
		member := &model.SquadMember{
			SquadID:       squad.ID,
			AgentID:       m.AgentID,
			Role:          m.Role,
			AgentConfigID: m.AgentConfigID,
			Status:        "active",
			AssignedAt:    time.Now(),
		}
		if err := s.db.Create(member).Error; err != nil {
			return nil, err
		}
		squad.Members = append(squad.Members, *member)
	}

	return s.buildResponse(squad), nil
}

func (s *SquadService) Get(id uint64) (*response.SquadResponse, error) {
	var squad model.Squad
	if err := s.db.Preload("Members").First(&squad, id).Error; err != nil {
		return nil, err
	}
	return s.buildResponse(&squad), nil
}

func (s *SquadService) List(wid uint64, projectID *uint64) ([]*response.SquadResponse, error) {
	var squads []model.Squad
	q := s.db.Where("workspace_id = ?", wid)
	if projectID != nil {
		q = q.Where("project_id = ?", projectID)
	}
	if err := q.Preload("Members").Find(&squads).Error; err != nil {
		return nil, err
	}
	var resp []*response.SquadResponse
	for _, sq := range squads {
		resp = append(resp, s.buildResponse(&sq))
	}
	return resp, nil
}

func (s *SquadService) Update(id uint64, req request.SquadUpdate) (*response.SquadResponse, error) {
	var squad model.Squad
	if err := s.db.First(&squad, id).Error; err != nil {
		return nil, err
	}

	if req.Name != nil {
		squad.Name = *req.Name
	}
	if req.Description != nil {
		squad.Description = *req.Description
	}
	if req.LeaderAgentID != nil {
		squad.LeaderAgentID = req.LeaderAgentID
	}
	if req.Goal != nil {
		squad.Goal = *req.Goal
	}
	if req.Config != nil {
		b, _ := json.Marshal(req.Config)
		squad.Config = model.FromRawMessage(b)
	}

	if err := s.db.Save(&squad).Error; err != nil {
		return nil, err
	}

	return s.buildResponse(&squad), nil
}

func (s *SquadService) Delete(id uint64) error {
	return s.db.Delete(&model.Squad{}, id).Error
}

func (s *SquadService) AddMember(squadID uint64, req request.SquadMemberAdd) (*response.SquadMemberResponse, error) {
	var squad model.Squad
	if err := s.db.First(&squad, squadID).Error; err != nil {
		return nil, err
	}

	member := &model.SquadMember{
		SquadID:       squadID,
		AgentID:       req.AgentID,
		Role:          req.Role,
		AgentConfigID: req.AgentConfigID,
		Status:        "active",
		AssignedAt:    time.Now(),
	}

	if err := s.db.Create(member).Error; err != nil {
		return nil, err
	}

	return s.buildMemberResponse(member), nil
}

func (s *SquadService) RemoveMember(squadID, memberID uint64) error {
	return s.db.Model(&model.SquadMember{}).Where("squad_id = ? AND id = ?", squadID, memberID).Update("status", "removed").Error
}

// subtaskSpec is a single decomposed unit of work produced by the leader
// (or the fallback synthesizer when no leader is available).
type subtaskSpec struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Role        string `json:"role"` // "contributor" | "reviewer" | "leader"
}

// checkPermissions verifies the user is an active workspace member.
func (s *SquadService) checkPermissions(workspaceID, userID uint64) error {
	if s.db == nil {
		return nil // test mode: skip
	}
	var m model.WorkspaceMember
	if err := s.db.Where("workspace_id = ? AND user_id = ? AND is_active = ?",
		workspaceID, userID, true).First(&m).Error; err != nil {
		return common.Forbidden("Workspace member required")
	}
	return nil
}

// StartExecution creates a pending execution and launches the async pipeline.
func (s *SquadService) StartExecution(squadID uint64, req request.SquadExecutionStart) (*response.SquadExecutionResponse, error) {
	var squad model.Squad
	if err := s.db.Preload("Members").First(&squad, squadID).Error; err != nil {
		return nil, common.NotFound("Squad not found")
	}

	// Permission check
	if err := s.checkPermissions(squad.WorkspaceID, req.UserID); err != nil {
		return nil, err
	}

	exec := &model.SquadExecution{
		SquadID:   squadID,
		Status:    "pending",
		Goal:      req.Goal,
		StartedAt: &time.Time{},
	}
	*exec.StartedAt = time.Now()

	if req.InputData != nil {
		exec.InputData, _ = json.Marshal(req.InputData)
	}

	if err := s.db.Create(exec).Error; err != nil {
		return nil, err
	}

	// Launch async goroutine
	go s.executeAsync(exec.ID, squad, req.UserID)

	// Broadcast execution started
	s.broadcastEvent("squad.execution.started", map[string]interface{}{
		"execution_id": exec.ID,
		"squad_id":     squadID,
		"goal":         req.Goal,
	})

	return s.buildExecutionResponse(exec), nil
}

// executeAsync runs the 4-phase squad pipeline in a background goroutine.
func (s *SquadService) executeAsync(executionID uint64, squad model.Squad, userID uint64) {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelStore.Store(executionID, cancel)
	defer s.cancelStore.Delete(executionID)

	// Panic recovery
	defer func() {
		if r := recover(); r != nil {
			s.failExecution(executionID, fmt.Sprint(r))
		}
	}()

	// Load execution and set to running
	var exec model.SquadExecution
	if err := s.db.First(&exec, executionID).Error; err != nil {
		return
	}
	exec.Status = "running"
	s.db.Save(&exec)

	// Timeout from Squad.Config
	timeout := 300 * time.Second
	if squad.Config != nil {
		var cfg struct {
			TimeoutSeconds int `json:"timeout_seconds"`
		}
		json.Unmarshal(squad.Config.ToRawMessage(), &cfg)
		if cfg.TimeoutSeconds > 0 {
			timeout = time.Duration(cfg.TimeoutSeconds) * time.Second
		}
	}
	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()

	logs := []string{}
	outputs := []string{}
	failedCount := 0
	totalTasks := 0

	logs = append(logs, fmt.Sprintf("[%s] Squad execution started: %s", time.Now().Format("15:04:05"), exec.Goal))

	// ===== Phase 1: Decompose =====
	s.broadcastPhaseStart(executionID, "decompose")
	subtasks := s.decomposeGoal(&squad, request.SquadExecutionStart{Goal: exec.Goal, UserID: userID}, &logs)

	if len(subtasks) == 0 {
		for _, m := range squad.Members {
			if m.Role == "observer" || m.Role == "leader" {
				continue
			}
			subtasks = append(subtasks, subtaskSpec{
				Title:       truncateStr(exec.Goal, 80),
				Description: exec.Goal,
				Role:        m.Role,
			})
		}
	}
	if len(subtasks) == 0 {
		logs = append(logs, fmt.Sprintf("[%s] No executable members; aborting", time.Now().Format("15:04:05")))
	} else {
		logs = append(logs, fmt.Sprintf("[%s] Decomposed into %d subtask(s)", time.Now().Format("15:04:05"), len(subtasks)))
	}

	// ===== Phase 2: Execute (with context check + retry) =====
	s.broadcastPhaseStart(executionID, "execute")
	var contributorOutputs []string
	for _, st := range subtasks {
		if ctx.Err() != nil {
			logs = append(logs, fmt.Sprintf("[%s] Execution cancelled during execute phase", time.Now().Format("15:04:05")))
			break
		}
		if st.Role == "reviewer" || st.Role == "leader" || st.Role == "observer" {
			continue
		}
		member := s.findMemberByRole(squad.Members, st.Role)
		if member == nil {
			logs = append(logs, fmt.Sprintf("[%s] No member for role %q; skipping", time.Now().Format("15:04:05"), st.Role))
			continue
		}
		taskDesc := st.Description
		if len(contributorOutputs) > 0 {
			taskDesc = fmt.Sprintf("%s\n\n以下是上游成员已完成的工作：\n%s", st.Description, strings.Join(contributorOutputs, "\n---\n"))
		}

		totalTasks++
		result := s.executeSubtaskWithRetry(ctx, executionID, &squad, member, taskDesc, st.Title, userID, &logs)
		if result != "" {
			contributorOutputs = append(contributorOutputs, fmt.Sprintf("[%s] %s", st.Title, result))
			outputs = append(outputs, result)
		} else {
			failedCount++
		}
	}

	// ===== Phase 3: Review =====
	s.broadcastPhaseStart(executionID, "review")
	if len(contributorOutputs) > 0 && ctx.Err() == nil {
		for _, member := range squad.Members {
			if member.Role != "reviewer" || ctx.Err() != nil {
				continue
			}
			reviewDesc := fmt.Sprintf("你是审核者。请审核以下成员产出并给出结论：\n\n%s", strings.Join(contributorOutputs, "\n---\n"))
			totalTasks++
			result := s.runMemberTask(&squad, &member, reviewDesc, "审核成员产出", userID, executionID, &logs)
			if result != "" {
				outputs = append(outputs, fmt.Sprintf("[审核反馈] %s", result))
			} else {
				failedCount++
			}
		}
	}

	// ===== Phase 4: Aggregate =====
	s.broadcastPhaseStart(executionID, "aggregate")
	switch {
	case totalTasks > 0 && failedCount >= totalTasks:
		exec.Status = "failed"
	case failedCount > 0:
		exec.Status = "partial_failed"
	default:
		exec.Status = "completed"
	}
	completedAt := time.Now()
	exec.CompletedAt = &completedAt
	logsJSON, _ := json.Marshal(logs)
	outputsJSON, _ := json.Marshal(outputs)
	exec.Logs = logsJSON
	exec.OutputData = outputsJSON
	s.db.Save(&exec)

	s.broadcastEvent("squad.execution.completed", map[string]interface{}{
		"execution_id": executionID,
		"status":       exec.Status,
	})
}

// failExecution marks an execution as failed (used by panic recovery and error paths).
func (s *SquadService) failExecution(executionID uint64, errMsg string) {
	var exec model.SquadExecution
	if err := s.db.First(&exec, executionID).Error; err != nil {
		return
	}
	exec.Status = "failed"
	exec.ErrorInfo = errMsg
	completedAt := time.Now()
	exec.CompletedAt = &completedAt
	s.db.Save(&exec)
	s.broadcastEvent("squad.execution.completed", map[string]interface{}{
		"execution_id": executionID,
		"status":       "failed",
	})
}

// CancelExecution cancels a running execution via context cancellation.
func (s *SquadService) CancelExecution(executionID uint64) error {
	val, ok := s.cancelStore.Load(executionID)
	if !ok {
		return common.NotFound("Execution not running or not found")
	}
	cancelFunc := val.(context.CancelFunc)
	cancelFunc()
	s.cancelStore.Delete(executionID)

	// Update execution status
	var exec model.SquadExecution
	if err := s.db.First(&exec, executionID).Error; err != nil {
		return common.NotFound("Execution not found")
	}
	exec.Status = "cancelled"
	exec.CancelReason = "User cancelled"
	now := time.Now()
	exec.CancelledAt = &now
	exec.CompletedAt = &now
	s.db.Save(&exec)

	s.broadcastEvent("squad.execution.cancelled", map[string]interface{}{
		"execution_id": executionID,
		"reason":       "User cancelled",
	})

	return nil
}

// broadcastEvent sends an SSE event to all connected clients.
func (s *SquadService) broadcastEvent(event string, data interface{}) {
	SSE.BroadcastEvent(event, data)
}

// broadcastPhaseStart broadcasts a phase_start event for the given phase.
func (s *SquadService) broadcastPhaseStart(executionID uint64, phase string) {
	s.broadcastEvent("squad.execution.phase_start", map[string]interface{}{
		"execution_id": executionID,
		"phase":        phase,
	})
}

// executeSubtaskWithRetry dispatches a subtask with automatic retry on failure.
func (s *SquadService) executeSubtaskWithRetry(ctx context.Context, executionID uint64, squad *model.Squad, member *model.SquadMember, taskDesc, title string, userID uint64, logs *[]string) string {
	maxRetries := 2
	if squad.Config != nil {
		var cfg struct {
			MaxRetries int `json:"max_retries"`
		}
		json.Unmarshal(squad.Config.ToRawMessage(), &cfg)
		if cfg.MaxRetries > 0 {
			maxRetries = cfg.MaxRetries
		}
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if ctx.Err() != nil {
			*logs = append(*logs, fmt.Sprintf("[%s] Execution cancelled, stopping retry for agent %d",
				time.Now().Format("15:04:05"), member.AgentID))
			return ""
		}

		if attempt > 0 {
			*logs = append(*logs, fmt.Sprintf("[%s] Retrying agent %d (attempt %d/%d)",
				time.Now().Format("15:04:05"), member.AgentID, attempt+1, maxRetries+1))
			s.broadcastEvent("squad.execution.subtask_progress", map[string]interface{}{
				"execution_id": executionID,
				"retry":        attempt,
			})
		}

		result := s.runMemberTask(squad, member, taskDesc, title, userID, executionID, logs)
		if result != "" {
			return result
		}
		// Update retry_count and error_message on the last failed task
		if s.db != nil {
			var lastTask model.SquadTask
			if tx := s.db.Where("squad_id = ? AND member_id = ? AND status = ?",
				squad.ID, member.ID, "failed").Order("id DESC").First(&lastTask); tx.Error == nil {
				lastTask.RetryCount = attempt + 1
				lastTask.ErrorMessage = "task returned empty result"
				s.db.Save(&lastTask)
			}
		}
	}

	*logs = append(*logs, fmt.Sprintf("[%s] Agent %d failed after %d retries",
		time.Now().Format("15:04:05"), member.AgentID, maxRetries))
	return ""
}

// decomposeGoal dispatches the goal to the squad's leader agent and parses
// the response into subtasks. Returns an empty slice when no leader is
// configured or decomposition fails (caller must fall back).
func (s *SquadService) decomposeGoal(squad *model.Squad, req request.SquadExecutionStart, logs *[]string) []subtaskSpec {
	if squad.LeaderAgentID == nil || s.agentSvc == nil {
		return nil
	}

	var memberDesc []string
	for _, m := range squad.Members {
		memberDesc = append(memberDesc, fmt.Sprintf("- agent_id=%d, role=%s", m.AgentID, m.Role))
	}
	leaderTask := fmt.Sprintf(
		"你是团队的 Leader。请将下面的目标拆分为可执行的子任务，并指派给团队成员。\n\n"+
			"目标：%s\n\n"+
			"团队成员：\n%s\n\n"+
			"请仅返回一个 JSON 对象（不要附加任何说明文字），格式为：\n"+
			`{"subtasks":[{"title":"简短标题","description":"具体可执行的子任务描述","role":"contributor"}]}`+"\n"+
			"role 可选值：contributor, reviewer。reviewer 角色用于审核 contributor 的产出，请不要为 reviewer 拆分过多子任务（一个即可）。",
		req.Goal, strings.Join(memberDesc, "\n"),
	)

	result, err := s.agentSvc.DispatchAgent(*squad.LeaderAgentID, req.UserID, leaderTask, &AgentDispatchContext{
		WorkspaceID: squad.WorkspaceID,
		TriggeredBy: "squad_leader_decompose",
	})
	if err != nil {
		*logs = append(*logs, fmt.Sprintf("[%s] Leader decomposition failed: %v", time.Now().Format("15:04:05"), err))
		return nil
	}

	subtasks := parseSubtasksFromLLM(result.ResultSummary)
	if len(subtasks) == 0 {
		*logs = append(*logs, fmt.Sprintf("[%s] Leader returned no parseable subtasks; falling back", time.Now().Format("15:04:05")))
	} else {
		*logs = append(*logs, fmt.Sprintf("[%s] Leader decomposed goal into %d subtask(s)", time.Now().Format("15:04:05"), len(subtasks)))
	}
	return subtasks
}

// runMemberTask dispatches a single subtask to a member agent, records the
// SquadTask row, appends a log entry, and returns the result summary. Returns
// an empty string when the dispatch fails.
func (s *SquadService) runMemberTask(squad *model.Squad, member *model.SquadMember, taskDesc, title string, userID uint64, executionID uint64, logs *[]string) string {
	startedAt := time.Now()
	task := &model.SquadTask{
		SquadID:         squad.ID,
		MemberID:        member.ID,
		Status:          "running",
		Priority:        "medium",
		Progress:        0,
		TaskDescription: taskDesc,
		StartedAt:       &startedAt,
	}
	if err := s.db.Create(task).Error; err != nil {
		*logs = append(*logs, fmt.Sprintf("[%s] Failed to create SquadTask for agent %d: %v", time.Now().Format("15:04:05"), member.AgentID, err))
		return ""
	}

	*logs = append(*logs, fmt.Sprintf("[%s] Agent %d (%s) started: %s",
		time.Now().Format("15:04:05"), member.AgentID, member.Role, title))

	// Broadcast subtask_start
	s.broadcastEvent("squad.execution.subtask_start", map[string]interface{}{
		"execution_id": executionID,
		"task_id":      task.ID,
		"member_id":    member.ID,
		"title":        title,
	})

	var result string
	if s.agentSvc != nil {
		dispatchResult, err := s.agentSvc.DispatchAgent(member.AgentID, userID, taskDesc, &AgentDispatchContext{
			WorkspaceID: squad.WorkspaceID,
			TriggeredBy: "squad",
		})
		if err != nil {
			*logs = append(*logs, fmt.Sprintf("[%s] Agent %d (%s) failed: %v",
				time.Now().Format("15:04:05"), member.AgentID, member.Role, err))
			task.Status = "failed"
			task.Feedback = err.Error()
			task.ErrorMessage = err.Error()
		} else {
			result = dispatchResult.ResultSummary
			*logs = append(*logs, fmt.Sprintf("[%s] Agent %d (%s) completed",
				time.Now().Format("15:04:05"), member.AgentID, member.Role))
			task.Status = "completed"
			task.Progress = 100
			task.Feedback = result
		}
	} else {
		// Fallback when no agent service is wired (e.g. in tests).
		result = fmt.Sprintf("Agent %d (%s) would execute: %s", member.AgentID, member.Role, title)
		*logs = append(*logs, fmt.Sprintf("[%s] Agent %d (%s) executed (stub)",
			time.Now().Format("15:04:05"), member.AgentID, member.Role))
		task.Status = "completed"
		task.Progress = 100
		task.Feedback = result
	}

	completedAt := time.Now()
	task.CompletedAt = &completedAt
	if err := s.db.Save(task).Error; err != nil {
		*logs = append(*logs, fmt.Sprintf("[%s] Failed to save SquadTask for agent %d: %v",
			time.Now().Format("15:04:05"), member.AgentID, err))
	}

	// Broadcast subtask_done
	s.broadcastEvent("squad.execution.subtask_done", map[string]interface{}{
		"execution_id": executionID,
		"task_id":      task.ID,
		"status":       task.Status,
	})

	return result
}

// findMemberByRole returns the first active member matching the given role.
func (s *SquadService) findMemberByRole(members []model.SquadMember, role string) *model.SquadMember {
	// Normalize role for matching.
	want := strings.ToLower(strings.TrimSpace(role))
	if want == "" || want == "member" {
		want = "contributor"
	}
	for i := range members {
		m := &members[i]
		if m.Status != "active" {
			continue
		}
		got := strings.ToLower(strings.TrimSpace(m.Role))
		if got == "member" {
			got = "contributor"
		}
		if got == want {
			return m
		}
	}
	return nil
}

// parseSubtasksFromLLM extracts subtasks from the leader's free-form response.
// Handles JSON wrapped in markdown code fences and tolerates trailing prose.
func parseSubtasksFromLLM(response string) []subtaskSpec {
	if response == "" {
		return nil
	}

	// Strip markdown code fences if present.
	trimmed := strings.TrimSpace(response)
	fenceRe := regexp.MustCompile("(?s)```(?:json)?\\s*(.*?)```")
	if matches := fenceRe.FindStringSubmatch(trimmed); len(matches) >= 2 {
		trimmed = strings.TrimSpace(matches[1])
	}

	// Find the outermost JSON object.
	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start < 0 || end <= start {
		return nil
	}
	jsonStr := trimmed[start : end+1]

	var parsed struct {
		Subtasks []subtaskSpec `json:"subtasks"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return nil
	}

	// Filter out subtasks with empty descriptions and normalize role.
	out := make([]subtaskSpec, 0, len(parsed.Subtasks))
	for _, st := range parsed.Subtasks {
		if strings.TrimSpace(st.Description) == "" {
			continue
		}
		st.Role = strings.ToLower(strings.TrimSpace(st.Role))
		if st.Role == "" || st.Role == "member" {
			st.Role = "contributor"
		}
		if st.Title == "" {
			st.Title = truncateStr(st.Description, 80)
		}
		out = append(out, st)
	}
	return out
}

// GetExecution returns a single execution by ID.
func (s *SquadService) GetExecution(id uint64) (*response.SquadExecutionResponse, error) {
	var exec model.SquadExecution
	if err := s.db.First(&exec, id).Error; err != nil {
		return nil, common.NotFound("Execution not found")
	}
	return s.buildExecutionResponse(&exec), nil
}

// ListExecutions returns all executions for a given squad, ordered by most recent.
func (s *SquadService) ListExecutions(squadID uint64) ([]*response.SquadExecutionResponse, error) {
	var execs []model.SquadExecution
	if err := s.db.Where("squad_id = ?", squadID).Order("id DESC").Find(&execs).Error; err != nil {
		return nil, err
	}
	var resp []*response.SquadExecutionResponse
	for i := range execs {
		resp = append(resp, s.buildExecutionResponse(&execs[i]))
	}
	return resp, nil
}

func (s *SquadService) buildResponse(squad *model.Squad) *response.SquadResponse {
	resp := &response.SquadResponse{
		ID:            squad.ID,
		WorkspaceID:   squad.WorkspaceID,
		ProjectID:     squad.ProjectID,
		Name:          squad.Name,
		Description:   squad.Description,
		LeaderAgentID: squad.LeaderAgentID,
		Status:        squad.Status,
		Goal:          squad.Goal,
		Config:        json.RawMessage(squad.Config),
		Members:       make([]response.SquadMemberResponse, 0, len(squad.Members)),
		CreatedAt:     squad.CreatedAt,
		UpdatedAt:     squad.UpdatedAt,
	}
	for _, m := range squad.Members {
		resp.Members = append(resp.Members, response.SquadMemberResponse{
			ID:            m.ID,
			SquadID:       m.SquadID,
			AgentID:       m.AgentID,
			Role:          m.Role,
			AgentConfigID: m.AgentConfigID,
			Status:        m.Status,
			AssignedAt:    m.AssignedAt,
			RemovedAt:     m.RemovedAt,
		})
	}
	return resp
}

func (s *SquadService) buildMemberResponse(member *model.SquadMember) *response.SquadMemberResponse {
	return &response.SquadMemberResponse{
		ID:            member.ID,
		SquadID:       member.SquadID,
		AgentID:       member.AgentID,
		Role:          member.Role,
		AgentConfigID: member.AgentConfigID,
		Status:        member.Status,
		AssignedAt:    member.AssignedAt,
		RemovedAt:     member.RemovedAt,
	}
}

func (s *SquadService) buildExecutionResponse(exec *model.SquadExecution) *response.SquadExecutionResponse {
	return &response.SquadExecutionResponse{
		ID:           exec.ID,
		SquadID:      exec.SquadID,
		Status:       exec.Status,
		Goal:         exec.Goal,
		InputData:    exec.InputData,
		OutputData:   exec.OutputData,
		Logs:         exec.Logs,
		StartedAt:    exec.StartedAt,
		CompletedAt:  exec.CompletedAt,
		FailedAt:     exec.FailedAt,
		ErrorInfo:    exec.ErrorInfo,
		CancelledAt:  exec.CancelledAt,
		CancelReason: exec.CancelReason,
		CreatedAt:    exec.CreatedAt,
	}
}

// truncateStr returns at most n characters of s, suffixed with "..." when truncated.
func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}