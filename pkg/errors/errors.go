package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorType represents the category of error
type ErrorType string

const (
	ErrorTypeNotFound      ErrorType = "NOT_FOUND"
	ErrorTypeAlreadyExists ErrorType = "ALREADY_EXISTS"
	ErrorTypeValidation    ErrorType = "VALIDATION"
	ErrorTypeUnauthorized  ErrorType = "UNAUTHORIZED"
	ErrorTypeForbidden     ErrorType = "FORBIDDEN"
	ErrorTypeInternal      ErrorType = "INTERNAL"
	ErrorTypeBadRequest    ErrorType = "BAD_REQUEST"
)

// AppError represents an application error with additional context
type AppError struct {
	Type    ErrorType
	Message string
	Err     error
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap implements the errors.Unwrap interface
func (e *AppError) Unwrap() error {
	return e.Err
}

// HTTPStatus returns the HTTP status code for the error type
func (e *AppError) HTTPStatus() int {
	switch e.Type {
	case ErrorTypeNotFound:
		return http.StatusNotFound
	case ErrorTypeAlreadyExists:
		return http.StatusConflict
	case ErrorTypeValidation:
		return http.StatusBadRequest
	case ErrorTypeUnauthorized:
		return http.StatusUnauthorized
	case ErrorTypeForbidden:
		return http.StatusForbidden
	case ErrorTypeBadRequest:
		return http.StatusBadRequest
	case ErrorTypeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// New creates a new AppError
func New(errType ErrorType, message string) *AppError {
	return &AppError{
		Type:    errType,
		Message: message,
	}
}

// Wrap wraps an existing error with additional context
func Wrap(errType ErrorType, message string, err error) *AppError {
	return &AppError{
		Type:    errType,
		Message: message,
		Err:     err,
	}
}

// NotFound creates a not found error
func NotFound(message string) *AppError {
	return New(ErrorTypeNotFound, message)
}

// AlreadyExists creates an already exists error
func AlreadyExists(message string) *AppError {
	return New(ErrorTypeAlreadyExists, message)
}

// Validation creates a validation error
func Validation(message string) *AppError {
	return New(ErrorTypeValidation, message)
}

// Unauthorized creates an unauthorized error
func Unauthorized(message string) *AppError {
	return New(ErrorTypeUnauthorized, message)
}

// Forbidden creates a forbidden error
func Forbidden(message string) *AppError {
	return New(ErrorTypeForbidden, message)
}

// Internal creates an internal error
func Internal(message string, err error) *AppError {
	return Wrap(ErrorTypeInternal, message, err)
}

// BadRequest creates a bad request error
func BadRequest(message string) *AppError {
	return New(ErrorTypeBadRequest, message)
}

// Is checks if the error matches the given type
func Is(err error, errType ErrorType) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == errType
	}
	return false
}

// GetType returns the error type if it's an AppError
func GetType(err error) ErrorType {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type
	}
	return ErrorTypeInternal
}

// GetHTTPStatus returns the HTTP status code for the error
func GetHTTPStatus(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.HTTPStatus()
	}
	return http.StatusInternalServerError
}