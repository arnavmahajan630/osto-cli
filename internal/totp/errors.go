package totp

import "errors"

var (
	ErrInvalidTOTP = errors.New("invalid TOTP code")
)
