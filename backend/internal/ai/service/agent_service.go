package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/reqmango/backend/internal/ai/common"
	"github.com/reqmango/backend/internal/ai/llm"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// AgentService manages AI agents and their actions.
type AgentService struct {
	db      *gorm.DB
	llm     *llm.LLMClient
	aiSvc   *AIService
	memSvc  MemoryServiceInterface
}

// NewAgentService creates a new AgentService.
func NewAgentService(db *gorm.DB, llmClient *llm.LLMClient, aiSvc *AIService) *AgentService {
	return &AgentService{db: db, llm: llmClient, aiSvc: aiSvc}
}

// SetMemoryService sets the memory service
func (s *AgentService) SetMemoryService(memSvc MemoryServiceInterface) {
	s.memSvc = memSvc
}

// ======== CRUD ========

// ListByWorkspace returns all agents in a workspace.
func (s *AgentService) ListByWorkspace(workspaceID uint64) ([]model.Agent, error) {
	var agents []model.Agent
	if err := s.db.Where("workspace_id = ?", workspaceID).Order("name ASC").Find(&agents).Error; err != nil {
		return nil, common.Internal("Failed to list agents")
	}
	return agents, nil
}

// ListVisibleAgents returns agents visible to a specific user (workspace agents + private agents owned by user).
func (s *AgentService) ListVisibleAgents(workspaceID, userID uint64) ([]model.Agent, error) {
	var agents []model.Agent
	if err := s.db.Where("workspace_id = ? AND (visibility = ? OR (visibility = ? AND created_by_id = ?))",
		workspaceID, model.VisibilityWorkspace, model.VisibilityPrivate, userID).
		Order("name ASC").Find(&agents).Error; err != nil {
		return nil, common.Internal("Failed to list visible agents")
	}
	return agents, nil
}

// CanInvoke checks if a user can invoke an agent based on permission_mode and invocation_targets.
func (s *AgentService) CanInvoke(agentID, userID uint64) (bool, error) {
	agent, err := s.GetByID(agentID)
	if err != nil {
		return false, err
	}

	// Owner can always invoke
	if agent.CreatedByID != nil && *agent.CreatedByID == userID {
		return true, nil
	}

	// Check permission mode
	switch agent.PermissionMode {
	case model.PermissionModePrivate:
		// Private mode: only owner can invoke
		return false, nil
	case model.PermissionModePublicTo:
		// Public_to mode: check invocation targets
		var targets []model.AgentInvocationTarget
		if err := json.Unmarshal(agent.InvocationTargets, &targets); err != nil {
			return false, nil
		}
		for _, target := range targets {
			switch target.TargetType {
			case "workspace":
				// Workspace target: any workspace member can invoke
				return true, nil
			case "member":
				// Member target: check if user matches
				if target.TargetID != "" {
					// Convert targetID to uint64 and compare
					targetUserID, err := strconv.ParseUint(target.TargetID, 10, 64)
					if err == nil && targetUserID == userID {
						return true, nil
					}
				}
			case "team":
				// Team target: reserved, INERT in v1
				continue
			}
		}
		return false, nil
	default:
		return false, nil
	}
}

// GetByID returns a single agent by ID.
func (s *AgentService) GetByID(agentID uint64) (*model.Agent, error) {
	var agent model.Agent
	if err := s.db.First(&agent, agentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.AgentNotFound()
		}
		return nil, common.Internal("Failed to get agent")
	}
	return &agent, nil
}

type AgentCreateRequest struct {
	Name               string                  `json:"name" binding:"required"`
	Avatar             string                  `json:"avatar"`
	AgentType          string                  `json:"agent_type"`
	Capabilities       []string                `json:"capabilities"`
	Status             string                  `json:"status"`
	ModelOverride      *string                 `json:"model_override"`
	SystemPrompt       *string                 `json:"system_prompt"`
	
	// Permission fields
	PermissionMode     *string                 `json:"permission_mode"`
	Visibility         *string                 `json:"visibility"`
	InvocationTargets  []model.AgentInvocationTarget `json:"invocation_targets"`
}

type AgentUpdateRequest struct {
	Name               *string                 `json:"name"`
	Avatar             *string                 `json:"avatar"`
	AgentType          *string                 `json:"agent_type"`
	Capabilities       *[]string               `json:"capabilities"`
	Status             *string                 `json:"status"`
	ModelOverride      *string                 `json:"model_override"`
	SystemPrompt       *string                 `json:"system_prompt"`
	
	// Permission fields (owner-only)
	PermissionMode     *string                 `json:"permission_mode"`
	Visibility         *string                 `json:"visibility"`
	InvocationTargets  *[]model.AgentInvocationTarget `json:"invocation_targets"`
}

