package model

type SlackConnection struct {
	BaseModel
	WorkspaceID      uint64 `gorm:"not null;index" json:"workspace_id"`
	ProjectID        uint64 `gorm:"not null;index" json:"project_id"`
	ChannelName      string `gorm:"size:255;not null" json:"channel_name"`
	WebhookURL       string `gorm:"size:500;not null" json:"webhook_url"`
	BotToken         string `gorm:"size:500" json:"bot_token"`
	IsEnabled        bool   `gorm:"default:true" json:"is_enabled"`
	NotifyOnCreate   bool   `gorm:"default:true" json:"notify_on_create"`
	NotifyOnUpdate   bool   `gorm:"default:true" json:"notify_on_update"`
	NotifyOnComment  bool   `gorm:"default:false" json:"notify_on_comment"`
	NotifyOnComplete bool   `gorm:"default:true" json:"notify_on_complete"`

	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
	Project   Project   `gorm:"foreignKey:ProjectID" json:"-"`
}

func (SlackConnection) TableName() string {
	return "slack_connections"
}
