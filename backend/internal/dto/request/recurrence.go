package request

// RecurrenceCreateRequest creates a recurrence rule.
type RecurrenceCreateRequest struct {
	Frequency string  `json:"frequency" binding:"required"` // daily | weekly | monthly | cron
	Interval  int     `json:"interval"`
	CronExpr  *string `json:"cron_expr"`
	NextRun   string  `json:"next_run"` // RFC3339
	EndDate   *string `json:"end_date"`
}

// RecurrenceUpdateRequest updates a recurrence rule.
type RecurrenceUpdateRequest struct {
	Frequency *string `json:"frequency"`
	Interval  *int    `json:"interval"`
	CronExpr  *string `json:"cron_expr"`
	NextRun   *string `json:"next_run"`
	EndDate   *string `json:"end_date"`
	IsActive  *bool   `json:"is_active"`
}
