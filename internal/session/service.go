package session

import (
	"context"
	"time"

	"osto-auth-cli/internal/config"
	"osto-auth-cli/internal/models"
	"osto-auth-cli/internal/repository"
	"osto-auth-cli/internal/secure"
)

// SessionService defines the contract for session management.
type SessionService interface {
	Create(ctx context.Context, userID int64) (rawToken string, err error)
	Revoke(ctx context.Context, rawToken string) error
}

// DefaultSessionService implements SessionService.
type DefaultSessionService struct {
	repo repository.SessionRepository
	cfg  *config.Config
}

// NewSessionService creates a new DefaultSessionService.
func NewSessionService(repo repository.SessionRepository, cfg *config.Config) *DefaultSessionService {
	return &DefaultSessionService{
		repo: repo,
		cfg:  cfg,
	}
}

// Create generates a cryptographically secure token, hashes it,
// and persists the session with the configured expiration time.
func (s *DefaultSessionService) Create(ctx context.Context, userID int64) (string, error) {
	rawToken, err := secure.NewSessionToken()
	if err != nil {
		return "", err
	}

	tokenHash := secure.HashToken(rawToken)

	session := &models.Session{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(s.cfg.SessionTime),
	}

	if _, err := s.repo.Create(ctx, session); err != nil {
		return "", err
	}

	return rawToken, nil
}

// Revoke revokes the session associated with the provided raw token.
func (s *DefaultSessionService) Revoke(ctx context.Context, rawToken string) error {
	tokenHash := secure.HashToken(rawToken)
	return s.repo.Revoke(ctx, tokenHash)
}
