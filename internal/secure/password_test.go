package secure_test

import (
	"testing"

	"osto-auth-cli/internal/secure"
)

func TestPasswordHashing(t *testing.T) {
	t.Run("round-trip success", func(t *testing.T) {
		plain := "correcthorsebatterystaple123!"
		hash, err := secure.HashPassword(plain)
		if err != nil {
			t.Fatalf("failed to hash password: %v", err)
		}

		if hash == "" {
			t.Fatal("expected non-empty hash")
		}

		if !secure.VerifyPassword(hash, plain) {
			t.Error("expected VerifyPassword to return true for correct password")
		}
	})

	t.Run("verification failure with incorrect plaintext", func(t *testing.T) {
		plain := "my-secret"
		hash, err := secure.HashPassword(plain)
		if err != nil {
			t.Fatalf("failed to hash password: %v", err)
		}

		if secure.VerifyPassword(hash, "wrong-secret") {
			t.Error("expected VerifyPassword to return false for incorrect password")
		}
	})
}
