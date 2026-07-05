package response

import "time"

type RelationTypeResponse struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	InwardName  string    `json:"inward_name"`
	OutwardName string    `json:"outward_name"`
	WorkspaceID uint64    `json:"workspace_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RelationTypeLite struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	OutwardName string `json:"outward_name"`
}

type IssueRelationResponse struct {
	ID             uint64            `json:"id"`
	IssueID        uint64            `json:"issue_id"`
	RelatedIssueID uint64            `json:"related_issue_id"`
	RelationTypeID uint64            `json:"relation_type_id"`
	Comment        *string           `json:"comment"`
	Direction      string            `json:"direction"` // "outbound" (I relate to others) or "inbound" (others relate to me)
	RelationType   *RelationTypeLite `json:"relation_type,omitempty"`
	RelatedIssue   *RelatedIssueLite `json:"related_issue,omitempty"`
	// Legacy flat fields (keep for backward compat)
	RelationName   string `json:"relation_name,omitempty"`
	InwardName     string `json:"inward_name,omitempty"`
	OutwardName    string `json:"outward_name,omitempty"`
	RelatedName    string `json:"related_name,omitempty"`
	RelatedSeqID   int    `json:"related_seq_id,omitempty"`
	RelatedProject string `json:"related_project,omitempty"`
}
