package model

// Label represents a project label for categorizing issues.
type Label struct {
	BaseModel

	Name        string  `gorm:"size:255;not null" json:"name"`
	Color       string  `gorm:"size:50;default:#6B7280" json:"color"`
	Description *string `gorm:"size:255" json:"description"`
	ProjectID   *uint64 `gorm:"index" json:"project_id"`
	WorkspaceID uint64  `gorm:"not null" json:"workspace_id"`

	// Relationships
	Project    Project      `gorm:"foreignKey:ProjectID" json:"-"`
	IssueLinks []IssueLabel `gorm:"foreignKey:LabelID" json:"-"`
}

func (Label) TableName() string {
	return "labels"
}
