package common

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupTestContext() *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	return c
}

func TestNewError_Basic(t *testing.T) {
	err := NewError(ErrBadRequest, "Bad input")
	if err.Code != 400 {
		t.Errorf("Code = %d, want 400", err.Code)
	}
	if err.ErrorCode != ErrBadRequest {
		t.Errorf("ErrorCode = %s, want %s", err.ErrorCode, ErrBadRequest)
	}
	if err.Message != "Bad input" {
		t.Errorf("Message = %s, want 'Bad input'", err.Message)
	}
	if err.Detail != "" {
		t.Errorf("Detail should be empty, got %s", err.Detail)
	}
}

func TestNewErrorDetail(t *testing.T) {
	err := NewErrorDetail(ErrValidation, "Invalid request", "email is required")
	if err.Code != 422 {
		t.Errorf("Code = %d, want 422", err.Code)
	}
	if err.Detail != "email is required" {
		t.Errorf("Detail = %s, want 'email is required'", err.Detail)
	}
}

func TestError_Methods(t *testing.T) {
	err := NewError(ErrNotFound, "Not found")
	if err.Error() != "Not found" {
		t.Errorf("Error() = %s, want 'Not found'", err.Error())
	}
	if err.HTTPCodeVal() != 404 {
		t.Errorf("HTTPCodeVal() = %d, want 404", err.HTTPCodeVal())
	}
}

func TestAppError_JSONSerialization(t *testing.T) {
	err := NewErrorDetail(ErrValidation, "Invalid", "field x is required")
	data, jsonErr := json.Marshal(err)
	if jsonErr != nil {
		t.Fatalf("json.Marshal failed: %v", jsonErr)
	}

	var parsed map[string]interface{}
	if ue := json.Unmarshal(data, &parsed); ue != nil {
		t.Fatalf("json.Unmarshal failed: %v", ue)
	}
	if parsed["error_code"] != "VALIDATION_ERROR" {
		t.Errorf("error_code = %v, want VALIDATION_ERROR", parsed["error_code"])
	}
	if parsed["message"] != "Invalid" {
		t.Errorf("message = %v, want Invalid", parsed["message"])
	}
	if parsed["detail"] != "field x is required" {
		t.Errorf("detail = %v, want 'field x is required'", parsed["detail"])
	}
	// Code should be omitted (json:"-")
	if _, ok := parsed["Code"]; ok {
		t.Errorf("Code should be omitted in JSON (json:- tag)")
	}
}

func TestConvenienceConstructors(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(string) *AppError
		code     ErrorCode
		httpCode int
		msg      string
	}{
		{"NotFound", NotFound, ErrNotFound, 404, "test"},
		{"Conflict", Conflict, ErrConflict, 409, "test"},
		{"Unauthorized", Unauthorized, ErrUnauthorized, 401, "test"},
		{"Forbidden", Forbidden, ErrForbidden, 403, "test"},
		{"Validation", Validation, ErrValidation, 422, "test"},
		{"Internal", Internal, ErrInternal, 500, "test"},
		{"BadRequest", BadRequest, ErrBadRequest, 400, "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn(tt.msg)
			if err.ErrorCode != tt.code {
				t.Errorf("ErrorCode = %s, want %s", err.ErrorCode, tt.code)
			}
			if err.Code != tt.httpCode {
				t.Errorf("HTTP Code = %d, want %d", err.Code, tt.httpCode)
			}
			if err.Message != tt.msg {
				t.Errorf("Message = %s, want %s", err.Message, tt.msg)
			}
		})
	}
}

func TestSpecificNotFoundConstructors(t *testing.T) {
	tests := []struct {
		name     string
		fn       func() *AppError
		code     ErrorCode
		expected string
	}{
		{"ProjectNotFound", ProjectNotFound, ErrProjectNotFound, "Project not found"},
		{"IssueNotFound", IssueNotFound, ErrIssueNotFound, "Issue not found"},
		{"UserNotFound", UserNotFound, ErrUserNotFound, "User not found"},
		{"WorkspaceNotFound", WorkspaceNotFound, ErrWorkspaceNotFound, "Workspace not found"},
		{"PageNotFound", PageNotFound, ErrPageNotFound, "Page not found"},
		{"CycleNotFound", CycleNotFound, ErrCycleNotFound, "Cycle not found"},
		{"ModuleNotFound", ModuleNotFound, ErrModuleNotFound, "Module not found"},
		{"StateNotFound", StateNotFound, ErrStateNotFound, "State not found"},
		{"LabelNotFound", LabelNotFound, ErrLabelNotFound, "Label not found"},
		{"ReleaseNotFound", ReleaseNotFound, ErrReleaseNotFound, "Release not found"},
		{"TemplateNotFound", TemplateNotFound, ErrTemplateNotFound, "Template not found"},
		{"ViewNotFound", ViewNotFound, ErrViewNotFound, "Saved view not found"},
		{"CommentNotFound", CommentNotFound, ErrCommentNotFound, "Comment not found"},
		{"AttachmentNotFound", AttachmentNotFound, ErrAttachmentNotFound, "Attachment not found"},
		{"TimeEntryNotFound", TimeEntryNotFound, ErrTimeEntryNotFound, "Time entry not found"},
		{"RecurrenceNotFound", RecurrenceNotFound, ErrRecurrenceNotFound, "Recurrence rule not found"},
		{"NotificationNotFound", NotificationNotFound, ErrNotificationNotFound, "Notification not found"},
		{"AgentNotFound", AgentNotFound, ErrAgentNotFound, "Agent not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			if err.ErrorCode != tt.code {
				t.Errorf("ErrorCode = %s, want %s", err.ErrorCode, tt.code)
			}
			if err.Code != 404 {
				t.Errorf("HTTP Code = %d, want 404", err.Code)
			}
			if err.Message != tt.expected {
				t.Errorf("Message = %s, want %s", err.Message, tt.expected)
			}
		})
	}
}

func TestRespondError_AppError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	appErr := NewError(ErrBadRequest, "Something bad")
	RespondError(c, appErr)

	if w.Code != 400 {
		t.Errorf("status = %d, want 400", w.Code)
	}

	var body AppError
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if body.ErrorCode != ErrBadRequest {
		t.Errorf("ErrorCode = %s, want %s", body.ErrorCode, ErrBadRequest)
	}
}

func TestRespondError_GenericError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	// Simulate a generic error (not *AppError)
	RespondError(c, http.ErrBodyNotAllowed)

	if w.Code != 500 {
		t.Errorf("status = %d, want 500", w.Code)
	}

	var body AppError
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if body.ErrorCode != ErrInternal {
		t.Errorf("ErrorCode = %s, want %s", body.ErrorCode, ErrInternal)
	}
}

func TestRespondOK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	data := map[string]string{"status": "ok"}
	RespondOK(c, data)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestRespondCreated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/test", nil)

	data := map[string]int{"id": 42}
	RespondCreated(c, data)

	if w.Code != 201 {
		t.Errorf("status = %d, want 201", w.Code)
	}
}

func TestRespondError_WithLang(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Set("lang", "zh")

	appErr := NewError(ErrNotFound, "Not found")
	RespondError(c, appErr)

	if w.Code != 404 {
		t.Errorf("status = %d, want 404", w.Code)
	}
	// Response should still be valid JSON
	var body AppError
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
}
