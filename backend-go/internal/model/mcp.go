package model

import "time"

type MCPConfig struct {
	BaseModel
	WorkspaceID   uint64     `gorm:"not null;index" json:"workspace_id"`
	Name          string     `gorm:"size:255;not null" json:"name"`
	Description   string     `gorm:"size:500" json:"description"`
	ServerURL     string     `gorm:"size:500;not null" json:"server_url"`
	TransportType string     `gorm:"size:20;default:sse" json:"transport_type"`
	APIKey        string     `gorm:"size:500" json:"api_key"`
	ToolsConfig   string     `gorm:"type:text" json:"tools_config"`
	IsEnabled     bool       `gorm:"default:true" json:"is_enabled"`
	LastSyncAt    *time.Time `json:"last_sync_at"`
	Workspace     Workspace  `gorm:"foreignKey:WorkspaceID" json:"-"`
}

func (MCPConfig) TableName() string {
	return "mcp_configs"
}