// Create creates a new agent.
func (s *AgentService) Create(workspaceID, userID uint64, req *AgentCreateRequest) (*model.Agent, error) {
	if req.Avatar == "" {
		req.Avatar = "🤖"
	}
	if req.AgentType == "" {
		req.AgentType = "builtin"
	}
	if req.Status == "" {
		req.Status = "active"
	}

	capsJSON, _ := json.Marshal(req.Capabilities)

	// Permission mode: use requested or default to private
	permissionMode := model.PermissionModePrivate
	if req.PermissionMode != nil {
		permissionMode = model.AgentPermissionMode(*req.PermissionMode)
	}

	// Visibility: derive from permission mode if not explicitly set
	visibility := model.VisibilityPrivate
	if req.Visibility != nil {
		visibility = model.AgentVisibility(*req.Visibility)
	} else {
		// Derive visibility from permission mode
		if permissionMode == model.PermissionModePublicTo && len(req.InvocationTargets) > 0 {
			for _, target := range req.InvocationTargets {
				if target.TargetType == "workspace" {
					visibility = model.VisibilityWorkspace
					break
				}
			}
		}
	}

	// Invocation targets
	targetsJSON, _ := json.Marshal(req.InvocationTargets)

	agent := model.Agent{
		WorkspaceID:       workspaceID,
		Name:              req.Name,
		Avatar:            req.Avatar,
		AgentType:         req.AgentType,
		Capabilities:      capsJSON,
		Status:            req.Status,
		ModelOverride:     req.ModelOverride,
		SystemPrompt:      req.SystemPrompt,
		PermissionMode:    permissionMode,
		InvocationTargets: targetsJSON,
		Visibility:        visibility,
	}
	agent.CreatedByID = &userID

	if err := s.db.Create(&agent).Error; err != nil {
		return nil, common.Internal("Failed to create agent")
	}
	return &agent, nil
}

