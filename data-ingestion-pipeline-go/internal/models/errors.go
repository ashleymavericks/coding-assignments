package models

import (
	"errors"
	"fmt"
)

// Go Concept: Custom error definitions
// In Go, errors are just values that implement the error interface
// We define common errors as package-level variables

// Validation errors
var (
	ErrInvalidPostID = errors.New("post ID must be greater than 0")
	ErrInvalidUserID = errors.New("user ID must be greater than 0")
	ErrEmptyTitle    = errors.New("post title cannot be empty")
	ErrEmptySource   = errors.New("post source cannot be empty")
)

// Repository errors
var (
	ErrPostNotFound       = errors.New("post not found")
	ErrPostAlreadyExists  = errors.New("post already exists")
	ErrDatabaseConnection = errors.New("database connection failed")
)

// API errors
var (
	ErrAPIUnavailable     = errors.New("external API is unavailable")
	ErrInvalidAPIResponse = errors.New("invalid API response format")
	ErrAPIRateLimit       = errors.New("API rate limit exceeded")
)

// Custom error types for more detailed error information
// Go Concept: Custom error types with additional context

// ValidationError wraps validation errors with field information
type ValidationError struct {
	Field   string      // Which field failed validation
	Value   interface{} // The invalid value
	Message string      // Human-readable message
}

// Error implements the error interface
// Go Concept: Any type with an Error() string method implements error
func (ve ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s (value: %v)",
		ve.Field, ve.Message, ve.Value)
}

// APIError represents errors from external API calls
type APIError struct {
	StatusCode int    // HTTP status code
	Message    string // Error message
	URL        string // The URL that failed
}

// Error implements the error interface
func (ae APIError) Error() string {
	return fmt.Sprintf("API error %d for %s: %s", ae.StatusCode, ae.URL, ae.Message)
}

// IsTemporary indicates if this error might be retryable
// Go Concept: Methods can provide additional behavior
func (ae APIError) IsTemporary() bool {
	// 5xx errors and some 4xx errors might be temporary
	return ae.StatusCode >= 500 || ae.StatusCode == 429 || ae.StatusCode == 408
}

// DatabaseError represents database-related errors
type DatabaseError struct {
	Operation string // What operation failed (SELECT, INSERT, etc.)
	Table     string // Which table was involved
	Err       error  // The underlying error
}

// Error implements the error interface
func (de DatabaseError) Error() string {
	return fmt.Sprintf("database error during %s on %s: %v",
		de.Operation, de.Table, de.Err)
}

// Unwrap allows error unwrapping for Go 1.13+ error handling
// Go Concept: Error unwrapping for better error handling
func (de DatabaseError) Unwrap() error {
	return de.Err
}

// Helper functions for creating specific errors

// NewValidationError creates a new validation error
func NewValidationError(field string, value interface{}, message string) ValidationError {
	return ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}

// NewAPIError creates a new API error
func NewAPIError(statusCode int, message, url string) APIError {
	return APIError{
		StatusCode: statusCode,
		Message:    message,
		URL:        url,
	}
}

// NewDatabaseError creates a new database error
func NewDatabaseError(operation, table string, err error) DatabaseError {
	return DatabaseError{
		Operation: operation,
		Table:     table,
		Err:       err,
	}
}
