package response

import "time"

// CycleResponse is the full cycle representation returned by the API.
// NOTE: Cycle dates use "2006-01-02" (date-only) format strings rather than
// *time.Time (RFC3339). This is intentional — cycles are date-range-oriented,
// so the time-of-day component is irrelevant. Issue dates (*time.Time / RFC3339)
// carry full timestamps because they represent point-in-time events.
type CycleResponse struct {
	ID              uint64       `json:"id"`
	Name            string       `json:"name"`
	Description     *string      `json:"description"`
	Status          string       `json:"status"`           // computed: upcoming|active|completed|cancelled
	Progress        float64      `json:"progress"`          // 0-100
	TotalIssues     int64        `json:"total_issues"`
	CompletedIssues int64        `json:"completed_issues"`
	StartDate       string       `json:"start_date"`        // "2006-01-02"
	EndDate         *string      `json:"end_date"`          // "2006-01-02", nullable
	ProjectID       uint64       `json:"project_id"`
	WorkspaceID     uint64       `json:"workspace_id"`
	OwnedBy         *UserLite    `json:"owned_by"`
	Project         *ProjectLite `json:"project"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
	CreatedByID     *uint64      `json:"created_by_id"`
	UpdatedByID     *uint64      `json:"updated_by_id"`
}

// CycleLite is a compact cycle representation.
type CycleLite struct {
	ID        uint64  `json:"id"`
	Name      string  `json:"name"`
	StartDate string `json:"start_date"`
	EndDate   *string `json:"end_date"`
}

// CycleProgress represents cycle progress statistics.
type CycleProgress struct {
	CycleID         uint64           `json:"cycle_id"`
	CycleName       string           `json:"cycle_name"`
	TotalIssues     int64            `json:"total_issues"`
	CompletedIssues int64            `json:"completed_issues"`
	Progress        float64          `json:"progress"`
	StateBreakdown  []StateBreakdown `json:"state_breakdown"`
}

// StateBreakdown shows issue count per state.
type StateBreakdown struct {
	State string `json:"state"`
	Group string `json:"group"`
	Count int64  `json:"count"`
}

// CycleStatistics extends progress with priority breakdown and date info.
type CycleStatistics struct {
	CycleProgress
	PriorityBreakdown map[string]int64 `json:"priority_breakdown"`
	IssueStats        IssueStats       `json:"issue_stats"`
	DateRange         DateRange        `json:"date_range"`
}

// IssueStats represents issue statistics for date fields.
type IssueStats struct {
	Total          int64 `json:"total"`
	WithStartDate  int64 `json:"with_start_date"`
	WithTargetDate int64 `json:"with_target_date"`
}

// DateRange represents a date range.
type DateRange struct {
	StartDate string `json:"start_date"`
	EndDate   *string `json:"end_date"`
}

// BurndownData represents burndown chart data.
type BurndownData struct {
	CycleID         uint64  `json:"cycle_id"`
	CycleName       string  `json:"cycle_name"`
	StartDate       string  `json:"start_date"`
	EndDate         string  `json:"end_date"`
	TotalIssues     int64   `json:"total_issues"`
	TotalDays       int     `json:"total_days"`
	DaysElapsed     int     `json:"days_elapsed"`
	IdealDailyBurn  float64 `json:"ideal_daily_burn"`
	IdealRemaining  float64 `json:"ideal_remaining"`
	ActualCompleted int64   `json:"actual_completed"`
	ActualRemaining int64   `json:"actual_remaining"`
	IsOnTrack       bool    `json:"is_on_track"`
}
