package secure

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

var ErrInvalidKeyLength = errors.New("encryption key must be exactly 32 bytes")
var ErrInvalidPayload = errors.New("invalid or corrupted encrypted payload")

// Encrypt encrypts the plaintext using AES-256-GCM.
// The resulting payload format is base64( nonce || ciphertext ).
func Encrypt(plaintext []byte, key []byte) (string, error) {
	if len(key) != 32 {
		return "", ErrInvalidKeyLength
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Seal appends the ciphertext to the nonce, producing exactly nonce || ciphertext
	sealed := gcm.Seal(nonce, nonce, plaintext, nil)

	return base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt reverses the Encrypt process, returning the original plaintext.
func Decrypt(b64 string, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, ErrInvalidKeyLength
	}

	payload, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(payload) < nonceSize {
		return nil, ErrInvalidPayload
	}

	nonce := payload[:nonceSize]
	ciphertext := payload[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
