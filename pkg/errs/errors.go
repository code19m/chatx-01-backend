package errs

import "errors"

// Generic repository errors.
var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

// NotFoundError represents a resource not found error.
type NotFoundError struct {
	Field   string
	Message string
}

func NewNotFoundError(field, message string) error {
	return NotFoundError{
		Field:   field,
		Message: message,
	}
}

func (e NotFoundError) Error() string {
	return e.Message
}

// ConflictError represents a resource conflict error.
type ConflictError struct {
	Field   string
	Message string
}

func NewConflictError(field, message string) error {
	return ConflictError{
		Field:   field,
		Message: message,
	}
}

func (e ConflictError) Error() string {
	return e.Message
}

// ReplaceOn replaces target error with replacement if err matches target.
// This should only be used for user input errors.
func ReplaceOn(err error, target error, replacement error) error {
	if errors.Is(err, target) {
		return replacement
	}
	return err
}
