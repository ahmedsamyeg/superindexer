package errors

import (
	"fmt"
)

// AppError represents an application-level error.
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Error implements the error interface.
func (e *AppError) Error() string {
	return e.Message
}

// NewAppError creates a new application error.
func NewAppError(code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// NewAppErrorWithDetails creates a new application error with details.
func NewAppErrorWithDetails(code, message, details string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// NotFound returns a not found error.
func NotFound(resource string, id interface{}) *AppError {
	return NewAppError("not_found", fmt.Sprintf("%s not found: %v", resource, id))
}

// InvalidInput returns an invalid input error.
func InvalidInput(field, message string) *AppError {
	return NewAppErrorWithDetails(
		"invalid_input",
		fmt.Sprintf("invalid %s: %s", field, message),
		fmt.Sprintf("field=%s,message=%s", field, message),
	)
}

// Conflict returns a conflict error.
func Conflict(resource string, message string) *AppError {
	return NewAppError("conflict", fmt.Sprintf("%s: %s", resource, message))
}

// Internal returns an internal server error.
func Internal(message string) *AppError {
	return NewAppError("internal_error", message)
}

// Unauthorized returns an unauthorized error.
func Unauthorized(message string) *AppError {
	return NewAppError("unauthorized", message)
}

// Forbidden returns a forbidden error.
func Forbidden(message string) *AppError {
	return NewAppError("forbidden", message)
}

// IsAppError checks if an error is an AppError.
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// UnwrapAppError extracts the AppError from a wrapped error.
func UnwrapAppError(err error) *AppError {
	var appErr *AppError
	if err == nil {
		return nil
	}
	for err != nil {
		if e, ok := err.(*AppError); ok {
			return e
		}
		err = fmt.Errorf("%w", err)
	}
	return nil
}
