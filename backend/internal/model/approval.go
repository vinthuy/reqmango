package model

import "time"

// Approval represents a single approval request on an issue.
type Approval struct {
	BaseModel
	IssueID              uint64     `gorm:"index;not null" json:"issue_id"`
	WorkflowID           uint64     `gorm:"not null" json:"workflow_id"`
	TransitionID         uint64     `gorm:"index;not null" json:"transition_id"`
	ProjectID            uint64     `gorm:"index;not null" json:"project_id"`
	WorkspaceID          uint64     `gorm:"index;not null" json:"workspace_id"`
	RequesterID          uint64     `gorm:"not null" json:"requester_id"`
	RequestNote          string     `gorm:"type:text" json:"request_note"`
	SourceStateID        uint64     `gorm:"not null" json:"source_state_id"`
	ApproveTargetStateID uint64     `gorm:"not null" json:"approve_target_state_id"`
	RejectTargetStateID  uint64     `gorm:"not null" json:"reject_target_state_id"`
	ApproverIDs          string     `gorm:"type:jsonb;not null;default:'[]'" json:"approver_ids"`
	Status               string     `gorm:"type:varchar(20);not null;default:'pending';index" json:"status"`
	DecidedBy            *uint64    `json:"decided_by"`
	DecidedAt            *time.Time `json:"decided_at"`
	DecisionNote         string     `gorm:"type:text" json:"decision_note"`
}

func (Approval) TableName() string { return "approvals" }

// ApprovalRecord records each approver's individual decision on an approval.
type ApprovalRecord struct {
	BaseModel
	ApprovalID uint64    `gorm:"index;not null" json:"approval_id"`
	ApproverID uint64    `gorm:"index;not null" json:"approver_id"`
	Decision   string    `gorm:"type:varchar(20);not null" json:"decision"`
	Note       string    `gorm:"type:text" json:"note"`
	DecidedAt  time.Time `gorm:"not null" json:"decided_at"`
}

func (ApprovalRecord) TableName() string { return "approval_records" }
