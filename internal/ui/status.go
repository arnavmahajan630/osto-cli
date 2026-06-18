package ui

import (
	"fmt"
	"osto-auth-cli/internal/state"
)

func StatusLine(s *state.AppState) string {
	if s.IsAuthenticated() {
		mfa := "Disabled"
		if s.CurrentUser.MFAEnabled {
			mfa = "Enabled"
		}
		return fmt.Sprintf("%s | MFA: %s", s.CurrentUser.Username, mfa)
	}
	return "Guest | MFA: N/A"
}
