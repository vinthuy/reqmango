package common

import "testing"

func TestErrorCode_HTTPStatus(t *testing.T) {
	tests := []struct {
		name     string
		code     ErrorCode
		expected int
	}{
		// General
		{"Internal", ErrInternal, 500},
		{"BadRequest", ErrBadRequest, 400},
		{"Validation", ErrValidation, 422},
		{"Unauthorized", ErrUnauthorized, 401},
		{"Forbidden", ErrForbidden, 403},

		// Resource errors
		{"NotFound", ErrNotFound, 404},
		{"AlreadyExists", ErrAlreadyExists, 409},
		{"Conflict", ErrConflict, 409},

		// Specific not found — all map to 404
		{"ProjectNotFound", ErrProjectNotFound, 404},
		{"IssueNotFound", ErrIssueNotFound, 404},
		{"UserNotFound", ErrUserNotFound, 404},
		{"WorkspaceNotFound", ErrWorkspaceNotFound, 404},
		{"PageNotFound", ErrPageNotFound, 404},
		{"CycleNotFound", ErrCycleNotFound, 404},
		{"ModuleNotFound", ErrModuleNotFound, 404},
		{"StateNotFound", ErrStateNotFound, 404},
		{"LabelNotFound", ErrLabelNotFound, 404},
		{"ReleaseNotFound", ErrReleaseNotFound, 404},
		{"TemplateNotFound", ErrTemplateNotFound, 404},
		{"ViewNotFound", ErrViewNotFound, 404},
		{"CommentNotFound", ErrCommentNotFound, 404},
		{"AttachmentNotFound", ErrAttachmentNotFound, 404},
		{"TimeEntryNotFound", ErrTimeEntryNotFound, 404},
		{"RecurrenceNotFound", ErrRecurrenceNotFound, 404},
		{"NotificationNotFound", ErrNotificationNotFound, 404},
		{"AgentNotFound", ErrAgentNotFound, 404},

		// Validation errors
		{"RequiredField", ErrRequiredField, 422},
		{"InvalidFormat", ErrInvalidFormat, 422},
		{"InvalidValue", ErrInvalidValue, 422},
		{"MaxDepthExceeded", ErrMaxDepthExceeded, 422},
		{"SelfReference", ErrSelfReference, 422},

		// Business errors
		{"AlreadyAssigned", ErrAlreadyAssigned, 409},
		{"AlreadyLabelled", ErrAlreadyLabelled, 409},
		{"StateTransition", ErrStateTransition, 422},
		{"WorkflowViolation", ErrWorkflowViolation, 422},
		{"DuplicateEntry", ErrDuplicateEntry, 409},
		{"TimerRunning", ErrTimerRunning, 500}, // No explicit mapping, defaults to 500

		// AI errors
		{"AIConfigMissing", ErrAIConfigMissing, 500},
		{"AIAPIFailure", ErrAIAPIFailure, 502},
		{"AITimeout", ErrAITimeout, 502},
		{"AIQuotaExceeded", ErrAIQuotaExceeded, 429},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.code.HTTPStatus()
			if got != tt.expected {
				t.Errorf("ErrorCode(%q).HTTPStatus() = %d, want %d", tt.code, got, tt.expected)
			}
		})
	}
}

func TestErrorCode_NOT_FOUND_Suffix(t *testing.T) {
	// All error codes ending with _NOT_FOUND should return 404.
	notFoundCodes := []ErrorCode{
		ErrProjectNotFound, ErrIssueNotFound, ErrUserNotFound,
		ErrWorkspaceNotFound, ErrPageNotFound, ErrCycleNotFound,
		ErrModuleNotFound, ErrStateNotFound, ErrLabelNotFound,
		ErrReleaseNotFound, ErrTemplateNotFound, ErrViewNotFound,
		ErrCommentNotFound, ErrAttachmentNotFound, ErrTimeEntryNotFound,
		ErrRecurrenceNotFound, ErrNotificationNotFound, ErrAgentNotFound,
	}

	for _, code := range notFoundCodes {
		if !stringsHasSuffix(string(code), "_NOT_FOUND") {
			t.Errorf("%q does not end with _NOT_FOUND", code)
		}
		if code.HTTPStatus() != 404 {
			t.Errorf("%q should return 404, got %d", code, code.HTTPStatus())
		}
	}
}

func TestStringsHasSuffix(t *testing.T) {
	tests := []struct {
		s, suffix string
		expected  bool
	}{
		{"PROJECT_NOT_FOUND", "_NOT_FOUND", true},
		{"INTERNAL_ERROR", "_NOT_FOUND", false},
		{"", "_NOT_FOUND", false},
		{"NOT_FOUND", "_NOT_FOUND", false}, // "NOT_FOUND" is shorter than "_NOT_FOUND"
		{"_NOT_FOUND_EXTRA", "_NOT_FOUND", false},
		{"long-string-NOT_FOUND", "_NOT_FOUND", false},
		{"ALREADY_ASSIGNED", "_ASSIGNED", true},
	}

	for _, tt := range tests {
		t.Run(tt.s+"/"+tt.suffix, func(t *testing.T) {
			got := stringsHasSuffix(tt.s, tt.suffix)
			if got != tt.expected {
				t.Errorf("stringsHasSuffix(%q, %q) = %v, want %v", tt.s, tt.suffix, got, tt.expected)
			}
		})
	}
}

func TestErrorCode_Uniqueness(t *testing.T) {
	// Verify no duplicate error code values.
	allCodes := []ErrorCode{
		ErrInternal, ErrBadRequest, ErrValidation, ErrUnauthorized, ErrForbidden,
		ErrNotFound, ErrAlreadyExists, ErrConflict,
		ErrProjectNotFound, ErrIssueNotFound, ErrUserNotFound, ErrWorkspaceNotFound,
		ErrPageNotFound, ErrCycleNotFound, ErrModuleNotFound, ErrStateNotFound,
		ErrLabelNotFound, ErrReleaseNotFound, ErrTemplateNotFound, ErrViewNotFound,
		ErrCommentNotFound, ErrAttachmentNotFound, ErrTimeEntryNotFound,
		ErrRecurrenceNotFound, ErrNotificationNotFound, ErrAgentNotFound,
		ErrRequiredField, ErrInvalidFormat, ErrInvalidValue, ErrMaxDepthExceeded, ErrSelfReference,
		ErrAlreadyAssigned, ErrAlreadyLabelled, ErrStateTransition, ErrWorkflowViolation,
		ErrDuplicateEntry, ErrTimerRunning,
		ErrAIConfigMissing, ErrAIAPIFailure, ErrAITimeout, ErrAIQuotaExceeded,
	}

	seen := make(map[string]bool)
	for _, code := range allCodes {
		if seen[string(code)] {
			t.Errorf("Duplicate error code: %s", code)
		}
		seen[string(code)] = true
	}
}
