package routers

import (
	"errors"
	"fmt"
)

var ErrValidation = errors.New("validation error")

func NewValidationError(position string, message string, actual string) error {
	return fmt.Errorf(
		"%w: %s",
		ErrValidation,
		fmt.Sprintf("%s Found: \"%s\" on %s", message, actual, position),
	)
}
