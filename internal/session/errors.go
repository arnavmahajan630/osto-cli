package session

import "errors"

var (
	ErrNotAuthenticated = errors.New("not authenticated")
	ErrSessionExpired   = errors.New("session expired")
)
