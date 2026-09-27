package model

import (
	"encoding/json"
	"time"
)

// IssueSuggestion is one proposed work-item change an agent surfaced instead of
// applying it directly. Suggestions are stored on the agent's Comment so the UI
// can offer a one-click apply and the user stays in control of the mutation.
type IssueSuggestion struct {
	Field   string      `json:"field"`             // title|priority|type|state|assignee|description
	Value   interface{} `json:"value"`             // target value: string, or numeric id
	Label   string      `json:"label"`             // human-readable target, e.g. "Bug (type_id=3)"
	Current string      `json:"current,omitempty"` // human-readable current value
	Reason  string      `json:"reason,omitempty"`  // why the agent suggests this change
	// NoOp marks a proposal that recommends keeping the current value. It is
	// surfaced for its reasoning but must not offer an apply action.
	NoOp bool `json:"noop,omitempty"`
}

// Comment represents a comment on an issue.
type Comment struct {
	BaseModel

	IssueID    uint64  `gorm:"not null;index" json:"issue_id"`
	AuthorID   *uint64 `json:"author_id"`
	AgentID    *uint64 `gorm:"index" json:"agent_id,omitempty"` // set when the comment was written by an agent
	Body       string  `gorm:"type:text;not null" json:"body"`
	IsResolved bool    `gorm:"default:false" json:"is_resolved"`
	ParentID   *uint64 `json:"parent_id"` // for threaded replies

	// Suggestions holds the structured proposals attached to an agent comment.
	// SuggestionsAppliedAt is set once the user has applied (or dismissed) them,
	// so the actions never render twice.
	Suggestions          json.RawMessage `gorm:"type:jsonb" json:"suggestions,omitempty"`
	SuggestionsAppliedAt *time.Time      `json:"suggestions_applied_at,omitempty"`

	Issue  Issue  `gorm:"foreignKey:IssueID" json:"-"`
	Author *User  `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	Agent  *Agent `gorm:"foreignKey:AgentID" json:"agent,omitempty"`
}

func (Comment) TableName() string { return "comments" }
