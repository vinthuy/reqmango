package model

import "time"

// Issue is the core work item model.
type Issue struct {
	BaseModel

	Name                string     `gorm:"size:255;not null" json:"name"`
	DescriptionHTML     string     `gorm:"type:text;default:<p></p>" json:"description_html"`
	DescriptionJSON     *string    `gorm:"type:jsonb" json:"description_json"`
	DescriptionStripped *string    `gorm:"type:text" json:"description_stripped"`
	Priority            string     `gorm:"size:30;default:none" json:"priority"`
	SequenceID          int        `gorm:"default:1" json:"sequence_id"`
	SortOrder           float64    `gorm:"default:65535" json:"sort_order"`
	StartDate           *time.Time `json:"start_date"`
	TargetDate          *time.Time `json:"target_date"`
	CompletedAt         *time.Time `json:"completed_at"`
	IsDraft             bool       `gorm:"default:false" json:"is_draft"`
	ArchivedAt          *time.Time `json:"archived_at"`

	ProjectID      uint64  `gorm:"not null;index" json:"project_id"`
	WorkspaceID    uint64  `gorm:"not null" json:"workspace_id"`
	ParentID       *uint64 `json:"parent_id"`
	IssueTypeID    *uint64 `json:"issue_type_id"`
	StateID        uint64  `gorm:"not null;index" json:"state_id"`
	ExternalID     *string `gorm:"size:255" json:"external_id"`
	ExternalSource *string `gorm:"size:255" json:"external_source"`

	// Relationships
	Project        Project         `gorm:"foreignKey:ProjectID" json:"-"`
	State          State           `gorm:"foreignKey:StateID" json:"-"`
	IssueType      IssueType       `gorm:"foreignKey:IssueTypeID" json:"-"`
	Parent         *Issue          `gorm:"foreignKey:ParentID" json:"-"`
	SubIssues      []Issue         `gorm:"foreignKey:ParentID" json:"-"`
	AssigneeLinks  []IssueAssignee `gorm:"foreignKey:IssueID" json:"-"`
	LabelLinks     []IssueLabel    `gorm:"foreignKey:IssueID" json:"-"`
	CycleLink      *IssueCycle     `gorm:"foreignKey:IssueID" json:"-"`
	Activities     []IssueActivity `gorm:"foreignKey:IssueID" json:"-"`
}

func (Issue) TableName() string {
	return "issues"
}

// IssueAssignee is a join table for issue-user (assignee) associations.
type IssueAssignee struct {
	IssueID uint64 `gorm:"primaryKey;autoIncrement:false" json:"issue_id"`
	UserID  uint64 `gorm:"primaryKey;autoIncrement:false" json:"user_id"`

	Issue Issue `gorm:"foreignKey:IssueID;constraint:OnDelete:CASCADE" json:"-"`
	User  User  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

func (IssueAssignee) TableName() string {
	return "issue_assignees"
}

// IssueLabel is a join table for issue-label associations.
type IssueLabel struct {
	IssueID uint64 `gorm:"primaryKey;autoIncrement:false" json:"issue_id"`
	LabelID uint64 `gorm:"primaryKey;autoIncrement:false" json:"label_id"`

	Issue Issue `gorm:"foreignKey:IssueID;constraint:OnDelete:CASCADE" json:"-"`
	Label Label `gorm:"foreignKey:LabelID;constraint:OnDelete:CASCADE" json:"-"`
}

func (IssueLabel) TableName() string {
	return "issue_labels"
}

// IssueCycle is a join table for issue-cycle associations.
type IssueCycle struct {
	IssueID uint64 `gorm:"primaryKey;autoIncrement:false" json:"issue_id"`
	CycleID uint64 `gorm:"primaryKey;autoIncrement:false" json:"cycle_id"`

	Issue Issue `gorm:"foreignKey:IssueID;constraint:OnDelete:CASCADE" json:"-"`
	Cycle Cycle `gorm:"foreignKey:CycleID;constraint:OnDelete:CASCADE" json:"-"`
}

func (IssueCycle) TableName() string {
	return "issue_cycles"
}

// IssueActivity tracks changes to an issue.
type IssueActivity struct {
	BaseModel

	IssueID  *uint64 `json:"issue_id"`
	Verb     string  `gorm:"size:255;default:created" json:"verb"`
	Field    *string `gorm:"size:255" json:"field"`
	OldValue *string `gorm:"type:text" json:"old_value"`
	NewValue *string `gorm:"type:text" json:"new_value"`
	Comment  *string `gorm:"type:text" json:"comment"`
	ActorID  *uint64 `json:"actor_id"`

	Issue *Issue `gorm:"foreignKey:IssueID" json:"-"`
}

func (IssueActivity) TableName() string {
	return "issue_activities"
}
