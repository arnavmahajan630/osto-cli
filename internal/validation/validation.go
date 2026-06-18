package validation

import (
	"regexp"
	"time"
	"unicode"
	"unicode/utf8"
)

// ValidationError represents a user-facing validation failure.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)
)

// Username validates that the username is 3-32 characters long,
// and consists only of alphanumeric characters and underscores.
func Username(s string) error {
	if !usernameRegex.MatchString(s) {
		return &ValidationError{Message: "Username must be 3-32 characters long and contain only letters, numbers, and underscores."}
	}
	return nil
}

// Password validates that the password is 8-72 characters long,
// does not exceed 72 bytes (bcrypt limitation), and contains at least one letter and one digit.
func Password(s string) error {
	if len(s) > 72 {
		return &ValidationError{Message: "Password cannot exceed 72 bytes."}
	}
	if utf8.RuneCountInString(s) < 8 {
		return &ValidationError{Message: "Password must be at least 8 characters long."}
	}

	hasLetter := false
	hasDigit := false

	for _, r := range s {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}

	if !hasLetter {
		return &ValidationError{Message: "Password must contain at least one letter."}
	}
	if !hasDigit {
		return &ValidationError{Message: "Password must contain at least one digit."}
	}

	return nil
}

// Date validates an optional birth date string format (YYYY-MM-DD).
func Date(s string) error {
	if s == "" {
		return nil
	}
	// "2006-01-02" is the reference layout for YYYY-MM-DD in Go
	if _, err := time.Parse("2006-01-02", s); err != nil {
		return &ValidationError{Message: "Date must be in YYYY-MM-DD format."}
	}
	return nil
}

