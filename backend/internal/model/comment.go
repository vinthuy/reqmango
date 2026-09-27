package model

// Comment represents a comment on an issue.
type Comment struct {
	BaseModel

	IssueID    uint64  `gorm:"not null;index" json:"issue_id"`
	AuthorID   *uint64 `json:"author_id"`
	AgentID    *uint64 `gorm:"index" json:"agent_id,omitempty"` // set when the comment was written by an agent
	Body       string  `gorm:"type:text;not null" json:"body"`
	IsResolved bool    `gorm:"default:false" json:"is_resolved"`
	ParentID   *uint64 `json:"parent_id"` // for threaded replies

	Issue  Issue  `gorm:"foreignKey:IssueID" json:"-"`
	Author *User  `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	Agent  *Agent `gorm:"foreignKey:AgentID" json:"agent,omitempty"`
}

func (Comment) TableName() string { return "comments" }
