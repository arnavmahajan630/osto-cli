package totp

import (
	"context"
	"errors"

	"osto-auth-cli/internal/repository"
	"osto-auth-cli/internal/secure"
	"osto-auth-cli/internal/session"
)

// EnrollmentService handles the orchestration of TOTP enrollment securely.
type EnrollmentService struct {
	userRepo       repository.UserRepository
	sessionService session.SessionService
	totpService    TOTPService
	appKey         []byte
}

// NewEnrollmentService creates a new EnrollmentService.
func NewEnrollmentService(
	userRepo repository.UserRepository,
	sessionService session.SessionService,
	totpService TOTPService,
	appKey []byte,
) *EnrollmentService {
	return &EnrollmentService{
		userRepo:       userRepo,
		sessionService: sessionService,
		totpService:    totpService,
		appKey:         appKey,
	}
}

// ConfirmEnrollment validates the user's code, encrypts the secret, and finalizes the DB setup.
func (s *EnrollmentService) ConfirmEnrollment(
	ctx context.Context,
	userID int64,
	sessionToken string,
	rawSecret string,
	code string,
) error {
	if !s.totpService.Validate(rawSecret, code) {
		return ErrInvalidTOTP
	}

	encSecret, err := secure.Encrypt([]byte(rawSecret), s.appKey)
	if err != nil {
		return err
	}

	if err := s.userRepo.SetMFA(ctx, userID, true, &encSecret); err != nil {
		return err
	}

	if err := s.sessionService.Revoke(ctx, sessionToken); err != nil {
		return err
	}

	return nil
}

// Disable validates the user's code, removes the MFA secret, and revokes the session.
func (s *EnrollmentService) Disable(
	ctx context.Context,
	userID int64,
	sessionToken string,
	encSecret *string,
	code string,
) error {
	if encSecret == nil {
		return errors.New("MFA is not enabled")
	}

	plaintext, err := secure.Decrypt(*encSecret, s.appKey)
	if err != nil {
		return err
	}

	if !s.totpService.Validate(string(plaintext), code) {
		return ErrInvalidTOTP
	}

	if err := s.userRepo.SetMFA(ctx, userID, false, nil); err != nil {
		return err
	}

	if err := s.sessionService.Revoke(ctx, sessionToken); err != nil {
		return err
	}

	return nil
}
