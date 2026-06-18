package integration_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"osto-auth-cli/internal/auth"
	"osto-auth-cli/internal/config"
	"osto-auth-cli/internal/db"
	"osto-auth-cli/internal/repository"
	"osto-auth-cli/internal/session"
	"osto-auth-cli/internal/state"
	"osto-auth-cli/internal/totp"
	"osto-auth-cli/migrations"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dbConn, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}

	if err := db.Migrate(dbConn, migrations.FS); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	return dbConn
}

func TestIntegration_DuplicateUsernameRejection(t *testing.T) {
	dbConn := setupTestDB(t)
	defer dbConn.Close()

	userRepo := repository.NewSQLiteUserRepository(dbConn)
	sessionRepo := repository.NewSQLiteSessionRepository(dbConn)
	
	cfg := &config.Config{
		SessionTime:      time.Hour,
		AppEncryptionKey: make([]byte, 32),
		LockoutThreshold: 3,
		LockoutDuration:  time.Minute * 15,
	}
	
	sessionService := session.NewSessionService(sessionRepo, cfg)
	totpService := totp.NewTOTPService()
	authService := auth.NewAuthService(userRepo, sessionService, totpService, cfg.AppEncryptionKey, cfg)

	ctx := context.Background()
	name := "Test User"
	in := auth.RegisterInput{
		Username: "testuser",
		Password: "Password123!",
		Name:     &name,
	}

	// First registration should succeed
	err := authService.Register(ctx, in)
	if err != nil {
		t.Fatalf("expected successful first registration, got: %v", err)
	}

	// Second registration with the same username should fail
	err = authService.Register(ctx, in)
	if !errors.Is(err, auth.ErrUserExists) {
		t.Errorf("expected ErrUserExists, got: %v", err)
	}
}

func TestIntegration_LockoutThresholdTriggering(t *testing.T) {
	dbConn := setupTestDB(t)
	defer dbConn.Close()

	userRepo := repository.NewSQLiteUserRepository(dbConn)
	sessionRepo := repository.NewSQLiteSessionRepository(dbConn)
	
	cfg := &config.Config{
		SessionTime:      time.Hour,
		AppEncryptionKey: make([]byte, 32),
		LockoutThreshold: 3,
		LockoutDuration:  time.Minute * 15,
	}
	
	sessionService := session.NewSessionService(sessionRepo, cfg)
	totpService := totp.NewTOTPService()
	authService := auth.NewAuthService(userRepo, sessionService, totpService, cfg.AppEncryptionKey, cfg)

	ctx := context.Background()
	_ = authService.Register(ctx, auth.RegisterInput{
		Username: "lockoutuser",
		Password: "CorrectPassword123!",
	})

	// Fail 2 times safely
	for i := 0; i < 2; i++ {
		_, err := authService.Login(ctx, "lockoutuser", "WrongPassword123!")
		if !errors.Is(err, auth.ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials on attempt %d, got: %v", i+1, err)
		}
	}

	// 3rd attempt should trigger the lockout and return ErrAccountLocked immediately
	_, err := authService.Login(ctx, "lockoutuser", "WrongPassword123!")
	if !errors.Is(err, auth.ErrAccountLocked) {
		t.Fatalf("expected ErrAccountLocked on attempt 3, got: %v", err)
	}

	// 4th attempt should return ErrAccountLocked immediately, even with correct password
	_, err = authService.Login(ctx, "lockoutuser", "CorrectPassword123!")
	if !errors.Is(err, auth.ErrAccountLocked) {
		t.Errorf("expected ErrAccountLocked, got: %v", err)
	}
}

func TestIntegration_SessionExpiryHonoredByAuthGuard(t *testing.T) {
	dbConn := setupTestDB(t)
	defer dbConn.Close()

	userRepo := repository.NewSQLiteUserRepository(dbConn)
	sessionRepo := repository.NewSQLiteSessionRepository(dbConn)
	
	cfg := &config.Config{
		SessionTime:      -time.Hour, // Create sessions that are instantly expired
		AppEncryptionKey: make([]byte, 32),
	}
	
	sessionService := session.NewSessionService(sessionRepo, cfg)
	totpService := totp.NewTOTPService()
	authService := auth.NewAuthService(userRepo, sessionService, totpService, cfg.AppEncryptionKey, cfg)
	authGuard := session.NewAuthGuard(sessionRepo, userRepo)

	ctx := context.Background()
	_ = authService.Register(ctx, auth.RegisterInput{
		Username: "expiryuser",
		Password: "Password123!",
	})

	res, err := authService.Login(ctx, "expiryuser", "Password123!")
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}

	appState := &state.AppState{
		SessionToken: res.SessionToken,
	}

	// Because cfg.SessionTime is negative, the session was created already expired in the database
	_, err = authGuard.Require(ctx, appState)
	if !errors.Is(err, session.ErrSessionExpired) {
		t.Errorf("expected ErrSessionExpired for a token that is past its ExpiresAt, got: %v", err)
	}
}

func TestIntegration_Enable2FA_FailedVerification_NoSecretStored(t *testing.T) {
	dbConn := setupTestDB(t)
	defer dbConn.Close()

	userRepo := repository.NewSQLiteUserRepository(dbConn)
	sessionRepo := repository.NewSQLiteSessionRepository(dbConn)
	
	cfg := &config.Config{
		SessionTime:      time.Hour,
		AppEncryptionKey: make([]byte, 32),
	}
	
	sessionService := session.NewSessionService(sessionRepo, cfg)
	totpService := totp.NewTOTPService()
	authService := auth.NewAuthService(userRepo, sessionService, totpService, cfg.AppEncryptionKey, cfg)
	enrollmentService := totp.NewEnrollmentService(userRepo, sessionService, totpService, cfg.AppEncryptionKey)

	ctx := context.Background()
	_ = authService.Register(ctx, auth.RegisterInput{
		Username: "totpuser",
		Password: "Password123!",
	})

	res, err := authService.Login(ctx, "totpuser", "Password123!")
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}

	// 1. Generate Secret
	rawSecret, _, err := totpService.GenerateSecret("totpuser")
	if err != nil {
		t.Fatalf("failed to generate secret: %v", err)
	}

	// 2. Attempt ConfirmEnrollment with an explicitly invalid code
	err = enrollmentService.ConfirmEnrollment(ctx, res.User.ID, res.SessionToken, rawSecret, "000000")
	if !errors.Is(err, totp.ErrInvalidTOTP) {
		t.Fatalf("expected ErrInvalidTOTP, got: %v", err)
	}

	// 3. Verify user in DB still has MFAEnabled = false and MFASecretEnc = nil
	u, err := userRepo.GetByID(ctx, res.User.ID)
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}

	if u.MFAEnabled {
		t.Error("expected MFAEnabled to be false after failed enrollment")
	}

	if u.MFASecretEnc != nil {
		t.Error("expected MFASecretEnc to be nil after failed enrollment")
	}
}
