package common

import (
	"testing"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name    string
		err     *AppError
		wantMsg string
	}{
		{
			name:    "with detail",
			err:     &AppError{Code: 404, Message: "not found", Detail: "user 42"},
			wantMsg: "not found: user 42",
		},
		{
			name:    "without detail",
			err:     &AppError{Code: 500, Message: "internal error"},
			wantMsg: "internal error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.wantMsg {
				t.Errorf("Error() = %q, want %q", got, tt.wantMsg)
			}
		})
	}
}

func TestNotFound(t *testing.T) {
	err := NotFound("resource missing")
	if err.Code != 404 {
		t.Errorf("Code = %d, want 404", err.Code)
	}
	if err.Message != "resource missing" {
		t.Errorf("Message = %q, want 'resource missing'", err.Message)
	}
}

func TestBadRequest(t *testing.T) {
	err := BadRequest("invalid input")
	if err.Code != 400 {
		t.Errorf("Code = %d, want 400", err.Code)
	}
}

func TestInternal(t *testing.T) {
	err := Internal("something went wrong")
	if err.Code != 500 {
		t.Errorf("Code = %d, want 500", err.Code)
	}
}

func TestValidation(t *testing.T) {
	err := Validation("field required")
	if err.Code != 422 {
		t.Errorf("Code = %d, want 422", err.Code)
	}
}

func TestUnauthorized(t *testing.T) {
	err := Unauthorized("invalid token")
	if err.Code != 401 {
		t.Errorf("Code = %d, want 401", err.Code)
	}
}

func TestAgentNotFound(t *testing.T) {
	err := AgentNotFound()
	if err.Code != 404 {
		t.Errorf("Code = %d, want 404", err.Code)
	}
	if err.Message != "Agent not found" {
		t.Errorf("Message = %q, want 'Agent not found'", err.Message)
	}
}
