package validation

import (
	"strings"
	"testing"
)

func TestUsername(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"Valid Short", "abc", false},
		{"Valid Long", strings.Repeat("a", 32), false},
		{"Valid with Numbers and Underscores", "user_123", false},
		{"Too Short", "ab", true},
		{"Too Long", strings.Repeat("a", 33), true},
		{"Contains Space", "user name", true},
		{"Contains Invalid Char", "user@name", true},
		{"Empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Username(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Username() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPassword(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"Valid Minimum", "abcdefg1", false},
		{"Valid Long", strings.Repeat("a", 71) + "1", false},
		{"Valid with Specials", "Passw0rd!", false},
		{"Too Short", "pass12", true},
		{"No Digit", "password", true},
		{"No Letter", "12345678", true},
		{"Exactly 72 Bytes", strings.Repeat("a", 71) + "1", false},
		{"Exceeds 72 Bytes (ASCII)", strings.Repeat("a", 72) + "1", true},
		{"Exceeds 72 Bytes (Multi-byte)", strings.Repeat("é", 36) + "1", true}, // 'é' is 2 bytes in UTF-8
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Password(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Password() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"Valid Date", "1990-01-01", false},
		{"Empty String (Optional)", "", false},
		{"Wrong Format (MM-DD-YYYY)", "01-01-1990", true},
		{"Wrong Format (YYYY/MM/DD)", "1990/01/01", true},
		{"Invalid Month", "1990-13-01", true},
		{"Invalid Day", "1990-01-32", true},
		{"Random String", "not-a-date", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Date(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Date() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
