package serviceLogs

import (
	"fmt"
	"github.com/pkg/errors"
)

// ErrInvalidRequest represents errors related to invalid API requests
type ErrInvalidRequest struct {
	Op      string
	Message string
	Err     error
}

func (e *ErrInvalidRequest) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Op, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Op, e.Message)
}

func (e *ErrInvalidRequest) Unwrap() error {
	return e.Err
}

// ErrLogResponse represents errors related to log response parsing or validation
type ErrLogResponse struct {
	StatusCode int
	Message    string
	Err        error
}

func (e *ErrLogResponse) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("log response error (status %d): %s: %v", e.StatusCode, e.Message, e.Err)
	}
	return fmt.Sprintf("log response error (status %d): %s", e.StatusCode, e.Message)
}

func (e *ErrLogResponse) Unwrap() error {
	return e.Err
}

// NewInvalidRequestError creates a new ErrInvalidRequest
func NewInvalidRequestError(op string, message string, err error) error {
	return &ErrInvalidRequest{
		Op:      op,
		Message: message,
		Err:     err,
	}
}

// NewLogResponseError creates a new ErrLogResponse
func NewLogResponseError(statusCode int, message string, err error) error {
	return &ErrLogResponse{
		StatusCode: statusCode,
		Message:    message,
		Err:        err,
	}
}

// IsInvalidRequestError checks if the error is an ErrInvalidRequest
func IsInvalidRequestError(err error) bool {
	var target *ErrInvalidRequest
	return errors.As(err, &target)
}

// IsLogResponseError checks if the error is an ErrLogResponse
func IsLogResponseError(err error) bool {
	var target *ErrLogResponse
	return errors.As(err, &target)
}
