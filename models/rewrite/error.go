package rewrite

import "fmt"

type RewriteError struct {
	message string
	cause   error
}

func (e *RewriteError) Error() string {
	return fmt.Sprintf("URL rewrite error: %s", e.message)
}

func (e *RewriteError) Unwrap() error {
	return e.cause
}

func newUrlRewriteError(msg string, cause error) *RewriteError {
	return &RewriteError{
		message: msg,
		cause:   cause,
	}
}
