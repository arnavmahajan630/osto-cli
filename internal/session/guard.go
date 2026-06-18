package session

import (
	"context"
	"errors"
	"time"

	"osto-auth-cli/internal/models"
	"osto-auth-cli/internal/repository"
	"osto-auth-cli/internal/secure"
	"osto-auth-cli/internal/state"
)

// AuthenticatedSession contains the user and session expiry time.
type AuthenticatedSession struct {
	User      *models.User
	ExpiresAt time.Time
}

// AuthGuard defines the interface for protecting commands.
type AuthGuard interface {
	Require(ctx context.Context, state *state.AppState) (*AuthenticatedSession, error)
}

// DefaultAuthGuard implements AuthGuard.
type DefaultAuthGuard struct {
	sessionRepo repository.SessionRepository
	userRepo    repository.UserRepository
}

// NewAuthGuard creates a new DefaultAuthGuard.
func NewAuthGuard(sessionRepo repository.SessionRepository, userRepo repository.UserRepository) *DefaultAuthGuard {
	return &DefaultAuthGuard{
		sessionRepo: sessionRepo,
		userRepo:    userRepo,
	}
}

// Require validates the current session and returns the AuthenticatedSession.
func (g *DefaultAuthGuard) Require(ctx context.Context, state *state.AppState) (*AuthenticatedSession, error) {
	if !state.IsAuthenticated() {
		return nil, ErrNotAuthenticated
	}

	tokenHash := secure.HashToken(state.SessionToken)
	session, err := g.sessionRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		state.Clear()
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotAuthenticated
		}
		return nil, err
	}

	if session.RevokedAt != nil {
		state.Clear()
		return nil, ErrNotAuthenticated
	}

	if session.ExpiresAt.Before(time.Now()) {
		state.Clear()
		return nil, ErrSessionExpired
	}

	if err := g.sessionRepo.Touch(ctx, tokenHash, time.Now()); err != nil {
		return nil, err
	}

	user, err := g.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		state.Clear()
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotAuthenticated
		}
		return nil, err
	}

	return &AuthenticatedSession{
		User:      user,
		ExpiresAt: session.ExpiresAt,
	}, nil
}
