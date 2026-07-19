package response

import "time"

type ApprovalResponse struct {
	ID                   uint64     `json:"id"`
	IssueID              uint64     `json:"issue_id"`
	WorkflowID           uint64     `json:"workflow_id"`
	TransitionID         uint64     `json:"transition_id"`
	ProjectID            uint64     `json:"project_id"`
	WorkspaceID          uint64     `json:"workspace_id"`
	RequesterID          uint64     `json:"requester_id"`
	RequesterName        string     `json:"requester_name"`
	RequestNote          string     `json:"request_note"`
	SourceStateID        uint64     `json:"source_state_id"`
	SourceStateName      string     `json:"source_state_name"`
	ApproveTargetStateID uint64     `json:"approve_target_state_id"`
	ApproveStateName     string     `json:"approve_target_state_name"`
	RejectTargetStateID  uint64     `json:"reject_target_state_id"`
	RejectStateName      string     `json:"reject_target_state_name"`
	ApproverIDs          []uint64   `json:"approver_ids"`
	ApproverNames        []string   `json:"approver_names"`
	Status               string     `json:"status"`
	DecidedBy            *uint64    `json:"decided_by"`
	DecidedByName        string     `json:"decided_by_name"`
	DecidedAt            *time.Time `json:"decided_at"`
	DecisionNote         string     `json:"decision_note"`
	CreatedAt            time.Time  `json:"created_at"`

	IssueKey    string `json:"issue_key"`
	IssueTitle  string `json:"issue_title"`
	ProjectName string `json:"project_name"`

	Records []ApprovalRecordResponse `json:"records"`
}

type ApprovalRecordResponse struct {
	ID           uint64    `json:"id"`
	ApproverID   uint64    `json:"approver_id"`
	ApproverName string    `json:"approver_name"`
	Decision     string    `json:"decision"`
	Note         string    `json:"note"`
	DecidedAt    time.Time `json:"decided_at"`
}

type ApprovalCountResponse struct {
	PendingCount int64 `json:"pending_count"`
}
