package secure_test

import (
	"encoding/base64"
	"testing"

	"osto-auth-cli/internal/secure"
)

func TestToken(t *testing.T) {
	t.Run("NewSessionToken generates correctly", func(t *testing.T) {
		token1, err := secure.NewSessionToken()
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		token2, err := secure.NewSessionToken()
		if err != nil {
			t.Fatalf("failed to generate second token: %v", err)
		}

		if token1 == token2 {
			t.Fatal("expected NewSessionToken to generate unique tokens")
		}

		// 32 bytes base64url encoded with no padding should be 43 characters
		if len(token1) != 43 {
			t.Errorf("expected token length 43, got %d", len(token1))
		}

		// Verify it's valid base64url (no padding)
		decoded, err := base64.RawURLEncoding.DecodeString(token1)
		if err != nil {
			t.Errorf("failed to decode token as base64url: %v", err)
		}

		if len(decoded) != 32 {
			t.Errorf("expected 32 bytes of entropy, got %d", len(decoded))
		}
	})

	t.Run("HashToken determinism", func(t *testing.T) {
		raw := "test-raw-token"
		
		hash1 := secure.HashToken(raw)
		hash2 := secure.HashToken(raw)

		if hash1 == "" {
			t.Fatal("expected non-empty hash")
		}

		if hash1 != hash2 {
			t.Errorf("expected deterministic hashing: hash1=%s, hash2=%s", hash1, hash2)
		}
		
		// Ensure different inputs yield different hashes
		if hash1 == secure.HashToken("different-raw-token") {
			t.Fatal("expected different hashes for different inputs")
		}
	})
}
