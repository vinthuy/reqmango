package model

import "encoding/json"

// Plugin represents a workspace-installed plugin instance.
type Plugin struct {
	BaseModel

	Name        string `gorm:"size:100;not null" json:"name"`
	Slug        string `gorm:"size:100;not null;index" json:"slug"`
	Description *string `gorm:"size:500" json:"description"`
	Author      string  `gorm:"size:100" json:"author"`
	Version     string  `gorm:"size:20;default:1.0.0" json:"version"`
	IconURL     string  `gorm:"size:500" json:"icon_url"`

	// Type: "webhook" | "notification" | "importer" | "exporter" | "automation"
	Type string `gorm:"size:30;not null" json:"type"`

	// EntryPoint: plugin entry point (URL, endpoint, script path, etc.)
	EntryPoint string `gorm:"size:500" json:"entry_point"`

	// ConfigSchema: JSON Schema defining what config fields the plugin requires
	ConfigSchema json.RawMessage `gorm:"type:jsonb" json:"config_schema"`

	// Config: user-provided runtime configuration values
	Config json.RawMessage `gorm:"type:jsonb" json:"config"`

	// SubscribedEvents: JSON array of event types this plugin listens to
	// e.g. ["issue.created", "issue.updated", "comment.created"]
	SubscribedEvents json.RawMessage `gorm:"type:jsonb" json:"subscribed_events"`

	Enabled       bool `gorm:"default:false" json:"enabled"`
	WorkspaceID   uint64 `gorm:"not null;index" json:"workspace_id"`
	InstalledByID uint64 `gorm:"not null" json:"installed_by_id"`

	// Relationships
	Workspace   Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
	InstalledBy User      `gorm:"foreignKey:InstalledByID" json:"-"`
}

func (Plugin) TableName() string {
	return "plugins"
}

// PluginEventLog records plugin execution history for debugging/auditing.
type PluginEventLog struct {
	BaseModel

	PluginID    uint64 `gorm:"not null;index" json:"plugin_id"`
	EventType   string  `gorm:"size:50;not null" json:"event_type"`
	Status      string  `gorm:"size:20;default:success" json:"status"` // success | error
	RequestBody string  `gorm:"type:text" json:"request_body"`
	ResponseBody string `gorm:"type:text" json:"response_body"`
	StatusCode  int     `json:"status_code"`
	DurationMs  int64   `json:"duration_ms"`

	// Relationships
	Plugin Plugin `gorm:"foreignKey:PluginID" json:"-"`
}

func (PluginEventLog) TableName() string {
	return "plugin_event_logs"
}
