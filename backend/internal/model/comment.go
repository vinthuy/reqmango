package model

// Comment represents a comment on an issue.
type Comment struct {
	BaseModel

	IssueID  uint64  `gorm:"not null;index" json:"issue_id"`
	AuthorID *uint64 `json:"author_id"`
	Body     string  `gorm:"type:text;not null" json:"body"`
	IsResolved bool `gorm:"default:false" json:"is_resolved"`
	ParentID *uint64 `json:"parent_id"` // for threaded replies

	Issue  Issue `gorm:"foreignKey:IssueID;constraint:OnDelete:CASCADE" json:"-"`
	Author *User `gorm:"foreignKey:AuthorID" json:"-"`
}

func (Comment) TableName() string { return "comments" }
