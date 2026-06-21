package request

// IssueCreateRequest is the request body for creating an issue.
type IssueCreateRequest struct {
	Name             string  `json:"name" binding:"required,min=1,max=255"`
	DescriptionHTML  string  `json:"description_html"`
	DescriptionJSON  *string `json:"description_json"`
	Priority         string  `json:"priority"`
	StateID          *uint64 `json:"state_id"`
	AssigneeIDs      []uint64 `json:"assignee_ids"`
	LabelIDs         []uint64 `json:"label_ids"`
	StartDate        *string `json:"start_date"`   // RFC3339 date
	TargetDate       *string `json:"target_date"`  // RFC3339 date
	ParentID         *uint64 `json:"parent_id"`
	EstimatePointID  *uint64 `json:"estimate_point_id"`
	TypeID           *uint64 `json:"type_id"`
	ExternalID       *string `json:"external_id"`
	ExternalSource   *string `json:"external_source"`
}

// IssueUpdateRequest is the request body for updating an issue.
type IssueUpdateRequest struct {
	Name             *string  `json:"name"`
	DescriptionHTML  *string  `json:"description_html"`
	DescriptionJSON  *string  `json:"description_json"`
	Priority         *string  `json:"priority"`
	StateID          *uint64  `json:"state_id"`
	AssigneeIDs      []uint64 `json:"assignee_ids"`
	LabelIDs         []uint64 `json:"label_ids"`
	StartDate        *string  `json:"start_date"`
	TargetDate       *string  `json:"target_date"`
	EstimatePointID  *uint64  `json:"estimate_point_id"`
	CycleID          *uint64  `json:"cycle_id"`
	ModuleIDs        []uint64 `json:"module_ids"`
	ParentID         *uint64  `json:"parent_id"`
}

// BulkUpdateRequest is the request body for bulk issue updates.
type BulkUpdateRequest struct {
	IssueIDs     []uint64 `json:"issue_ids" binding:"required"`
	Priority     *string  `json:"priority"`
	StateID      *uint64  `json:"state_id"`
	AssigneeIDs  []uint64 `json:"assignee_ids"`
	LabelIDs     []uint64 `json:"label_ids"`
	StartDate    *string  `json:"start_date"`
	TargetDate   *string  `json:"target_date"`
}

// BulkDeleteRequest is the request body for bulk issue deletion.
type BulkDeleteRequest struct {
	IssueIDs []uint64 `json:"issue_ids" binding:"required"`
}
