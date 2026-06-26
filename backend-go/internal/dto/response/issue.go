package response

import "time"

// IssueResponse is the full issue representation.
type IssueResponse struct {
	ID                uint64     `json:"id"`
	Name              string     `json:"name"`
	DescriptionHTML   string     `json:"description_html"`
	DescriptionJSON   *string    `json:"description_json"`
	Priority          string     `json:"priority"`
	SequenceID        int        `json:"sequence_id"`
	SortOrder         float64    `json:"sort_order"`
	StartDate         *time.Time `json:"start_date"`
	TargetDate        *time.Time `json:"target_date"`
	CompletedAt       *time.Time `json:"completed_at"`
	IsDraft           bool       `json:"is_draft"`
	ArchivedAt        *time.Time `json:"archived_at"`

	ProjectID         uint64        `json:"project_id"`
	Project           *ProjectLite  `json:"project"`
	WorkspaceID       uint64        `json:"workspace_id"`
	StateID           uint64        `json:"state_id"`
	StateName         string        `json:"state_name"`
	StateGroup        string        `json:"state_group"`

	ParentID          *uint64       `json:"parent_id"`
	Depth             int           `json:"depth"`
	Assignees         []UserLite    `json:"assignees"`
	Labels            []uint64      `json:"labels"`
	LabelDetails      []LabelLite   `json:"label_details"`
	SubIssuesCount    int64         `json:"sub_issues_count"`
	LinkCount         int           `json:"link_count"`
	AttachmentCount   int           `json:"attachment_count"`

	EstimatePointID   *uint64       `json:"estimate_point_id"`
	CycleID           *uint64       `json:"cycle_id"`
	ModuleIDs         []uint64      `json:"module_ids"`
	ExternalID        *string       `json:"external_id"`
	ExternalSource    *string       `json:"external_source"`
	CoverImageURL     *string       `json:"cover_image_url"`

	IssueType         *IssueTypeLite `json:"issue_type,omitempty"`

	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	CreatedByID       *uint64        `json:"created_by_id"`
	UpdatedByID       *uint64        `json:"updated_by_id"`
	DeletedAt         *time.Time     `json:"deleted_at"`
	IsDeleted         bool           `json:"is_deleted"`
}

// IssueLite is a compact issue representation.
type IssueLite struct {
	ID                uint64 `json:"id"`
	Name              string `json:"name"`
	SequenceID        int    `json:"sequence_id"`
	Priority          string `json:"priority"`
	StateID           uint64 `json:"state_id"`
	ProjectID         uint64 `json:"project_id"`
	ProjectIdentifier string `json:"project_identifier"`
}

// IssueActivityResponse represents an issue activity entry.
type IssueActivityResponse struct {
	ID        uint64     `json:"id"`
	IssueID   *uint64    `json:"issue_id"`
	Verb      string     `json:"verb"`
	Field     *string    `json:"field"`
	OldValue  *string    `json:"old_value"`
	NewValue  *string    `json:"new_value"`
	Comment   *string    `json:"comment"`
	ActorID   *uint64    `json:"actor_id"`
	CreatedAt time.Time  `json:"created_at"`
}

// IssueSearchResult represents a search result.
type IssueSearchResult struct {
	ID                uint64 `json:"id"`
	Name              string `json:"name"`
	SequenceID        int    `json:"sequence_id"`
	ProjectIdentifier string `json:"project_identifier"`
	ProjectID         uint64 `json:"project_id"`
	WorkspaceSlug     string `json:"workspace_slug"`
}

// IssueStatistics represents issue statistics for a project.
type IssueStatistics struct {
	Total          int64            `json:"total"`
	ByState        map[string]int64 `json:"by_state"`
	ByPriority     map[string]int64 `json:"by_priority"`
	CompletedCount int64            `json:"completed_count"`
	DraftCount     int64            `json:"draft_count"`
}

// LabelLite is a compact label representation used in issue responses.
type LabelLite struct {
	ID    uint64 `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// ProjectLite is a compact project representation.
type ProjectLite struct {
	ID         uint64 `json:"id"`
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
}

// ImportResult is the result of a bulk import operation.
type ImportResult struct {
	SuccessCount int            `json:"success_count"`
	FailCount    int            `json:"fail_count"`
	Errors       []ImportError  `json:"errors"`
	ImportedIDs  []uint64       `json:"imported_ids"`
}

// ImportError represents an error for a specific row in the import.
type ImportError struct {
	Row     int    `json:"row"`
	Title   string `json:"title"`
	Message string `json:"message"`
}
