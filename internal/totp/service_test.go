package totp_test

import (
	"strings"
	"testing"
	"time"

	"github.com/pquerna/otp"
	pquerna_totp "github.com/pquerna/otp/totp"
	"osto-auth-cli/internal/totp"
)

func TestTOTPService(t *testing.T) {
	service := totp.NewTOTPService()

	t.Run("GenerateSecret creates valid payload", func(t *testing.T) {
		secret, qr, err := service.GenerateSecret("testuser")
		if err != nil {
			t.Fatalf("failed to generate secret: %v", err)
		}

		if secret == "" {
			t.Error("expected non-empty secret")
		}

		if !strings.Contains(qr, "\n") {
			t.Error("expected QR string to contain multiple lines")
		}
	})

	t.Run("Validate accepts valid code generated at time.Now()", func(t *testing.T) {
		secret, _, err := service.GenerateSecret("testuser")
		if err != nil {
			t.Fatalf("failed to generate secret: %v", err)
		}

		// Generate a valid code for right now
		code, err := pquerna_totp.GenerateCodeCustom(secret, time.Now().UTC(), pquerna_totp.ValidateOpts{
			Period:    30,
			Skew:      1,
			Digits:    otp.DigitsSix,
			Algorithm: otp.AlgorithmSHA1,
		})
		if err != nil {
			t.Fatalf("failed to generate test code: %v", err)
		}

		if !service.Validate(secret, code) {
			t.Error("expected Validate to accept a correctly generated code for current time")
		}
	})

	t.Run("Validate rejects code outside configured skew", func(t *testing.T) {
		secret, _, err := service.GenerateSecret("testuser")
		if err != nil {
			t.Fatalf("failed to generate secret: %v", err)
		}

		// Skew in our configuration is 1 (±30s). Generating a code from 3 minutes ago
		// means it's strictly outside the acceptable time window.
		oldTime := time.Now().UTC().Add(-3 * time.Minute)
		code, err := pquerna_totp.GenerateCodeCustom(secret, oldTime, pquerna_totp.ValidateOpts{
			Period:    30,
			Skew:      1,
			Digits:    otp.DigitsSix,
			Algorithm: otp.AlgorithmSHA1,
		})
		if err != nil {
			t.Fatalf("failed to generate test code: %v", err)
		}

		if service.Validate(secret, code) {
			t.Error("expected Validate to reject a code generated 3 minutes ago (outside skew)")
		}
	})
}
