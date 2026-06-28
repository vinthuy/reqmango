package model

// Notification represents a user notification.
type Notification struct {
	BaseModel

	Title    string `gorm:"size:255;not null" json:"title"`
	Message  string `gorm:"type:text;not null" json:"message"`
	Type     string `gorm:"size:20;default:info" json:"type"`        // info, warning, error, success
	Priority string `gorm:"size:20;default:medium" json:"priority"`  // low, medium, high, urgent
	IsRead   bool   `gorm:"default:false;index" json:"is_read"`
	ReadAt   *string `gorm:"type:timestamptz" json:"read_at"`
	ActionURL *string `gorm:"size:500" json:"action_url"`

	RecipientID uint64  `gorm:"not null;index" json:"recipient_id"`
	SenderID    *uint64 `json:"sender_id"`
	ProjectID   *uint64 `gorm:"index" json:"project_id"`
	IssueID     *uint64 `gorm:"index" json:"issue_id"`

	// Relationships
	Recipient User  `gorm:"foreignKey:RecipientID" json:"-"`
	Sender    *User `gorm:"foreignKey:SenderID" json:"-"`
}

func (Notification) TableName() string {
	return "notifications"
}
