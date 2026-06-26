package common

// ErrorCode is a machine-readable business error code.
type ErrorCode string

// Standard error codes. Format: CATEGORY_REASON
const (
	// General
	ErrInternal     ErrorCode = "INTERNAL_ERROR"
	ErrBadRequest   ErrorCode = "BAD_REQUEST"
	ErrValidation   ErrorCode = "VALIDATION_ERROR"
	ErrUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrForbidden    ErrorCode = "FORBIDDEN"

	// Resource errors
	ErrNotFound       ErrorCode = "NOT_FOUND"
	ErrAlreadyExists  ErrorCode = "ALREADY_EXISTS"
	ErrConflict       ErrorCode = "CONFLICT"

	// Specific resource not found
	ErrProjectNotFound   ErrorCode = "PROJECT_NOT_FOUND"
	ErrIssueNotFound     ErrorCode = "ISSUE_NOT_FOUND"
	ErrUserNotFound      ErrorCode = "USER_NOT_FOUND"
	ErrWorkspaceNotFound ErrorCode = "WORKSPACE_NOT_FOUND"
	ErrPageNotFound      ErrorCode = "PAGE_NOT_FOUND"
	ErrCycleNotFound     ErrorCode = "CYCLE_NOT_FOUND"
	ErrModuleNotFound    ErrorCode = "MODULE_NOT_FOUND"
	ErrStateNotFound     ErrorCode = "STATE_NOT_FOUND"
	ErrLabelNotFound     ErrorCode = "LABEL_NOT_FOUND"
	ErrReleaseNotFound   ErrorCode = "RELEASE_NOT_FOUND"
	ErrTemplateNotFound  ErrorCode = "TEMPLATE_NOT_FOUND"
	ErrViewNotFound      ErrorCode = "VIEW_NOT_FOUND"
	ErrCommentNotFound   ErrorCode = "COMMENT_NOT_FOUND"
	ErrAttachmentNotFound ErrorCode = "ATTACHMENT_NOT_FOUND"
	ErrTimeEntryNotFound ErrorCode = "TIME_ENTRY_NOT_FOUND"
	ErrRecurrenceNotFound ErrorCode = "RECURRENCE_NOT_FOUND"
	ErrNotificationNotFound ErrorCode = "NOTIFICATION_NOT_FOUND"

	// Validation errors
	ErrRequiredField  ErrorCode = "REQUIRED_FIELD"
	ErrInvalidFormat  ErrorCode = "INVALID_FORMAT"
	ErrInvalidValue   ErrorCode = "INVALID_VALUE"
	ErrMaxDepthExceeded ErrorCode = "MAX_DEPTH_EXCEEDED"
	ErrSelfReference  ErrorCode = "SELF_REFERENCE"

	// Business errors
	ErrAlreadyAssigned  ErrorCode = "ALREADY_ASSIGNED"
	ErrAlreadyLabelled  ErrorCode = "ALREADY_LABELLED"
	ErrStateTransition  ErrorCode = "INVALID_STATE_TRANSITION"
	ErrWorkflowViolation ErrorCode = "WORKFLOW_VIOLATION"
	ErrDuplicateEntry   ErrorCode = "DUPLICATE_ENTRY"
	ErrTimerRunning     ErrorCode = "TIMER_ALREADY_RUNNING"

	// AI errors
	ErrAIConfigMissing ErrorCode = "AI_CONFIG_MISSING"
	ErrAIAPIFailure    ErrorCode = "AI_API_FAILURE"
	ErrAITimeout       ErrorCode = "AI_TIMEOUT"
	ErrAIQuotaExceeded ErrorCode = "AI_QUOTA_EXCEEDED"
)

// StatusCode maps error codes to HTTP status codes.
func (e ErrorCode) HTTPStatus() int {
	switch {
	case e == ErrUnauthorized:
		return 401
	case e == ErrForbidden:
		return 403
	case e == ErrNotFound || stringsHasSuffix(string(e), "_NOT_FOUND"):
		return 404
	case e == ErrAlreadyExists || e == ErrConflict || e == ErrDuplicateEntry:
		return 409
	case e == ErrValidation || e == ErrRequiredField || e == ErrInvalidFormat || e == ErrInvalidValue:
		return 422
	case e == ErrBadRequest:
		return 400
	case e == ErrMaxDepthExceeded || e == ErrSelfReference:
		return 422
	case e == ErrStateTransition || e == ErrWorkflowViolation:
		return 422
	case e == ErrAIAPIFailure || e == ErrAITimeout:
		return 502
	case e == ErrAIQuotaExceeded:
		return 429
	default:
		return 500
	}
}

func stringsHasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}
