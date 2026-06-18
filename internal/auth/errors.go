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
// Retained for backward compatibility.
type ErrorAccountLocked struct {
	Until time.Time
}

func (e *ErrorAccountLocked) Error() string {
	return ErrAccountLocked.Error()
}

func (e *ErrorAccountLocked) Is(target error) bool {
	return target == ErrAccountLocked
}

// AuthFailure represents an authentication error containing lockout metadata.
type AuthFailure struct {
	Err               error
	AttemptsRemaining int
	LockedUntil       *time.Time
}

func (e *AuthFailure) Error() string {
	return e.Err.Error()
}

func (e *AuthFailure) Unwrap() error {
	return e.Err
}
