package model

// PageTemplate represents a reusable page template.
type PageTemplate struct {
	BaseModel

	Name        string  `gorm:"size:255;not null" json:"name"`
	Description string  `gorm:"type:text;default:''" json:"description"`
	Content     string  `gorm:"type:text;default:''" json:"content"`
	ContentJSON *string `gorm:"type:jsonb" json:"content_json"`
	IsDefault   bool    `gorm:"default:false" json:"is_default"`

	WorkspaceID uint64  `gorm:"not null;index" json:"workspace_id"`
	ProjectID   *uint64 `gorm:"index" json:"project_id"`
}

func (PageTemplate) TableName() string {
	return "page_templates"
}
