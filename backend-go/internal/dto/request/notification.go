package request

// NotificationCreateRequest represents the request to create a notification.
type NotificationCreateRequest struct {
	Title       string  `json:"title" binding:"required"`
	Message     string  `json:"message" binding:"required"`
	Type        string  `json:"type"`
	Priority    string  `json:"priority"`
	ActionURL   *string `json:"action_url"`
	RecipientID uint64  `json:"recipient_id" binding:"required"`
	SenderID    *uint64 `json:"sender_id"`
	ProjectID   *uint64 `json:"project_id"`
	IssueID     *uint64 `json:"issue_id"`
}

// NotificationBulkCreateRequest represents a bulk notification creation request.
type NotificationBulkCreateRequest struct {
	Title        string   `json:"title" binding:"required"`
	Message      string   `json:"message" binding:"required"`
	Type         string   `json:"type"`
	Priority     string   `json:"priority"`
	ActionURL    *string  `json:"action_url"`
	SenderID     *uint64  `json:"sender_id"`
	ProjectID    *uint64  `json:"project_id"`
	IssueID      *uint64  `json:"issue_id"`
	RecipientIDs []uint64 `json:"recipient_ids" binding:"required"`
}
