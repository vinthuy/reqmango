package response

import "time"

// TimeTrackResponse is the API response for a time entry.
type TimeTrackResponse struct {
	ID          uint64     `json:"id"`
	IssueID     uint64     `json:"issue_id"`
	UserID      uint64     `json:"user_id"`
	Description *string    `json:"description"`
	StartedAt   time.Time  `json:"started_at"`
	EndedAt     *time.Time `json:"ended_at"`
	Duration    int64      `json:"duration"` // seconds
	CreatedAt   time.Time  `json:"created_at"`
}

// TimeTrackSummary is the summary response for an issue.
type TimeTrackSummary struct {
	IssueID      uint64  `json:"issue_id"`
	TotalSeconds int64   `json:"total_seconds"`
	TotalHours   float64 `json:"total_hours"`
	EntryCount   int     `json:"entry_count"`
	IsRunning    bool    `json:"is_running"`
}
