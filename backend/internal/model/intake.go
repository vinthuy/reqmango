package model

// ProjectIntakeSetting holds per-project intake channel configuration.
type ProjectIntakeSetting struct {
	BaseModel

	ProjectID      uint64 `gorm:"not null;uniqueIndex" json:"project_id"`
	FormEnabled    bool   `gorm:"default:true" json:"form_enabled"`
	WebhookEnabled bool   `gorm:"default:true" json:"webhook_enabled"`
	EmailEnabled   bool   `gorm:"default:true" json:"email_enabled"`
	// Token authenticates webhook and inbound-email deliveries; rotate to revoke.
	Token    string `gorm:"size:64;uniqueIndex" json:"token"`
	SLAHours int    `gorm:"default:48" json:"sla_hours"`
}
