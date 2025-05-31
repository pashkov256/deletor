package errors

import (
	"time"
)

// ErrorType represents different types of errors that can occur
type ErrorType int

const (
	ErrorTypeValidation ErrorType = iota
	ErrorTypeFileSystem
	ErrorTypePermission
	ErrorTypeConfiguration
)

// Error represents an application error with additional context
type Error struct {
	Type      ErrorType
	Message   string
	visible   bool
	Timestamp time.Time
}

// New creates a new error with the given type and message
func New(errType ErrorType, message string) *Error {
	return &Error{
		Type:      errType,
		Message:   message,
		visible:   true,
		Timestamp: time.Now(),
	}
}

// Hide makes the error invisible
func (e *Error) Hide() {
	e.visible = false
}

// Show makes the error visible
func (e *Error) Show() {
	e.visible = true
}

// IsVisible returns whether the error is currently visible
func (e *Error) IsVisible() bool {
	return e.visible
}

// GetMessage returns the error message
func (e *Error) GetMessage() string {
	return e.Message
}

// GetType returns the error type
func (e *Error) GetType() ErrorType {
	return e.Type
}
