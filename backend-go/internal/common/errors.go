package common

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/reqmanpy/backend-go/internal/i18n"
)

// AppError represents an application-level error with an HTTP status and business error code.
type AppError struct {
	Code      int       `json:"-"` // HTTP status code (backward compat)
	ErrorCode ErrorCode `json:"error_code"`
	Message   string    `json:"message"`
	Detail    string    `json:"detail,omitempty"`
}

// HTTPCode returns the HTTP status code.
func (e *AppError) HTTPCodeVal() int { return e.Code }

func (e *AppError) Error() string { return e.Message }

// ==================== Error Constructors ====================

func NewError(code ErrorCode, msg string) *AppError {
	return &AppError{Code: code.HTTPStatus(), ErrorCode: code, Message: msg}
}

func NewErrorDetail(code ErrorCode, msg, detail string) *AppError {
	return &AppError{Code: code.HTTPStatus(), ErrorCode: code, Message: msg, Detail: detail}
}

func NotFound(msg string) *AppError          { return NewError(ErrNotFound, msg) }
func Conflict(msg string) *AppError          { return NewError(ErrConflict, msg) }
func Unauthorized(msg string) *AppError      { return NewError(ErrUnauthorized, msg) }
func Forbidden(msg string) *AppError         { return NewError(ErrForbidden, msg) }
func Validation(msg string) *AppError        { return NewError(ErrValidation, msg) }
func Internal(msg string) *AppError          { return NewError(ErrInternal, msg) }
func BadRequest(msg string) *AppError        { return NewError(ErrBadRequest, msg) }

// ProjectNotFound etc.
func ProjectNotFound() *AppError    { return NewError(ErrProjectNotFound, "Project not found") }
func IssueNotFound() *AppError      { return NewError(ErrIssueNotFound, "Issue not found") }
func UserNotFound() *AppError       { return NewError(ErrUserNotFound, "User not found") }
func WorkspaceNotFound() *AppError  { return NewError(ErrWorkspaceNotFound, "Workspace not found") }
func PageNotFound() *AppError       { return NewError(ErrPageNotFound, "Page not found") }
func CycleNotFound() *AppError      { return NewError(ErrCycleNotFound, "Cycle not found") }
func ModuleNotFound() *AppError     { return NewError(ErrModuleNotFound, "Module not found") }
func StateNotFound() *AppError      { return NewError(ErrStateNotFound, "State not found") }
func LabelNotFound() *AppError      { return NewError(ErrLabelNotFound, "Label not found") }
func ReleaseNotFound() *AppError    { return NewError(ErrReleaseNotFound, "Release not found") }
func TemplateNotFound() *AppError   { return NewError(ErrTemplateNotFound, "Template not found") }
func ViewNotFound() *AppError       { return NewError(ErrViewNotFound, "Saved view not found") }
func CommentNotFound() *AppError    { return NewError(ErrCommentNotFound, "Comment not found") }
func AttachmentNotFound() *AppError { return NewError(ErrAttachmentNotFound, "Attachment not found") }
func TimeEntryNotFound() *AppError  { return NewError(ErrTimeEntryNotFound, "Time entry not found") }
func RecurrenceNotFound() *AppError { return NewError(ErrRecurrenceNotFound, "Recurrence rule not found") }
func NotificationNotFound() *AppError { return NewError(ErrNotificationNotFound, "Notification not found") }
func AgentNotFound() *AppError          { return NewError(ErrAgentNotFound, "Agent not found") }

// ==================== Standard Response Helpers ====================

// RespondError writes a standardized error response.
func RespondError(c *gin.Context, err error) {
	if ae, ok := err.(*AppError); ok {
		lang := getLang(c)
		if msg := i18n.GetMessage(lang, strings.ToLower(string(ae.ErrorCode))); msg != "" {
			ae.Message = msg
		}
		c.JSON(ae.Code, ae)
		return
	}
	c.JSON(http.StatusInternalServerError, AppError{
		Code:  500,
		ErrorCode: ErrInternal,
		Message:   "Internal server error",
	})
}

func getLang(c *gin.Context) string {
	if l, exists := c.Get("lang"); exists { return l.(string) }
	return "zh"
}

// RespondOK writes a success response.
func RespondOK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

// RespondCreated writes a 201 response.
func RespondCreated(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, data)
}
