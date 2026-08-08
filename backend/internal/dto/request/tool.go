package request

import "encoding/json"

// CreateToolRequest represents the request to create a new tool.
type CreateToolRequest struct {
	Name        string          `json:"name" binding:"required"`
	Description string          `json:"description"`
	Category    string          `json:"category"`
	ToolType    string          `json:"tool_type" binding:"required,oneof=api function workflow"`
	Endpoint    *string         `json:"endpoint"`
	Method      *string         `json:"method"`
	AuthType    *string         `json:"auth_type"`
	AuthConfig  json.RawMessage `json:"auth_config"`
	Params      json.RawMessage `json:"params"`
	RateLimit   int             `json:"rate_limit"`
	Timeout     int             `json:"timeout"`
}

// UpdateToolRequest represents the request to update a tool.
type UpdateToolRequest struct {
	Name        *string          `json:"name"`
	Description *string          `json:"description"`
	Category    *string          `json:"category"`
	Status      *string          `json:"status"`
	ToolType    *string          `json:"tool_type"`
	Endpoint    *string          `json:"endpoint"`
	Method      *string          `json:"method"`
	AuthType    *string          `json:"auth_type"`
	AuthConfig  *json.RawMessage `json:"auth_config"`
	Params      *json.RawMessage `json:"params"`
	RateLimit   *int             `json:"rate_limit"`
	Timeout     *int             `json:"timeout"`
}

// CallToolRequest represents the request to call a tool.
type CallToolRequest struct {
	ToolID      uint64          `json:"tool_id" binding:"required"`
	InputParams json.RawMessage `json:"input_params"`

	// === Hardening fields ===
	AgentTemplateID *uint64 `json:"agent_template_id,omitempty"` // ToolPermission 白/黑名单匹配
	AgentTaskID     *uint64 `json:"agent_task_id,omitempty"`     // 任务关联（审计）
	CallerUserID    uint64  `json:"-"`                           // 从 auth middleware 注入，禁止客户端伪造
}

// CreateToolPermissionRequest represents the request to create a tool permission.
type CreateToolPermissionRequest struct {
	AgentTemplateID *uint64 `json:"agent_template_id"`
	ToolID          uint64  `json:"tool_id" binding:"required"`
	Allowed         bool    `json:"allowed"`
}
