package routers

import (
	"errors"
	"fmt"
)

var ErrValidation = errors.New("validation error")

func NewValidationError(detailMsg string) error {
	return fmt.Errorf("%w: %s", ErrValidation, detailMsg)
}