// Update modifies an existing agent.
func (s *AgentService) Update(agentID, userID uint64, req *AgentUpdateRequest) (*model.Agent, error) {
	agent, err := s.GetByID(agentID)
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{"updated_by_id": userID}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Avatar != nil {
		updates["avatar"] = *req.Avatar
	}
	if req.AgentType != nil {
		updates["agent_type"] = *req.AgentType
	}
	if req.Capabilities != nil {
		capsJSON, _ := json.Marshal(*req.Capabilities)
		updates["capabilities"] = capsJSON
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.ModelOverride != nil {
		updates["model_override"] = *req.ModelOverride
	}
	if req.SystemPrompt != nil {
		updates["system_prompt"] = *req.SystemPrompt
	}

	// Permission updates (owner-only, silently ignored for non-owners)
	if agent.CreatedByID != nil && *agent.CreatedByID == userID {
		if req.PermissionMode != nil {
			updates["permission_mode"] = model.AgentPermissionMode(*req.PermissionMode)
		}
		if req.Visibility != nil {
			updates["visibility"] = model.AgentVisibility(*req.Visibility)
		}
		if req.InvocationTargets != nil {
			targetsJSON, _ := json.Marshal(*req.InvocationTargets)
			updates["invocation_targets"] = targetsJSON
		}
	}

	if err := s.db.Model(agent).Updates(updates).Error; err != nil {
		return nil, common.Internal("Failed to update agent")
	}
	return s.GetByID(agentID)
}

// Delete soft-deletes an agent.
func (s *AgentService) Delete(agentID uint64) error {
	agent, err := s.GetByID(agentID)
	if err != nil {
		return err
	}
	if err := s.db.Delete(agent).Error; err != nil {
		return common.Internal("Failed to delete agent")
	}
	return nil
}

// ======== Agent Activity ========

// GetActivity returns the activity log for an agent.
func (s *AgentService) GetActivity(agentID uint64) ([]model.AgentActivity, error) {
	var activities []model.AgentActivity
	if err := s.db.Where("agent_id = ?", agentID).
		Order("executed_at DESC").
		Limit(50).
		Find(&activities).Error; err != nil {
		return nil, common.Internal("Failed to list agent activities")
	}
	return activities, nil
}

// ListWorkspaceActivity returns activities for all agents in a workspace.
func (s *AgentService) ListWorkspaceActivity(workspaceID uint64, agentID *uint64, action string, limit int) ([]model.AgentActivity, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	var agentIDs []uint64
	if agentID != nil {
		agentIDs = []uint64{*agentID}
	} else {
		var agents []model.Agent
		if err := s.db.Model(&model.Agent{}).Where("workspace_id = ?", workspaceID).Pluck("id", &agents).Error; err != nil {
			return nil, common.Internal("Failed to list agents")
		}
		for _, a := range agents {
			agentIDs = append(agentIDs, a.ID)
		}
	}
	if len(agentIDs) == 0 {
		return []model.AgentActivity{}, nil
	}
	q := s.db.Where("agent_id IN ?", agentIDs)
	if action != "" {
		q = q.Where("action = ?", action)
	}
	var activities []model.AgentActivity
	if err := q.Order("executed_at DESC").Limit(limit).Find(&activities).Error; err != nil {
		return nil, common.Internal("Failed to list agent activities")
	}
	return activities, nil
}

// UpdateActivityFeedback records user feedback (rating) for an agent activity.
func (s *AgentService) UpdateActivityFeedback(activityID uint64, rating int) error {
	if rating != 1 && rating != -1 {
		return common.Validation("Rating must be 1 (positive) or -1 (negative)")
	}
	result := s.db.Model(&model.AgentActivity{}).Where("id = ?", activityID).Update("rating", rating)
	if result.Error != nil {
		return common.Internal("Failed to update feedback")
	}
	if result.RowsAffected == 0 {
		return common.NotFound("Activity not found")
	}
	return nil
}

func (s *AgentService) recordActivity(agent *model.Agent, issueID *uint64, action, summary, taskCtx string, userID uint64) {
	activity := model.AgentActivity{
		AgentID:       agent.ID,
		IssueID:       issueID,
		Action:        action,
		ResultSummary: summary,
		AgentName:     agent.Name,
		TaskContext:   &taskCtx,
	}
	activity.CreatedByID = &userID
	s.db.Create(&activity)
}

// ======== AI Dispatch ========

// DispatchContext holds context for an agent dispatch.
type DispatchContext struct {
	IssueID     *uint64 `json:"issue_id"`
	ProjectID   *uint64 `json:"project_id"`
	WorkspaceID uint64  `json:"workspace_id"`
	TriggeredBy string  `json:"triggered_by"`
}

// DispatchAgent sends a task to an agent and executes any tool calls it makes.
func (s *AgentService) DispatchAgent(agentID, userID uint64, task string, ctx *DispatchContext) (*model.AgentActivity, error) {
	agent, err := s.GetByID(agentID)
	if err != nil {
		return nil, err
	}

	if agent.Status != "active" {
		return nil, common.Validation("Agent is not active")
	}

	// Check permission to invoke
	canInvoke, err := s.CanInvoke(agentID, userID)
	if err != nil {
		return nil, err
	}
	if !canInvoke {
		return nil, common.Permission("You do not have permission to invoke this agent")
	}

	systemPrompt := s.buildAgentSystemPrompt(agent, ctx)
	tools := s.filterToolsByCapabilities(agent)

	actx := &AIContext{
		WorkspaceID: ctx.WorkspaceID,
		ProjectID:   0,
		Mode:        "agent",
		UserID:      userID,
	}
	if ctx.ProjectID != nil {
		actx.ProjectID = *ctx.ProjectID
	}

	// Retrieve relevant memories and inject into system prompt
	if s.memSvc != nil {
		memories, _ := s.retrieveAgentMemories(context.Background(), agent, actx)
		if len(memories) > 0 {
			var memBuilder strings.Builder
			memBuilder.WriteString("\n\n以下是相关历史记忆（帮助你理解上下文）：\n")
			for _, mem := range memories {
				memBuilder.WriteString(fmt.Sprintf("- [%s] %s\n", mem.ContextName, mem.Content[:minX(len(mem.Content), 150)]))
			}
			systemPrompt += memBuilder.String()
		}
	}

	// Execute skill if agent template has skill integration enabled
	var skillResult string
	if agent.TemplateID != nil {
		skillResult = s.executeAgentSkill(context.Background(), agent, task, userID)
		if skillResult != "" {
			systemPrompt += fmt.Sprintf("\n\n技能执行结果（作为参考）：\n%s", skillResult)
		}
	}

	executedTools := make([]string, 0)
	resp, llmErr := s.llm.ChatSyncWithTools(context.Background(), systemPrompt, []llm.Message{
		{Role: "user", Content: task},
	}, tools, func(name string, input json.RawMessage) (string, error) {
		result, execErr := s.aiSvc.ExecuteTool(name, input, actx)
		if execErr != nil {
			return "", execErr
		}
		executedTools = append(executedTools, fmt.Sprintf("%s(%v)", name, input))
		b, _ := json.Marshal(result)
		return string(b), nil
	})
	if llmErr != nil {
		s.recordActivity(agent, ctx.IssueID, "dispatch",
			fmt.Sprintf("Failed: %v", llmErr), task, userID)
		return nil, common.Internal(fmt.Sprintf("Agent LLM call failed: %v", llmErr))
	}

	var resultBuilder strings.Builder
	resultBuilder.WriteString(fmt.Sprintf("Agent %s processed task: %s\n", agent.Name, task))

	if len(executedTools) > 0 {
		resultBuilder.WriteString(fmt.Sprintf("Executed %d tool(s):\n", len(executedTools)))
		for _, t := range executedTools {
			resultBuilder.WriteString(fmt.Sprintf("- %s\n", t))
		}
	}

	if resp.Content != "" {
		resultBuilder.WriteString(fmt.Sprintf("\nResponse:\n%s", resp.Content))
	}

	summary := resultBuilder.String()
	if len(summary) > 500 {
		summary = summary[:500] + "..."
	}

	s.recordActivity(agent, ctx.IssueID, "dispatch", summary, task, userID)

	// Record heartbeat to mark agent as online
	s.RecordHeartbeat(agent.ID)

	// Save task result as memory after completion
	if s.memSvc != nil && resp.Content != "" {
		s.saveAgentTaskMemory(context.Background(), agent, actx, task, resp.Content, executedTools)
	}

	return s.getLatestActivity(agent.ID)
}

// retrieveAgentMemories retrieves relevant memories for an agent task
func (s *AgentService) retrieveAgentMemories(ctx context.Context, agent *model.Agent, actx *AIContext) ([]*model.MemoryEntry, error) {
	filters := map[string]interface{}{
		"project_id":   actx.ProjectID,
		"agent_id":     agent.ID,
		"memory_type":  model.MemoryMediumTerm,
		"limit":        5,
	}
	return s.memSvc.ListMemories(ctx, actx.WorkspaceID, filters)
}

// saveAgentTaskMemory saves the agent task result as a memory entry
func (s *AgentService) saveAgentTaskMemory(ctx context.Context, agent *model.Agent, actx *AIContext, task, result string, executedTools []string) {
	contextKey := fmt.Sprintf("agent_%d_task", agent.ID)

	content := fmt.Sprintf("任务：%s\n\n执行结果：%s", task, result)
	if len(executedTools) > 0 {
		content += fmt.Sprintf("\n\n执行工具：%v", executedTools)
	}

	entry := &model.MemoryEntry{
		WorkspaceID:    actx.WorkspaceID,
		ProjectID:      &actx.ProjectID,
		AgentID:        &agent.ID,
		MemoryType:     model.MemoryMediumTerm,
		Scope:          model.ScopeAgent,
		Content:        content,
		ContextKey:     contextKey,
		ContextName:    agent.Name,
		RelevanceScore: 0.7,
	}

	go func() {
		s.memSvc.CreateMemory(ctx, entry)
	}()
}

// ======== Auto Triage ========

// AutoTriage analyzes an unsorted issue and suggests type, priority, and labels.
func (s *AgentService) AutoTriage(issueID, userID uint64) (*model.AgentActivity, error) {
	agent, err := s.findAgentByCapability("analyze")
	if err != nil {
		return nil, err
	}

	task := fmt.Sprintf(
		"Analyze the following issue and suggest the appropriate type, priority, labels, and assignee. "+
			"Use the project context to understand what issue types, states, labels, and members are available. "+
			"Then provide your triage recommendation. Issue ID: %d", issueID)
	ctx := &DispatchContext{
		IssueID:     &issueID,
		WorkspaceID: agent.WorkspaceID,
		TriggeredBy: "auto_triage",
	}

	return s.DispatchAgent(agent.ID, userID, task, ctx)
}

// ======== Auto Assign ========

// AutoAssign analyzes the issue and project workload to suggest the best assignee.
func (s *AgentService) AutoAssign(issueID, userID uint64) (*model.AgentActivity, error) {
	agent, err := s.findAgentByCapability("analyze")
	if err != nil {
		return nil, err
	}

	task := fmt.Sprintf(
		"Analyze issue %d and the project members' current workload. "+
			"List the available members and their open issue counts, then recommend the best assignee. "+
			"Consider the issue's type, priority, and required skills. Issue ID: %d", issueID, issueID)
	ctx := &DispatchContext{
		IssueID:     &issueID,
		WorkspaceID: agent.WorkspaceID,
		TriggeredBy: "auto_assign",
	}

	return s.DispatchAgent(agent.ID, userID, task, ctx)
}

// ======== Handle Mention ========

// HandleMention processes an @agent-name mention in a comment.
func (s *AgentService) HandleMention(agentID, commentID, userID uint64, commentBody, issueName string, issueID *uint64) (*model.AgentActivity, error) {
	agent, err := s.GetByID(agentID)
	if err != nil {
		return nil, err
	}

	task := fmt.Sprintf(
		"You were mentioned in a comment on issue '%s'. The comment says:\n\n%s\n\n"+
			"Please respond appropriately based on your capabilities. If you can help, analyze the situation and provide your response.", issueName, commentBody)

	ctx := &DispatchContext{
		IssueID:     issueID,
		WorkspaceID: agent.WorkspaceID,
		TriggeredBy: "mention",
	}

	return s.DispatchAgent(agent.ID, userID, task, ctx)
}

// ======== Summarize Cycle ========

// SummarizeCycle generates a natural-language summary of a cycle's progress.
func (s *AgentService) SummarizeCycle(cycleID, userID uint64) (*model.AgentActivity, error) {
	agent, err := s.findAgentByCapability("summarize")
	if err != nil {
		return nil, err
	}

	task := fmt.Sprintf(
		"Summarize the progress of cycle %d. Get the cycle details, its issues, and their states. "+
			"Provide a concise summary of: what's completed, what's in progress, what's at risk, and what's not started. Cycle ID: %d", cycleID, cycleID)

	ctx := &DispatchContext{
		WorkspaceID: agent.WorkspaceID,
		TriggeredBy: "custom",
	}

	return s.DispatchAgent(agent.ID, userID, task, ctx)
}

// ======== Agent Presence ========

// UpdateAvailability updates the agent's availability status.
func (s *AgentService) UpdateAvailability(agentID uint64, availability model.AgentAvailability) error {
	result := s.db.Model(&model.Agent{}).Where("id = ?", agentID).Update("availability", availability)
	if result.RowsAffected == 0 {
		return common.NotFound("Agent not found")
	}
	return result.Error
}

// UpdateWorkload updates the agent's workload status.
func (s *AgentService) UpdateWorkload(agentID uint64, workload model.AgentWorkload) error {
	result := s.db.Model(&model.Agent{}).Where("id = ?", agentID).Update("workload", workload)
	if result.RowsAffected == 0 {
		return common.NotFound("Agent not found")
	}
	return result.Error
}

// RecordHeartbeat updates the agent's last active timestamp and marks as online.
func (s *AgentService) RecordHeartbeat(agentID uint64) error {
	now := time.Now()
	result := s.db.Model(&model.Agent{}).Where("id = ?", agentID).
		Updates(map[string]interface{}{
			"last_active_at": &now,
			"availability":   model.AvailabilityOnline,
		})
	if result.RowsAffected == 0 {
		return common.NotFound("Agent not found")
	}
	return result.Error
}

// GetPresence returns the presence info for an agent.
func (s *AgentService) GetPresence(agentID uint64) (*model.Agent, error) {
	var agent model.Agent
	if err := s.db.Select("id, name, avatar, availability, workload, last_active_at, running_task_id, queued_task_count").
		Where("id = ?", agentID).First(&agent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Agent not found")
		}
		return nil, common.Internal("Failed to get agent presence")
	}
	return &agent, nil
}

