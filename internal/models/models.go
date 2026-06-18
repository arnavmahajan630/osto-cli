package models

import "time"

type User struct {
	ID             int64
	Username       string
	PasswordHash   string
	Name           *string
	BirthDate      *string
	CreatedAt      time.Time
	LastLoginAt    *time.Time
	MFAEnabled     bool
	MFASecretEnc   *string
	FailedAttempts int
	LockedUntil    *time.Time
}

type Session struct {
	ID           int64
	UserID       int64
	TokenHash    string
	CreatedAt    time.Time
	ExpiresAt    time.Time
	LastActiveAt time.Time
	RevokedAt    *time.Time
}
