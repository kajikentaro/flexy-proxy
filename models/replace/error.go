package replace

import "fmt"

type UrlReplaceError struct {
	message string
	cause   error
}

func (e *UrlReplaceError) Error() string {
	return fmt.Sprintf("URL replacement error: %s", e.message)
}

func (e *UrlReplaceError) Unwrap() error {
	return e.cause
}

func NewUrlReplaceError(msg string, cause error) *UrlReplaceError {
	return &UrlReplaceError{
		message: msg,
		cause:   cause,
	}
}
