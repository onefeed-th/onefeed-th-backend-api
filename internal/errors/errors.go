package errors

import (
	"fmt"
	"runtime"
)

// ErrorType represents different types of errors in the application
type ErrorType string

const (
	ValidationError ErrorType = "VALIDATION_ERROR"
	DatabaseError   ErrorType = "DATABASE_ERROR"
	RedisError      ErrorType = "REDIS_ERROR"
	NetworkError    ErrorType = "NETWORK_ERROR"
	ParseError      ErrorType = "PARSE_ERROR"
	InternalError   ErrorType = "INTERNAL_ERROR"
)

// AppError represents a structured application error
type AppError struct {
	Type    ErrorType `json:"type"`
	Message string    `json:"message"`
	Code    string    `json:"code,omitempty"`
	Details string    `json:"details,omitempty"`
	Cause   error     `json:"-"`
	File    string    `json:"file,omitempty"`
	Line    int       `json:"line,omitempty"`
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

// WithCaller adds file and line information to the error
func (e *AppError) WithCaller() *AppError {
	_, file, line, ok := runtime.Caller(1)
	if ok {
		e.File = file
		e.Line = line
	}
	return e
}

// New creates a new application error
func New(errType ErrorType, message string) *AppError {
	return &AppError{
		Type:    errType,
		Message: message,
	}
}

// Newf creates a new application error with formatted message
func Newf(errType ErrorType, format string, args ...interface{}) *AppError {
	return &AppError{
		Type:    errType,
		Message: fmt.Sprintf(format, args...),
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, errType ErrorType, message string) *AppError {
	return &AppError{
		Type:    errType,
		Message: message,
		Cause:   err,
	}
}

// Wrapf wraps an existing error with formatted message
func Wrapf(err error, errType ErrorType, format string, args ...interface{}) *AppError {
	return &AppError{
		Type:    errType,
		Message: fmt.Sprintf(format, args...),
		Cause:   err,
	}
}

// WithCode adds an error code
func (e *AppError) WithCode(code string) *AppError {
	e.Code = code
	return e
}

// WithDetails adds additional details
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// IsType checks if error is of specific type
func IsType(err error, errType ErrorType) bool {
	var appErr *AppError
	if As(err, &appErr) {
		return appErr.Type == errType
	}
	return false
}

// As is a convenience function for errors.As
func As(err error, target interface{}) bool {
	// This would normally use errors.As from Go standard library
	// For simplicity, we'll implement basic type assertion
	if appErr, ok := err.(*AppError); ok {
		if targetPtr, ok := target.(**AppError); ok {
			*targetPtr = appErr
			return true
		}
	}
	return false
}