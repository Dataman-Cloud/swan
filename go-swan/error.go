package swan

import (
	"fmt"
)

// APIError represents a generic API error.
type APIError struct {
	// ErrCode specifies the nature of the error.
	ErrCode int
	message string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("Swan API error: %s", e.message)
}

// NewAPIError creates a new APIError instance from the given response code and content.
func NewAPIError(code int, content []byte) error {
	return &APIError{message: string(content), ErrCode: code}
}
