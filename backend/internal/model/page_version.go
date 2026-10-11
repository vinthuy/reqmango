package model

// PageVersion represents a historical snapshot of a page.
type PageVersion struct {
	BaseModel

	PageID        uint64  `gorm:"not null;index" json:"page_id"`
	Title         string  `gorm:"size:255;not null" json:"title"`
	Content       string  `gorm:"type:text;not null;default:''" json:"content"`
	ContentJSON   *string `gorm:"type:jsonb" json:"content_json"`
	VersionNumber int     `gorm:"not null;default:1" json:"version_number"`
	ChangeSummary *string `gorm:"size:500" json:"change_summary"`

	Page Page `gorm:"foreignKey:PageID" json:"-"`
}

func (PageVersion) TableName() string {
	return "page_versions"
}
