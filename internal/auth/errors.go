package auth

import (
	"errors"
	"time"
)

var (
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountLocked      = errors.New("account locked")
)

// ErrorAccountLocked encapsulates the time until which the account is locked.
type ErrorAccountLocked struct {
	Until time.Time
}

func (e *ErrorAccountLocked) Error() string {
	return ErrAccountLocked.Error()
}

func (e *ErrorAccountLocked) Is(target error) bool {
	return target == ErrAccountLocked
}
