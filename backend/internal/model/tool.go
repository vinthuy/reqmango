package model

// Tool represents a registered tool that can be called by agents.
// Tools provide the actual execution capability for agents.
type Tool struct {
	BaseModel

	Name        string          `gorm:"size:128;not null;uniqueIndex" json:"name"`
	Description string          `gorm:"type:text" json:"description"`
	Category    string          `gorm:"size:50;default:general" json:"category"`          // "project_management", "code", "ci_cd", "general", etc.
	IsBuiltin   bool            `gorm:"default:false" json:"is_builtin"`                   // Built-in tools cannot be deleted
	Status      string          `gorm:"size:20;default:active" json:"status"`              // "active" | "disabled"
	ToolType    string          `gorm:"size:30;not null" json:"tool_type"`                 // "api", "function", "workflow"
	Endpoint    *string         `gorm:"size:500" json:"endpoint,omitempty"`                // API endpoint for API type tools
	Method      *string         `gorm:"size:10" json:"method,omitempty"`                   // HTTP method for API type tools
	AuthType    *string         `gorm:"size:20" json:"auth_type,omitempty"`                // "none", "api_key", "oauth2", "bearer"
	AuthConfig  JSONRawMessage `gorm:"type:text;default:'{}'" json:"auth_config"`        // Auth configuration (API key, token, etc.)
	Params      JSONRawMessage `gorm:"type:text;default:'{}'" json:"params"`              // Parameter schema (JSON Schema format)
	RateLimit   int             `gorm:"default:0" json:"rate_limit"`                        // Requests per minute (0 = unlimited)
	Timeout     int             `gorm:"default:30" json:"timeout"`                          // Timeout in seconds
	WorkspaceID *uint64         `gorm:"index" json:"workspace_id,omitempty"`               // Null for built-in tools
	MCPConfigID *uint64         `gorm:"index" json:"mcp_config_id,omitempty"`              // MCP server that owns this tool (for tool_type=mcp)

	// Relationships
	Workspace  *Workspace  `gorm:"foreignKey:WorkspaceID" json:"-"`
	MCPConfig  *MCPConfig  `gorm:"foreignKey:MCPConfigID" json:"-"`
}

func (Tool) TableName() string { return "tools" }

// ToolPermission represents the permission for an agent to call a specific tool.
type ToolPermission struct {
	BaseModel

	WorkspaceID    uint64 `gorm:"not null;index" json:"workspace_id"`
	AgentTemplateID *uint64 `gorm:"index" json:"agent_template_id,omitempty"`    // Permission for specific template (or all if null)
	ToolID         uint64 `gorm:"not null;index" json:"tool_id"`
	Allowed        bool   `gorm:"default:true" json:"allowed"`                    // true = allow, false = deny

	// Relationships
	Tool *Tool `gorm:"foreignKey:ToolID" json:"-"`
}

func (ToolPermission) TableName() string { return "tool_permissions" }

// ToolCallLog represents a log entry for a tool call.
type ToolCallLog struct {
	BaseModel

	WorkspaceID     uint64         `gorm:"not null;index" json:"workspace_id"`
	AgentTaskID     *uint64        `gorm:"index" json:"agent_task_id,omitempty"`    // Related task (if called from a task)
	ToolID          uint64         `gorm:"not null;index" json:"tool_id"`
	AgentID         *uint64        `gorm:"index" json:"agent_id,omitempty"`          // Agent that made the call
	InputParams     JSONRawMessage `gorm:"type:text;default:'{}'" json:"input_params"`
	OutputResult    JSONRawMessage `gorm:"type:text;default:'{}'" json:"output_result"`
	Status          string         `gorm:"size:20;default:success" json:"status"`     // "success", "failed", "timeout"
	ErrorMessage    *string        `gorm:"type:text" json:"error_message,omitempty"`
	DurationMs      int64          `gorm:"default:0" json:"duration_ms"`
	RateLimited     bool           `gorm:"default:false" json:"rate_limited"`

	// Relationships
	Tool *Tool `gorm:"foreignKey:ToolID" json:"-"`
}

func (ToolCallLog) TableName() string { return "tool_call_logs" }
