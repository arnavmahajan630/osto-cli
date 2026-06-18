package auth

import "errors"

var (
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidTOTP        = errors.New("invalid TOTP code")
	ErrAccountLocked      = errors.New("account locked")
)
