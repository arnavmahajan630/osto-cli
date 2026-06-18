package totp

import (
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
)

// TOTPService defines the contract for TOTP operations.
type TOTPService interface {
	GenerateSecret(accountName string) (secret string, asciiQR string, err error)
	Validate(secret, code string) bool
}

// DefaultTOTPService implements TOTPService according to the strict specs.
type DefaultTOTPService struct{}

// NewTOTPService creates a new TOTPService.
func NewTOTPService() *DefaultTOTPService {
	return &DefaultTOTPService{}
}

// GenerateSecret generates a TOTP secret and its ASCII QR code.
func (s *DefaultTOTPService) GenerateSecret(accountName string) (string, string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "osto-auth-cli",
		AccountName: accountName,
		Period:      30,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return "", "", err
	}

	q, err := qrcode.New(key.URL(), qrcode.Medium)
	if err != nil {
		return "", "", err
	}

	asciiQR := q.ToSmallString(false)
	return key.Secret(), asciiQR, nil
}

// Validate checks the provided code against the secret using the required skew.
func (s *DefaultTOTPService) Validate(secret, code string) bool {
	valid, _ := totp.ValidateCustom(code, secret, time.Now().UTC(), totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	return valid
}
