package model

import "encoding/json"

// AIConfig stores workspace-level AI configuration.
type AIConfig struct {
	BaseModel

	Provider  string `gorm:"size:20;default:deepseek" json:"provider"`
	Model     string `gorm:"size:50;default:deepseek-chat" json:"model"`
	APIKey    string `gorm:"size:500;not null;column:api_key" json:"-"` // never serialized
	MaxTokens int    `gorm:"default:4096" json:"max_tokens"`
	IsActive  bool   `gorm:"default:true" json:"is_active"`

	WorkspaceID uint64 `gorm:"not null;uniqueIndex" json:"workspace_id"`
}

func (AIConfig) TableName() string {
	return "ai_configs"
}

// AIThread stores a conversation thread.
type AIThread struct {
	BaseModel

	Title       string  `gorm:"size:255" json:"title"`
	WorkspaceID uint64  `gorm:"not null;index" json:"workspace_id"`
	ProjectID   *uint64 `gorm:"index" json:"project_id"`
	UserID      uint64  `gorm:"not null;index" json:"user_id"`

	Messages []AIMessage `gorm:"foreignKey:ThreadID" json:"messages,omitempty"`
}

func (AIThread) TableName() string {
	return "ai_threads"
}

// AIMessage stores a single message in a conversation.
type AIMessage struct {
	BaseModel

	ThreadID uint64 `gorm:"not null;index" json:"thread_id"`
	Role     string `gorm:"size:20;not null" json:"role"` // user | assistant | system | tool
	Content  string `gorm:"type:text;not null" json:"content"`

	ToolCalls json.RawMessage `gorm:"type:jsonb;column:tool_calls" json:"tool_calls,omitempty"`
	ToolName  *string         `gorm:"size:50" json:"tool_name,omitempty"`

	Thread *AIThread `gorm:"foreignKey:ThreadID" json:"-"`
}

func (AIMessage) TableName() string {
	return "ai_messages"
}
