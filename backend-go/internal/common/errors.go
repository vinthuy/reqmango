package common

import "net/http"

// AppError represents an application-level error with an HTTP status code.
type AppError struct {
	Code    int    `json:"-"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}

// NotFound returns a 404 error.
func NotFound(msg string) *AppError {
	return &AppError{Code: http.StatusNotFound, Message: msg}
}

// Conflict returns a 409 error (duplicate resource).
func Conflict(msg string) *AppError {
	return &AppError{Code: http.StatusConflict, Message: msg}
}

// Unauthorized returns a 401 error.
func Unauthorized(msg string) *AppError {
	return &AppError{Code: http.StatusUnauthorized, Message: msg}
}

// Forbidden returns a 403 error.
func Forbidden(msg string) *AppError {
	return &AppError{Code: http.StatusForbidden, Message: msg}
}

// Validation returns a 422 error.
func Validation(msg string) *AppError {
	return &AppError{Code: http.StatusUnprocessableEntity, Message: msg}
}

// Internal returns a 500 error.
func Internal(msg string) *AppError {
	return &AppError{Code: http.StatusInternalServerError, Message: msg}
}

// BadRequest returns a 400 error.
func BadRequest(msg string) *AppError {
	return &AppError{Code: http.StatusBadRequest, Message: msg}
}
