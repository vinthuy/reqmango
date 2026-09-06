package client

import (
	"context"
	"encoding/json"
	"strconv"
	"time"
)

// Agent is a workspace AI agent.
type Agent struct {
	ID            uint64     `json:"id"`
	WorkspaceID   uint64     `json:"workspace_id"`
	Name          string     `json:"name"`
	Avatar        string     `json:"avatar"`
	AgentType     string     `json:"agent_type"` // "builtin" | "custom"
	Capabilities  []string   `json:"capabilities"`
	Status        string     `json:"status"` // "active" | "inactive"
	ModelOverride *string    `json:"model_override"`
	SystemPrompt  *string    `json:"system_prompt"`
	TemplateID    *uint64    `json:"template_id"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// ListAgents returns workspace agents (GET /workspaces/:ws/agents).
func (c *Client) ListAgents(ctx context.Context, workspaceID uint64) ([]Agent, error) {
	var out []Agent
	_, err := c.GetJSON(ctx, "/workspaces/"+strconv.FormatUint(workspaceID, 10)+"/agents", nil, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AgentActivity is the dispatch response.
type AgentActivity struct {
	ID            uint64     `json:"id"`
	AgentID       uint64     `json:"agent_id"`
	IssueID       *uint64    `json:"issue_id"`
	Action        string     `json:"action"` // "dispatch" | "auto_triage" | ...
	ResultSummary string     `json:"result_summary"`
	Rating        *int       `json:"rating"`
	ExecutedAt    *time.Time `json:"executed_at"`
	AgentName     string     `json:"agent_name"`
	TaskContext   string     `json:"task_context"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// DispatchAgent triggers an agent (POST /workspaces/:ws/agents/:id/dispatch).
// Long-running: uses a 5-minute context timeout.
func (c *Client) DispatchAgent(ctx context.Context, workspaceID, agentID uint64, task string, issueID, projectID *uint64) (*AgentActivity, error) {
	body := map[string]any{"task": task}
	if issueID != nil {
		body["issue_id"] = *issueID
	}
	if projectID != nil {
		body["project_id"] = *projectID
	}
	var out AgentActivity
	_, err := c.PostJSON(ctx,
		"/workspaces/"+strconv.FormatUint(workspaceID, 10)+"/agents/"+strconv.FormatUint(agentID, 10)+"/dispatch",
		nil, body, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// AgentTask is the GET /workspaces/:ws/agent-tasks/:id response.
type AgentTask struct {
	ID            uint64          `json:"id"`
	Title         string          `json:"title"`
	Description   string          `json:"description"`
	Status        string          `json:"status"` // enqueue|claimed|running|completed|failed|cancelled
	Priority      int             `json:"priority"`
	Progress      int             `json:"progress"`
	TaskType      string          `json:"task_type"`
	OutputData    json.RawMessage `json:"output_data"`
	ErrorInfo     string          `json:"error_info"`
	FailureReason string          `json:"failure_reason"`
	WorkspaceID   uint64          `json:"workspace_id"`
	ProjectID     *uint64         `json:"project_id"`
	IssueID       *uint64         `json:"issue_id"`
	EnqueuedAt    *time.Time      `json:"enqueued_at"`
	StartedAt     *time.Time      `json:"started_at"`
	CompletedAt   *time.Time      `json:"completed_at"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// GetAgentTask fetches an agent task by ID.
func (c *Client) GetAgentTask(ctx context.Context, workspaceID, taskID uint64) (*AgentTask, error) {
	var out AgentTask
	path := "/workspaces/" + strconv.FormatUint(workspaceID, 10) + "/agent-tasks/" + strconv.FormatUint(taskID, 10)
	if _, err := c.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// TaskLog is one agent task log entry.
type TaskLog struct {
	ID        uint64    `json:"id"`
	TaskID    uint64    `json:"task_id"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Metadata  string    `json:"metadata"`
	CreatedAt time.Time `json:"created_at"`
}

// GetAgentTaskLogs fetches task logs.
func (c *Client) GetAgentTaskLogs(ctx context.Context, workspaceID, taskID uint64) ([]TaskLog, error) {
	var out []TaskLog
	path := "/workspaces/" + strconv.FormatUint(workspaceID, 10) + "/agent-tasks/" + strconv.FormatUint(taskID, 10) + "/logs"
	if _, err := c.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