// ListPresence returns presence info for all agents in a workspace.
func (s *AgentService) ListPresence(workspaceID uint64) ([]model.Agent, error) {
	var agents []model.Agent
	if err := s.db.Select("id, name, avatar, availability, workload, last_active_at, running_task_id, queued_task_count").
		Where("workspace_id = ?", workspaceID).Find(&agents).Error; err != nil {
		return nil, common.Internal("Failed to list agent presence")
	}
	return agents, nil
}

// CreateSnapshot creates or updates an agent task snapshot.
func (s *AgentService) CreateSnapshot(agentID, taskID uint64, taskTitle, taskStatus string, progress int, currentStep *string) error {
	var snapshot model.AgentTaskSnapshot
	if err := s.db.Where("agent_id = ? AND task_id = ?", agentID, taskID).First(&snapshot).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return common.Internal("Failed to query snapshot")
		}
		// Create new snapshot
		snapshot = model.AgentTaskSnapshot{
			AgentID:   agentID,
			TaskID:    taskID,
			TaskTitle: taskTitle,
		}
	}

	snapshot.TaskStatus = taskStatus
	snapshot.Progress = progress
	snapshot.CurrentStep = currentStep
	now := time.Now()
	if snapshot.StartedAt == nil {
		snapshot.StartedAt = &now
	}
	snapshot.UpdatedAt = now

	return s.db.Save(&snapshot).Error
}

