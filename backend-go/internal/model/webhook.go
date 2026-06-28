package model

// Webhook defines an outgoing webhook configuration.
type Webhook struct {
	BaseModel

	Name    string `gorm:"size:100;not null" json:"name"`
	URL     string `gorm:"size:500;not null" json:"url"`
	Secret  string `gorm:"size:255" json:"-"` // HMAC signing secret, never exposed
	Events  string `gorm:"size:500;not null;default:'issue_created,issue_updated,state_changed'" json:"events"`
	IsActive bool  `gorm:"default:true" json:"is_active"`

	ProjectID   uint64 `gorm:"not null;index" json:"project_id"`
	WorkspaceID uint64 `gorm:"not null" json:"workspace_id"`
}

func (Webhook) TableName() string { return "webhooks" }
