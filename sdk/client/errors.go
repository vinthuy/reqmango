package client

import (
	"errors"
	"fmt"
)

// APIError is a typed error carrying the backend's {"message": ...} body.
type APIError struct {
	StatusCode int
	Message    string
	Body       map[string]any // full parsed error body (e.g. 409 approval payload)
}

func (e *APIError) Error() string {
	return fmt.Sprintf("api error %d: %s", e.StatusCode, e.Message)
}

// AsAPIError unwraps err to *APIError, or nil.
func AsAPIError(err error) *APIError {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}
	return nil
}
