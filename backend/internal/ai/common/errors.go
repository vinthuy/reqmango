package common

import "fmt"

type AppError struct {
	Code    int    `json:"-"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

func (e *AppError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Detail)
	}
	return e.Message
}

func NotFound(msg string) *AppError {
	return &AppError{Code: 404, Message: msg}
}

func BadRequest(msg string) *AppError {
	return &AppError{Code: 400, Message: msg}
}

func Internal(msg string) *AppError {
	return &AppError{Code: 500, Message: msg}
}

func Validation(msg string) *AppError {
	return &AppError{Code: 422, Message: msg}
}

func Unauthorized(msg string) *AppError {
	return &AppError{Code: 401, Message: msg}
}

func Permission(msg string) *AppError {
	return &AppError{Code: 403, Message: msg}
}

// AgentNotFound returns a 404 for missing agents.
func AgentNotFound() *AppError {
	return &AppError{Code: 404, Message: "Agent not found"}
}
