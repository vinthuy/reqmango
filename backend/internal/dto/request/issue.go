package request

// IssueCreateRequest is the request body for creating an issue.
type IssueCreateRequest struct {
	Name              string                 `json:"name" binding:"required,min=1,max=255"`
	DescriptionHTML   string                 `json:"description_html"`
	DescriptionJSON   *string                `json:"description_json"`
	Priority          string                 `json:"priority"`
	StateID           *uint64                `json:"state_id"`
	AssigneeIDs       []uint64               `json:"assignee_ids"`
	LabelIDs          []uint64               `json:"label_ids"`
	StartDate         *string                `json:"start_date"`  // RFC3339 date
	TargetDate        *string                `json:"target_date"` // RFC3339 date
	ParentID          *uint64                `json:"parent_id"`
	EstimatePointID   *uint64                `json:"estimate_point_id"`
	TypeID            *uint64                `json:"type_id"`
	ExternalID        *string                `json:"external_id"`
	ExternalSource    *string                `json:"external_source"`
	CoverImageURL     *string                `json:"cover_image_url"`
	CustomFieldValues map[uint64]interface{} `json:"custom_field_values"`
}

// IssueUpdateRequest is the request body for updating an issue.
type IssueUpdateRequest struct {
	Name            *string  `json:"name"`
	DescriptionHTML *string  `json:"description_html"`
	DescriptionJSON *string  `json:"description_json"`
	Priority        *string  `json:"priority"`
	StateID         *uint64  `json:"state_id"`
	AssigneeIDs     []uint64 `json:"assignee_ids"`
	LabelIDs        []uint64 `json:"label_ids"`
	StartDate       *string  `json:"start_date"`
	TargetDate      *string  `json:"target_date"`
	EstimatePointID *uint64  `json:"estimate_point_id"`
	CycleID         *uint64  `json:"cycle_id"`
	ModuleIDs       []uint64 `json:"module_ids"`
	ParentID        *uint64  `json:"parent_id"`
	TypeID          *uint64  `json:"type_id"`
	CoverImageURL   *string  `json:"cover_image_url"`
}

// BulkUpdateRequest is the request body for bulk issue updates.
type BulkUpdateRequest struct {
	IssueIDs    []uint64 `json:"issue_ids" binding:"required"`
	Priority    *string  `json:"priority"`
	StateID     *uint64  `json:"state_id"`
	AssigneeIDs []uint64 `json:"assignee_ids"`
	LabelIDs    []uint64 `json:"label_ids"`
	StartDate   *string  `json:"start_date"`
	TargetDate  *string  `json:"target_date"`
}

// BulkDeleteRequest is the request body for bulk issue deletion.
type BulkDeleteRequest struct {
	IssueIDs []uint64 `json:"issue_ids" binding:"required"`
}

// BulkCopyRequest is the request body for bulk copying issues to another project.
type BulkCopyRequest struct {
	IssueIDs        []uint64 `json:"issue_ids" binding:"required"`
	TargetProjectID uint64   `json:"target_project_id" binding:"required"`
	IncludeSubtasks bool     `json:"include_subtasks"`
}

// BulkMoveRequest is the request body for bulk moving issues to another project.
type BulkMoveRequest struct {
	IssueIDs        []uint64 `json:"issue_ids" binding:"required"`
	TargetProjectID uint64   `json:"target_project_id" binding:"required"`
	IncludeSubtasks bool     `json:"include_subtasks"`
}

// ConvertTypeRequest is the request body for converting issue type.
type ConvertTypeRequest struct {
	TargetTypeID uint64 `json:"target_type_id" binding:"required"`
}

// MergeDuplicatesRequest is the request body for merging duplicate issues.
type MergeDuplicatesRequest struct {
	TargetIssueID       uint64   `json:"target_issue_id" binding:"required"`
	SourceIssueIDs      []uint64 `json:"source_issue_ids" binding:"required"`
	KeepSourceComments  bool     `json:"keep_source_comments"`
	KeepSourceLabels    bool     `json:"keep_source_labels"`
	KeepSourceAssignees bool     `json:"keep_source_assignees"`
}

// ImportIssueItem represents a single issue item for import (JSON format).
type ImportIssueItem struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Priority       string   `json:"priority"`
	StateName      string   `json:"state_name"`
	TypeName       string   `json:"type_name"`
	AssigneeEmails []string `json:"assignee_emails"`
	LabelNames     []string `json:"label_names"`
	StartDate      string   `json:"start_date"`
	TargetDate     string   `json:"target_date"`
	ParentTitle    string   `json:"parent_title"`
	ModuleName     string   `json:"module_name"`
	CycleName      string   `json:"cycle_name"`
	EstimatePoints *float64 `json:"estimate_points"`
}

// CSVImportRow represents a row parsed from CSV.
type CSVImportRow map[string]string