// GetSnapshots returns all task snapshots for an agent.
func (s *AgentService) GetSnapshots(agentID uint64) ([]model.AgentTaskSnapshot, error) {
	var snapshots []model.AgentTaskSnapshot
	if err := s.db.Where("agent_id = ?", agentID).Order("updated_at DESC").Find(&snapshots).Error; err != nil {
		return nil, common.Internal("Failed to get snapshots")
	}
	return snapshots, nil
}

// UpdatePresenceOnTaskStateChange updates agent presence based on task state changes.
func (s *AgentService) UpdatePresenceOnTaskStateChange(agentID, taskID uint64, taskStatus string) error {
	var agent model.Agent
	if err := s.db.First(&agent, agentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // Agent not found, skip
		}
		return err
	}

	now := time.Now()
	updates := map[string]interface{}{"last_active_at": &now}

	switch taskStatus {
	case "enqueue":
		updates["queued_task_count"] = gorm.Expr("queued_task_count + 1")
		updates["workload"] = model.WorkloadQueued
		if agent.Availability != model.AvailabilityOnline {
			updates["availability"] = model.AvailabilityOnline
		}
	case "claimed":
	case "running":
		updates["running_task_id"] = taskID
		updates["workload"] = model.WorkloadWorking
		if agent.Availability != model.AvailabilityOnline {
			updates["availability"] = model.AvailabilityOnline
		}
	case "completed", "failed", "cancelled":
		// Only clear running_task_id if this is the current running task
		if agent.RunningTaskID != nil && *agent.RunningTaskID == taskID {
			updates["running_task_id"] = nil
		}
		updates["queued_task_count"] = gorm.Expr("GREATEST(queued_task_count - 1, 0)")
		// Check if there are any remaining tasks
		var queuedCount, runningCount int64
		s.db.Model(&model.AgentTask{}).Where("agent_id = ? AND status = ?", agentID, "enqueue").Count(&queuedCount)
		s.db.Model(&model.AgentTask{}).Where("agent_id = ? AND status = ?", agentID, "running").Count(&runningCount)
		if runningCount > 0 {
			updates["workload"] = model.WorkloadWorking
		} else if queuedCount > 0 {
			updates["workload"] = model.WorkloadQueued
		} else {
			updates["workload"] = model.WorkloadIdle
		}
	}

	return s.db.Model(&agent).Updates(updates).Error
}

// ======== Internal Helpers ========

