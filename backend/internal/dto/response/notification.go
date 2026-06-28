package response

import "time"

// NotificationResponse is the API response for a notification.
type NotificationResponse struct {
	ID          uint64     `json:"id"`
	Title       string     `json:"title"`
	Message     string     `json:"message"`
	Type        string     `json:"type"`
	Priority    string     `json:"priority"`
	IsRead      bool       `json:"is_read"`
	ReadAt      *string    `json:"read_at"`
	ActionURL   *string    `json:"action_url"`
	RecipientID uint64     `json:"recipient_id"`
	SenderID    *uint64    `json:"sender_id"`
	ProjectID   *uint64    `json:"project_id"`
	IssueID     *uint64    `json:"issue_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// NotificationSummaryResponse is the API response for notification counts.
type NotificationSummaryResponse struct {
	Total        int            `json:"total"`
	Unread       int            `json:"unread"`
	UnreadByType map[string]int `json:"unread_by_type"`
}

// MarkAllReadResponse is the response after marking all notifications read.
type MarkAllReadResponse struct {
	MarkedCount int64 `json:"marked_count"`
}
