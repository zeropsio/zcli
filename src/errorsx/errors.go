package errorsx

import (
	"github.com/pkg/errors"
)

type UserError struct {
	message  string
	previous error
}

func NewUserError(message string, previous error) *UserError {
	return &UserError{
		message:  message,
		previous: previous,
	}
}

func IsUserError(err error) bool {
	return AsUserError(err) != nil
}

func AsUserError(err error) *UserError {
	var userError *UserError
	if errors.As(err, &userError) {
		return userError
	}
	return nil
}

func (e *UserError) Error() string {
	return e.message
}

func (e *UserError) Unwrap() error {
	return e.previous
}

func (e *UserError) Is(target error) bool {
	return e.previous == target
}

func (e *UserError) As(target interface{}) bool {
	return errors.As(e.previous, target)
}
