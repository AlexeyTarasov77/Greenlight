package auth

import "errors"

type errInvalidData struct {
	msg string
}

func (e *errInvalidData) SetMessage(msg string) error {
	e.msg = msg
	return e
}

func (e *errInvalidData) Error() string {
	return e.msg
}

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidData          = &errInvalidData{}
	ErrUserAlreadyActivated = errors.New("user already activated")
)
