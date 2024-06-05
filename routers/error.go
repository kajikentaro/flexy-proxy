package routers

import "fmt"

type ValidationError struct {
	message string
	cause   error
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s", e.message)
}

func (e *ValidationError) Unwrap() error {
	return e.cause
}

func NewValidationError(msg string, cause error) *ValidationError {
	return &ValidationError{
		message: msg,
		cause:   cause,
	}
}
