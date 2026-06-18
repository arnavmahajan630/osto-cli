package commands

import (
	"context"
	"errors"
	"fmt"
	"osto-auth-cli/internal/session"
	"osto-auth-cli/internal/state"
)

func NewWhoamiCommand(guard session.AuthGuard) *Command {
	return &Command{
		Name:  "whoami",
		Usage: "whoami",
		Help:  "Display current authenticated user",
		Handler: func(s *state.AppState, args []string) error {
			authSession, err := guard.Require(context.Background(), s)
			if err != nil {
				if errors.Is(err, session.ErrNotAuthenticated) {
					fmt.Println("[ERROR] Not authenticated.")
				} else if errors.Is(err, session.ErrSessionExpired) {
					fmt.Println("[ERROR] Session expired.")
				} else {
					fmt.Printf("[ERROR] Failed to verify session: %v\n", err)
				}
				return nil
			}

			u := authSession.User
			var lastLogin string
			if u.LastLoginAt != nil {
				lastLogin = u.LastLoginAt.Format("2006-01-02 15:04:05")
			} else {
				lastLogin = "Never"
			}

			fmt.Printf("Username:       %s\n", u.Username)
			fmt.Printf("Created At:     %s\n", u.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("MFA Enabled:    %t\n", u.MFAEnabled)
			fmt.Printf("Last Login:     %s\n", lastLogin)
			fmt.Printf("Session Expiry: %s\n", authSession.ExpiresAt.Format("2006-01-02 15:04:05"))
			return nil
		},
	}
}
