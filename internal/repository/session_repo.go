package repository

import (
	"context"
	"time"

	"osto-auth-cli/internal/models"
)

// SessionRepository defines the persistence contract for sessions.
type SessionRepository interface {
	Create(ctx context.Context, s *models.Session) (int64, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (*models.Session, error)
	Touch(ctx context.Context, tokenHash string, at time.Time) error
	Revoke(ctx context.Context, tokenHash string) error
	DeleteExpired(ctx context.Context, now time.Time) (int64, error)
}
