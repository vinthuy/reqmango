package request

// PageLockRequest locks a page for editing.
type PageLockRequest struct{}

// PageConvertRequest converts a page to an issue.
type PageConvertRequest struct {
	IssueTypeID *uint64 `json:"issue_type_id"`
}
