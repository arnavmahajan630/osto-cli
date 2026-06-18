package auth

import (
	"context"
	"errors"

	"time"

	"osto-auth-cli/internal/models"
	"osto-auth-cli/internal/repository"
	"osto-auth-cli/internal/secure"
	"osto-auth-cli/internal/session"
	"osto-auth-cli/internal/validation"
)

// RegisterInput encapsulates the data required to register a new user.
type RegisterInput struct {
	Username  string
	Password  string
	Name      *string
	BirthDate *string
}

// LoginResult contains the session token and user details upon successful login.
type LoginResult struct {
	User         *models.User
	RequiresTOTP bool
	SessionToken string
}

// AuthService defines the contract for authentication operations.
type AuthService interface {
	Register(ctx context.Context, input RegisterInput) error
	Login(ctx context.Context, username, password string) (*LoginResult, error)
	Logout(ctx context.Context, token string) error
	VerifySession(ctx context.Context, token string) (*models.User, error)
	EnableMFA(ctx context.Context, token string) (secret string, err error)
	VerifyMFA(ctx context.Context, token, code string) error
}

// DefaultAuthService provides the concrete implementation of AuthService.
type DefaultAuthService struct {
	repo           repository.UserRepository
	sessionService session.SessionService
}

// NewAuthService creates a new DefaultAuthService.
func NewAuthService(repo repository.UserRepository, sessionService session.SessionService) *DefaultAuthService {
	return &DefaultAuthService{
		repo:           repo,
		sessionService: sessionService,
	}
}

// Register validates inputs, hashes the password, and creates a new user.
func (s *DefaultAuthService) Register(ctx context.Context, input RegisterInput) error {
	// 1. Validate inputs
	if err := validation.Username(input.Username); err != nil {
		return err
	}
	if err := validation.Password(input.Password); err != nil {
		return err
	}
	if input.BirthDate != nil {
		if err := validation.Date(*input.BirthDate); err != nil {
			return err
		}
	}

	// 2. Convenience check for duplicates (avoids hashing if already exists)
	exists, err := s.repo.ExistsByUsername(ctx, input.Username)
	if err != nil {
		return err
	}
	if exists {
		return ErrUserExists
	}

	// 3. Hash password
	hash, err := secure.HashPassword(input.Password)
	if err != nil {
		return err
	}

	// 4. Create user
	user := &models.User{
		Username:     input.Username,
		PasswordHash: hash,
		Name:         input.Name,
		BirthDate:    input.BirthDate,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		// Catch unique constraint violations in case of race conditions
		if errors.Is(err, repository.ErrDuplicateUsername) {
			return ErrUserExists
		}
		return err
	}

	return nil
}

// Login verifies credentials and creates a session.
func (s *DefaultAuthService) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		// Do not distinguish between non-existent user and bad password
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !secure.VerifyPassword(user.PasswordHash, password) {
		return nil, ErrInvalidCredentials
	}

	if err := s.repo.RecordLoginSuccess(ctx, user.ID, time.Now()); err != nil {
		return nil, err
	}

	rawToken, err := s.sessionService.Create(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		User:         user,
		RequiresTOTP: false,
		SessionToken: rawToken,
	}, nil
}

func (s *DefaultAuthService) Logout(ctx context.Context, token string) error {
	return errors.New("not implemented")
}

func (s *DefaultAuthService) VerifySession(ctx context.Context, token string) (*models.User, error) {
	return nil, errors.New("not implemented")
}

func (s *DefaultAuthService) EnableMFA(ctx context.Context, token string) (secret string, err error) {
	return "", errors.New("not implemented")
}

func (s *DefaultAuthService) VerifyMFA(ctx context.Context, token, code string) error {
	return errors.New("not implemented")
}
