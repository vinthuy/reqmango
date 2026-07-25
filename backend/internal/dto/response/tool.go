package response

import "time"

// ToolResponse represents the response for a tool.
type ToolResponse struct {
	ID          uint64       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Category    string       `json:"category"`
	IsBuiltin   bool         `json:"is_builtin"`
	Status      string       `json:"status"`
	ToolType    string       `json:"tool_type"`
	Endpoint    *string      `json:"endpoint"`
	Method      *string      `json:"method"`
	AuthType    *string      `json:"auth_type"`
	Params      interface{}  `json:"params"`
	RateLimit   int          `json:"rate_limit"`
	Timeout     int          `json:"timeout"`
	WorkspaceID *uint64      `json:"workspace_id"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// ToolCallResponse represents the response for a tool call.
type ToolCallResponse struct {
	ID          uint64        `json:"id"`
	ToolID      uint64        `json:"tool_id"`
	ToolName    string        `json:"tool_name"`
	InputParams interface{}   `json:"input_params"`
	OutputResult interface{}  `json:"output_result"`
	Status      string        `json:"status"`
	ErrorMessage *string      `json:"error_message"`
	DurationMs  int64         `json:"duration_ms"`
	CreatedAt   time.Time     `json:"created_at"`
}

// ToolCallLogResponse represents the response for a tool call log.
type ToolCallLogResponse struct {
	ID             uint64        `json:"id"`
	WorkspaceID    uint64        `json:"workspace_id"`
	AgentTaskID    *uint64       `json:"agent_task_id"`
	ToolID         uint64        `json:"tool_id"`
	ToolName       string        `json:"tool_name"`
	AgentID        *uint64       `json:"agent_id"`
	InputParams    interface{}   `json:"input_params"`
	OutputResult   interface{}   `json:"output_result"`
	Status         string        `json:"status"`
	ErrorMessage   *string       `json:"error_message"`
	DurationMs     int64         `json:"duration_ms"`
	RateLimited    bool          `json:"rate_limited"`
	CreatedAt      time.Time     `json:"created_at"`
}
