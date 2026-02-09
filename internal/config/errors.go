package config

import "fmt"

// ValidationError indicates one option value is invalid.
type ValidationError struct {
	Field   string
	Message string
}

// Error returns a user-facing validation error message.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("invalid %s: %s", e.Field, e.Message)
}

// ConflictError indicates two options cannot be used together.
type ConflictError struct {
	Left  string
	Right string
}

// Error returns a user-facing conflict error message.
func (e *ConflictError) Error() string {
	return fmt.Sprintf("options %s and %s cannot be used together", e.Left, e.Right)
}

// NewValidationError constructs a validation error.
func NewValidationError(field, message string) error {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// NewConflictError constructs an option conflict error.
func NewConflictError(left, right string) error {
	return &ConflictError{
		Left:  left,
		Right: right,
	}
}

// WrapError adds config operation context while preserving the original error.
func WrapError(op string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("config %s: %w", op, err)
}