func (s *AgentService) findAgentByCapability(capability string) (*model.Agent, error) {
	var agents []model.Agent
	if err := s.db.Where("status = ?", "active").Find(&agents).Error; err != nil {
		return nil, common.Internal("Failed to find agents")
	}

	for i := range agents {
		var caps []string
		if err := json.Unmarshal(agents[i].Capabilities, &caps); err != nil {
			continue
		}
		for _, c := range caps {
			if strings.EqualFold(c, capability) {
				return &agents[i], nil
			}
		}
		if len(caps) == 0 {
			return &agents[i], nil
		}
		for _, c := range caps {
			if c == "all" {
				return &agents[i], nil
			}
		}
	}

	if len(agents) > 0 {
		return &agents[0], nil
	}

	return nil, common.NotFound("No active agent found with capability: " + capability)
}

func (s *AgentService) buildAgentSystemPrompt(agent *model.Agent, ctx *DispatchContext) string {
	var sb strings.Builder

	if agent.SystemPrompt != nil && *agent.SystemPrompt != "" {
		sb.WriteString(*agent.SystemPrompt)
		sb.WriteString("\n\n")
	} else {
		sb.WriteString(fmt.Sprintf("You are %s, an AI agent in the reqmango project management system. ", agent.Name))
	}

	sb.WriteString("You have access to tools that let you read and modify project data. ")
	sb.WriteString("When given a task, use the available tools to gather information, then provide your analysis or take action. ")
	sb.WriteString("Always explain what you found or did in plain language.\n\n")

	if ctx != nil {
		sb.WriteString(fmt.Sprintf("Current context: workspace_id=%d", ctx.WorkspaceID))
		if ctx.ProjectID != nil {
			sb.WriteString(fmt.Sprintf(", project_id=%d", *ctx.ProjectID))
		}
		if ctx.IssueID != nil {
			sb.WriteString(fmt.Sprintf(", issue_id=%d", *ctx.IssueID))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// ======== Monitoring Dashboard ========

// GetMonitoringStats returns monitoring statistics for a workspace.
func (s *AgentService) GetMonitoringStats(workspaceID uint64) (*response.AgentMonitoringResponse, error) {
	result := &response.AgentMonitoringResponse{
		WorkspaceID: workspaceID,
		GeneratedAt: time.Now(),
	}

	// Get agent presence summary
	presence, err := s.getAgentPresenceSummary(workspaceID)
	if err != nil {
		return nil, err
	}
	result.AgentPresence = presence

	// Get task execution stats
	taskStats, err := s.getTaskExecutionStats(workspaceID)
	if err != nil {
		return nil, err
	}
	result.TaskExecution = taskStats

	// Get tool call stats
	toolStats, err := s.getToolCallStats(workspaceID)
	if err != nil {
		return nil, err
	}
	result.ToolCalls = toolStats

	// Get skill usage stats
	skillStats, err := s.getSkillUsageStats(workspaceID)
	if err != nil {
		return nil, err
	}
	result.SkillUsage = skillStats

	return result, nil
}

// getAgentPresenceSummary returns agent presence statistics.
func (s *AgentService) getAgentPresenceSummary(workspaceID uint64) (response.AgentPresenceSummary, error) {
	var summary response.AgentPresenceSummary

	// Total agents
	s.db.Model(&model.Agent{}).Where("workspace_id = ?", workspaceID).Count(&summary.Total)

	// By availability
	s.db.Model(&model.Agent{}).Where("workspace_id = ? AND availability = ?", workspaceID, model.AvailabilityOnline).Count(&summary.Online)
	s.db.Model(&model.Agent{}).Where("workspace_id = ? AND availability = ?", workspaceID, model.AvailabilityUnstable).Count(&summary.Unstable)
	s.db.Model(&model.Agent{}).Where("workspace_id = ? AND availability = ?", workspaceID, model.AvailabilityOffline).Count(&summary.Offline)
	s.db.Model(&model.Agent{}).Where("workspace_id = ? AND availability = ?", workspaceID, model.AvailabilityArchived).Count(&summary.Archived)

	// By workload
	s.db.Model(&model.Agent{}).Where("workspace_id = ? AND workload = ?", workspaceID, model.WorkloadWorking).Count(&summary.Working)
	s.db.Model(&model.Agent{}).Where("workspace_id = ? AND workload = ?", workspaceID, model.WorkloadQueued).Count(&summary.Queued)
	s.db.Model(&model.Agent{}).Where("workspace_id = ? AND workload = ?", workspaceID, model.WorkloadIdle).Count(&summary.Idle)

	return summary, nil
}

// getTaskExecutionStats returns task execution statistics.
func (s *AgentService) getTaskExecutionStats(workspaceID uint64) (response.TaskExecutionStats, error) {
	var stats response.TaskExecutionStats

	// Total tasks
	s.db.Model(&model.AgentTask{}).Where("workspace_id = ?", workspaceID).Count(&stats.Total)

	// By status
	s.db.Model(&model.AgentTask{}).Where("workspace_id = ? AND status = ?", workspaceID, "completed").Count(&stats.Completed)
	s.db.Model(&model.AgentTask{}).Where("workspace_id = ? AND status = ?", workspaceID, "failed").Count(&stats.Failed)
	s.db.Model(&model.AgentTask{}).Where("workspace_id = ? AND status = ?", workspaceID, "cancelled").Count(&stats.Cancelled)
	s.db.Model(&model.AgentTask{}).Where("workspace_id = ? AND status = ?", workspaceID, "running").Count(&stats.Running)
	s.db.Model(&model.AgentTask{}).Where("workspace_id = ? AND status = ?", workspaceID, "enqueue").Count(&stats.Enqueued)

	// Calculate success rate
	if stats.Total > 0 {
		stats.SuccessRate = float64(stats.Completed) / float64(stats.Total) * 100
	}

	// Calculate average duration
	var avgDuration int64
	s.db.Model(&model.AgentTask{}).
		Where("workspace_id = ? AND status = ? AND duration_ms > 0", workspaceID, "completed").
		Select("AVG(duration_ms)").Scan(&avgDuration)
	stats.AvgDurationMs = avgDuration

	return stats, nil
}

// getToolCallStats returns tool call statistics.
func (s *AgentService) getToolCallStats(workspaceID uint64) (response.ToolCallStats, error) {
	var stats response.ToolCallStats

	// Total tool calls
	s.db.Model(&model.ToolCallLog{}).Where("workspace_id = ?", workspaceID).Count(&stats.Total)

	// By status
	s.db.Model(&model.ToolCallLog{}).Where("workspace_id = ? AND status = ?", workspaceID, "success").Count(&stats.Success)
	s.db.Model(&model.ToolCallLog{}).Where("workspace_id = ? AND status = ?", workspaceID, "failed").Count(&stats.Failed)

	// Calculate average duration
	var avgDuration int64
	s.db.Model(&model.ToolCallLog{}).
		Where("workspace_id = ?", workspaceID).
		Select("AVG(duration_ms)").Scan(&avgDuration)
	stats.AvgDurationMs = avgDuration

	// Top tools (by call count)
	var topTools []struct {
		ToolID     uint64
		ToolName   string
		CallCount  int64
		SuccessRate float64
	}
	s.db.Raw(`
		SELECT tc.tool_id, t.name as tool_name, COUNT(tc.id) as call_count,
		       SUM(CASE WHEN tc.status = 'success' THEN 1 ELSE 0 END) * 100.0 / COUNT(tc.id) as success_rate
		FROM tool_call_logs tc
		JOIN tools t ON tc.tool_id = t.id
		WHERE tc.workspace_id = ?
		GROUP BY tc.tool_id, t.name
		ORDER BY call_count DESC
		LIMIT 10
	`, workspaceID).Scan(&topTools)

	for _, tt := range topTools {
		stats.TopTools = append(stats.TopTools, response.ToolCallFrequency{
			ToolName:   tt.ToolName,
			ToolID:     tt.ToolID,
			CallCount:  tt.CallCount,
			SuccessRate: tt.SuccessRate,
		})
	}

	return stats, nil
}

// getSkillUsageStats returns skill usage statistics.
func (s *AgentService) getSkillUsageStats(workspaceID uint64) (response.SkillUsageStats, error) {
	var stats response.SkillUsageStats

	// Total executions
	s.db.Model(&model.SkillExecutionLog{}).Where("workspace_id = ?", workspaceID).Count(&stats.TotalExecutions)

	// Active skills
	s.db.Model(&model.Skill{}).Where("workspace_id = ? AND status = ?", workspaceID, "active").Count(&stats.ActiveSkills)

	// Top skills (by execution count)
	var topSkills []struct {
		SkillID      uint64
		SkillName    string
		ExecutionCount int64
		AvgDurationMs int64
	}
	s.db.Raw(`
		SELECT sel.skill_id, s.name as skill_name, COUNT(sel.id) as execution_count,
		       AVG(sel.duration_ms) as avg_duration_ms
		FROM skill_execution_logs sel
		JOIN skills s ON sel.skill_id = s.id
		WHERE sel.workspace_id = ?
		GROUP BY sel.skill_id, s.name
		ORDER BY execution_count DESC
		LIMIT 10
	`, workspaceID).Scan(&topSkills)

	for _, ts := range topSkills {
		stats.TopSkills = append(stats.TopSkills, response.SkillUsageFrequency{
			SkillName:    ts.SkillName,
			SkillID:      ts.SkillID,
			ExecutionCount: ts.ExecutionCount,
			AvgDurationMs: ts.AvgDurationMs,
		})
	}

	return stats, nil
}

// ======== Skill Integration ========

// executeAgentSkill executes the agent's skill if configured in the template.
// Returns the skill execution result or empty string if no skill was executed.
func (s *AgentService) executeAgentSkill(ctx context.Context, agent *model.Agent, task string, userID uint64) string {
	if agent.TemplateID == nil {
		return ""
	}

	// Get the agent template
	var template model.AgentTemplate
	if err := s.db.First(&template, *agent.TemplateID).Error; err != nil {
		return ""
	}

	// Check skill mode
	if template.SkillMode == "manual" {
		// In manual mode, skill is selected by user, not automatically
		return ""
	}

	// Get skill ID to execute
	var skillID uint64
	if template.SkillMode == "forced" && template.DefaultSkillID != nil {
		skillID = *template.DefaultSkillID
	} else if template.SkillMode == "auto" {
		// Auto-detect based on available skills
		skillID = s.detectSkillFromTask(ctx, template, task)
	}

	if skillID == 0 {
		return ""
	}

	// Execute the skill
	var skill model.Skill
	if err := s.db.First(&skill, skillID).Error; err != nil {
		return ""
	}

	// Build skill parameters from task
	params := map[string]interface{}{
		"input": task,
	}

	// Execute skill using SkillService
	// Note: This is a simplified integration - in production, you would inject SkillService
	// For now, we parse the SKILL.md and execute the logic directly

	return s.executeSkillDirectly(ctx, &skill, params)
}

// detectSkillFromTask auto-detects the most relevant skill based on task content.
func (s *AgentService) detectSkillFromTask(ctx context.Context, template model.AgentTemplate, task string) uint64 {
	// Parse available skills from template
	var skillIDs []uint64
	if template.AvailableSkills != nil {
		json.Unmarshal(template.AvailableSkills, &skillIDs)
	}

	if len(skillIDs) == 0 {
		return 0
	}

	// Simple keyword matching to detect skill
	for _, id := range skillIDs {
		var skill model.Skill
		if err := s.db.First(&skill, id).Error; err != nil {
			continue
		}

		// Check if task keywords match skill tags
		var tags []string
		if skill.Tags != nil {
			json.Unmarshal(skill.Tags, &tags)
		}

		for _, tag := range tags {
			if strings.Contains(strings.ToLower(task), strings.ToLower(tag)) {
				return id
			}
		}
	}

	// If no match found and default skill is set, use default
	if template.DefaultSkillID != nil {
		return *template.DefaultSkillID
	}

	return 0
}

// executeSkillDirectly executes a skill by parsing SKILL.md and running the steps.
func (s *AgentService) executeSkillDirectly(ctx context.Context, skill *model.Skill, params map[string]interface{}) string {
	if skill == nil || skill.SkillMD == "" {
		return ""
	}

	// Parse SKILL.md and extract steps
	steps := parseSkillSteps(skill.SkillMD)
	if len(steps) == 0 {
		return ""
	}

	var resultBuilder strings.Builder
	resultBuilder.WriteString(fmt.Sprintf("技能: %s\n", skill.Name))
	resultBuilder.WriteString("执行步骤:\n")

	for i, step := range steps {
		resultBuilder.WriteString(fmt.Sprintf("%d. %s\n", i+1, step))
	}

	resultBuilder.WriteString("\n输入参数:\n")
	for k, v := range params {
		resultBuilder.WriteString(fmt.Sprintf("- %s: %v\n", k, v))
	}

	return resultBuilder.String()
}

// parseSkillSteps parses SKILL.md content and extracts step descriptions.
func parseSkillSteps(skillMD string) []string {
	var steps []string
	lines := strings.Split(skillMD, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "## Step") {
			// Extract step title
			step := strings.TrimPrefix(line, "## Step ")
			step = strings.Split(step, ":")[0]
			steps = append(steps, step)
		} else if strings.HasPrefix(line, "- ") && len(steps) > 0 {
			// Add bullet points to the last step
			steps[len(steps)-1] += " " + strings.TrimPrefix(line, "- ")
		}
	}

	return steps
}

// ======== Internal Helpers ========

func (s *AgentService) filterToolsByCapabilities(agent *model.Agent) []llm.Tool {
	allTools := s.aiSvc.GetTools()

	var caps []string
	if agent.Capabilities != nil {
		json.Unmarshal(agent.Capabilities, &caps)
	}

	if len(caps) == 0 {
		return allTools
	}
	for _, c := range caps {
		if c == "all" {
			return allTools
		}
	}

	capTools := map[string][]string{
		"search":    {"search_issues", "get_issue", "get_issue_activities"},
		"create":    {"create_issue"},
		"update":    {"update_issue"},
		"analyze":   {"get_project_stats", "get_issues_summary", "get_cycle_progress"},
		"comment":   {"add_comment"},
		"list":      {"list_members", "list_issue_types", "list_states", "list_labels", "list_cycles", "list_modules", "list_releases", "list_pages"},
		"summarize": {"get_project_stats", "get_issues_summary", "get_cycle_progress", "list_cycles", "get_issue", "get_issue_activities"},
	}

	allowedTools := make(map[string]bool)
	for _, cap := range caps {
		if tools, ok := capTools[cap]; ok {
			for _, t := range tools {
				allowedTools[t] = true
			}
		}
	}

	var filtered []llm.Tool
	for _, t := range allTools {
		if allowedTools[t.Name] {
			filtered = append(filtered, t)
		}
	}

	if len(filtered) == 0 {
		for _, t := range allTools {
			if t.Name != "create_issue" && t.Name != "update_issue" && t.Name != "add_comment" {
				filtered = append(filtered, t)
			}
		}
	}

	return filtered
}

func (s *AgentService) getLatestActivity(agentID uint64) (*model.AgentActivity, error) {
	var activity model.AgentActivity
	if err := s.db.Where("agent_id = ?", agentID).
		Order("executed_at DESC").
		First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("No activity found")
		}
		return nil, common.Internal("Failed to get activity")
	}
	return &activity, nil
}
