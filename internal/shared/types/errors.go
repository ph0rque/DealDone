package types

import (
	"fmt"
)

// ErrorType represents the type of error
type ErrorType string

const (
	// ErrorTypeValidation indicates a validation error
	ErrorTypeValidation ErrorType = "validation"
	// ErrorTypeNotFound indicates a resource was not found
	ErrorTypeNotFound ErrorType = "not_found"
	// ErrorTypePermission indicates a permission error
	ErrorTypePermission ErrorType = "permission"
	// ErrorTypeInternal indicates an internal error
	ErrorTypeInternal ErrorType = "internal"
	// ErrorTypeExternal indicates an external service error
	ErrorTypeExternal ErrorType = "external"
	// ErrorTypeTimeout indicates a timeout error
	ErrorTypeTimeout ErrorType = "timeout"
	// ErrorTypeConflict indicates a conflict error
	ErrorTypeConflict ErrorType = "conflict"
)

// AppError represents a structured application error
type AppError struct {
	Type    ErrorType              `json:"type"`
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
	Cause   error                  `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Cause
}

// NewAppError creates a new application error
func NewAppError(errType ErrorType, code string, message string) *AppError {
	return &AppError{
		Type:    errType,
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// WithCause adds a cause to the error
func (e *AppError) WithCause(cause error) *AppError {
	e.Cause = cause
	return e
}

// WithDetail adds a detail to the error
func (e *AppError) WithDetail(key string, value interface{}) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// Common error constructors

// NewValidationError creates a validation error
func NewValidationError(code string, message string) *AppError {
	return NewAppError(ErrorTypeValidation, code, message)
}

// NewNotFoundError creates a not found error
func NewNotFoundError(code string, message string) *AppError {
	return NewAppError(ErrorTypeNotFound, code, message)
}

// NewPermissionError creates a permission error
func NewPermissionError(code string, message string) *AppError {
	return NewAppError(ErrorTypePermission, code, message)
}

// NewInternalError creates an internal error
func NewInternalError(code string, message string) *AppError {
	return NewAppError(ErrorTypeInternal, code, message)
}

// NewExternalError creates an external service error
func NewExternalError(code string, message string) *AppError {
	return NewAppError(ErrorTypeExternal, code, message)
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(code string, message string) *AppError {
	return NewAppError(ErrorTypeTimeout, code, message)
}

// NewConflictError creates a conflict error
func NewConflictError(code string, message string) *AppError {
	return NewAppError(ErrorTypeConflict, code, message)
}
