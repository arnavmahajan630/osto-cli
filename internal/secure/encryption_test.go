package secure_test

import (
	"encoding/base64"
	"errors"
	"testing"

	"osto-auth-cli/internal/secure"
)

func TestEncryption(t *testing.T) {
	validKey := make([]byte, 32)
	for i := range validKey {
		validKey[i] = byte(i)
	}

	t.Run("round-trip success", func(t *testing.T) {
		plaintext := []byte("top secret payload")
		
		b64Payload, err := secure.Encrypt(plaintext, validKey)
		if err != nil {
			t.Fatalf("failed to encrypt: %v", err)
		}

		if b64Payload == "" {
			t.Fatal("expected non-empty encrypted payload")
		}

		decrypted, err := secure.Decrypt(b64Payload, validKey)
		if err != nil {
			t.Fatalf("failed to decrypt: %v", err)
		}

		if string(decrypted) != string(plaintext) {
			t.Errorf("expected decrypted text %q, got %q", string(plaintext), string(decrypted))
		}
	})

	t.Run("invalid key length rejection", func(t *testing.T) {
		plaintext := []byte("hello")
		shortKey := []byte("too_short")
		
		_, err := secure.Encrypt(plaintext, shortKey)
		if !errors.Is(err, secure.ErrInvalidKeyLength) {
			t.Errorf("Encrypt: expected %v, got %v", secure.ErrInvalidKeyLength, err)
		}

		_, err = secure.Decrypt("some-payload", shortKey)
		if !errors.Is(err, secure.ErrInvalidKeyLength) {
			t.Errorf("Decrypt: expected %v, got %v", secure.ErrInvalidKeyLength, err)
		}
	})

	t.Run("tampered ciphertext failure", func(t *testing.T) {
		plaintext := []byte("important data")
		
		b64Payload, err := secure.Encrypt(plaintext, validKey)
		if err != nil {
			t.Fatalf("failed to encrypt: %v", err)
		}

		// Decode the valid base64 payload to tamper with it
		rawPayload, err := base64.StdEncoding.DecodeString(b64Payload)
		if err != nil {
			t.Fatalf("failed to decode base64 payload: %v", err)
		}

		// Tamper with the last byte (part of ciphertext/auth tag)
		rawPayload[len(rawPayload)-1] ^= 0xFF

		tamperedB64 := base64.StdEncoding.EncodeToString(rawPayload)

		_, err = secure.Decrypt(tamperedB64, validKey)
		if err == nil {
			t.Fatal("expected Decrypt to fail on tampered payload, but it succeeded")
		}
	})
}
