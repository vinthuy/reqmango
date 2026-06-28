package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/reqmanpy/backend-go/internal/common"
	"github.com/reqmanpy/backend-go/internal/model"
	"gorm.io/gorm"
)

// AgentService manages AI agents and their actions.
type AgentService struct {
	db        *gorm.DB
	llm       *LLMClient
	issueSvc  *IssueService
	aiSvc     *AIService
}

// NewAgentService creates a new AgentService.
func NewAgentService(db *gorm.DB, llm *LLMClient, issueSvc *IssueService, aiSvc *AIService) *AgentService {
	return &AgentService{
		db:       db,
		llm:      llm,
		issueSvc: issueSvc,
		aiSvc:    aiSvc,
	}
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

// createRequest is the shape for creating/updating agents.
type AgentCreateRequest struct {
	Name         string   `json:"name" binding:"required"`
	Avatar       string   `json:"avatar"`
	AgentType    string   `json:"agent_type"`
	Capabilities []string `json:"capabilities"`
	Status       string   `json:"status"`
	ModelOverride *string `json:"model_override"`
	SystemPrompt  *string `json:"system_prompt"`
}

type AgentUpdateRequest struct {
	Name         *string  `json:"name"`
	Avatar       *string  `json:"avatar"`
	AgentType    *string  `json:"agent_type"`
	Capabilities *[]string `json:"capabilities"`
	Status       *string  `json:"status"`
	ModelOverride *string `json:"model_override"`
	SystemPrompt  *string `json:"system_prompt"`
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

	agent := model.Agent{
		WorkspaceID:  workspaceID,
		Name:         req.Name,
		Avatar:       req.Avatar,
		AgentType:    req.AgentType,
		Capabilities: capsJSON,
		Status:       req.Status,
		ModelOverride: req.ModelOverride,
		SystemPrompt:  req.SystemPrompt,
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

	if err := s.db.Model(agent).Updates(updates).Error; err != nil {
		return nil, common.Internal("Failed to update agent")
	}
	// Reload
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

// recordActivity writes an agent activity entry to the database.
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
	TriggeredBy string  `json:"triggered_by"` // e.g., "manual", "mention", "auto_triage"
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

	// Build system prompt from agent configuration
	systemPrompt := s.buildAgentSystemPrompt(agent, ctx)

	// Convert agent capabilities to LLM tools
	tools := s.filterToolsByCapabilities(agent)

	// Call the LLM
	messages := []Message{
		{Role: "user", Content: task},
	}

	resp, llmErr := s.llm.ChatSync(context.Background(), systemPrompt, messages, tools)
	if llmErr != nil {
		// Record failed attempt
		s.recordActivity(agent, ctx.IssueID, "dispatch",
			fmt.Sprintf("Failed: %v", llmErr), task, userID)
		return nil, common.Internal(fmt.Sprintf("Agent LLM call failed: %v", llmErr))
	}

	// Execute tool calls if any
	var resultBuilder strings.Builder
	resultBuilder.WriteString(fmt.Sprintf("Agent %s processed task: %s\n", agent.Name, task))

	if len(resp.ToolCalls) > 0 {
		resultBuilder.WriteString(fmt.Sprintf("Executed %d tool(s):\n", len(resp.ToolCalls)))
		for _, tc := range resp.ToolCalls {
			resultBuilder.WriteString(fmt.Sprintf("- %s(%v)\n", tc.Name, tc.Input))
		}
	}

	if resp.Content != "" {
		resultBuilder.WriteString(fmt.Sprintf("\nResponse:\n%s", resp.Content))
	}

	summary := resultBuilder.String()
	if len(summary) > 500 {
		summary = summary[:500] + "..."
	}

	// Record the activity
	s.recordActivity(agent, ctx.IssueID, "dispatch", summary, task, userID)

	// Also write to IssueActivity if there's an associated issue
	if ctx.IssueID != nil {
		verb := "agent_dispatch"
		comment := fmt.Sprintf("AI Agent %s: %s", agent.Name, summary)
		activity := model.IssueActivity{
			IssueID: ctx.IssueID,
			Verb:    verb,
			Comment: &comment,
			ActorID: &userID,
		}
		s.db.Create(&activity)
	}

	return s.getLatestActivity(agent.ID)
}

// ======== Auto Triage ========

// AutoTriage analyzes an unsorted issue and suggests type, priority, and labels.
func (s *AgentService) AutoTriage(issueID, userID uint64) (*model.AgentActivity, error) {
	// Find an active agent with "triage" or "analyze" capability
	agent, err := s.findAgentByCapability("analyze")
	if err != nil {
		return nil, err
	}

	// Load issue details via issue service
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

// ======== Internal Helpers ========

// findAgentByCapability finds an active agent with the given capability.
func (s *AgentService) findAgentByCapability(capability string) (*model.Agent, error) {
	// First, look for an agent with the specific capability in its JSON capabilities array.
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
		// Also match if capabilities is empty (agent can do everything) or contains "all"
		if len(caps) == 0 {
			return &agents[i], nil
		}
		for _, c := range caps {
			if c == "all" {
				return &agents[i], nil
			}
		}
	}

	// Fallback: return any active agent
	if len(agents) > 0 {
		return &agents[0], nil
	}

	return nil, common.NotFound("No active agent found with capability: " + capability)
}

// buildAgentSystemPrompt creates a system prompt from the agent configuration.
func (s *AgentService) buildAgentSystemPrompt(agent *model.Agent, ctx *DispatchContext) string {
	var sb strings.Builder

	if agent.SystemPrompt != nil && *agent.SystemPrompt != "" {
		sb.WriteString(*agent.SystemPrompt)
		sb.WriteString("\n\n")
	} else {
		sb.WriteString(fmt.Sprintf("You are %s, an AI agent in the ReqManPy project management system. ", agent.Name))
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

// filterToolsByCapabilities returns only the tools that match the agent's capabilities.
func (s *AgentService) filterToolsByCapabilities(agent *model.Agent) []Tool {
	allTools := s.aiSvc.GetTools()

	var caps []string
	if agent.Capabilities != nil {
		json.Unmarshal(agent.Capabilities, &caps)
	}

	// Empty capabilities or "all" means the agent can use all tools
	if len(caps) == 0 {
		return allTools
	}
	for _, c := range caps {
		if c == "all" {
			return allTools
		}
	}

	// Filter tools by the agent's capabilities
	capSet := make(map[string]bool)
	for _, c := range caps {
		capSet[strings.ToLower(c)] = true
	}

	// Map capabilities to tool names
	capTools := map[string][]string{
		"search":   {"search_issues", "get_issue", "get_issue_activities"},
		"create":   {"create_issue"},
		"update":   {"update_issue"},
		"analyze":  {"get_project_stats", "get_issues_summary", "get_cycle_progress"},
		"comment":  {"add_comment"},
		"list":     {"list_members", "list_issue_types", "list_states", "list_labels", "list_cycles", "list_modules", "list_releases", "list_pages"},
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

	var filtered []Tool
	for _, t := range allTools {
		if allowedTools[t.Name] {
			filtered = append(filtered, t)
		}
	}

	// If no tools matched, return all safe (read-only) tools
	if len(filtered) == 0 {
		for _, t := range allTools {
			if t.Name != "create_issue" && t.Name != "update_issue" && t.Name != "add_comment" {
				filtered = append(filtered, t)
			}
		}
	}

	return filtered
}

// getLatestActivity returns the most recent activity for an agent.
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
