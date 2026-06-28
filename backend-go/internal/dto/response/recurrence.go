package response

import "time"

// RecurrenceResponse is the API response for a recurrence rule.
type RecurrenceResponse struct {
	ID        uint64     `json:"id"`
	IssueID   uint64     `json:"issue_id"`
	Frequency string     `json:"frequency"`
	Interval  int        `json:"interval"`
	CronExpr  *string    `json:"cron_expr"`
	NextRun   time.Time  `json:"next_run"`
	EndDate   *time.Time `json:"end_date"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
}
