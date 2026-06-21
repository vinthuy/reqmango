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

type IssueRelationResponse struct {
	ID             uint64  `json:"id"`
	IssueID        uint64  `json:"issue_id"`
	RelatedIssueID uint64  `json:"related_issue_id"`
	RelationTypeID uint64  `json:"relation_type_id"`
	Comment        *string `json:"comment"`
	// Populated relations
	RelationName   string  `json:"relation_name,omitempty"`
	InwardName     string  `json:"inward_name,omitempty"`
	OutwardName    string  `json:"outward_name,omitempty"`
	RelatedName    string  `json:"related_name,omitempty"`
	RelatedSeqID   int     `json:"related_seq_id,omitempty"`
	RelatedProject string  `json:"related_project,omitempty"`
}
