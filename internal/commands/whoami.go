package commands

import (
	"context"
	"errors"
	"fmt"
	"osto-auth-cli/internal/session"
	"osto-auth-cli/internal/state"
	"osto-auth-cli/internal/style"
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
					style.Error("Not authenticated.")
				} else if errors.Is(err, session.ErrSessionExpired) {
					style.Error("Session expired.")
				} else {
					style.Error("Failed to verify session: %v", err)
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

			fmt.Println("User Profile:")
			fmt.Printf("  %-15s %s\n", "Username", u.Username)
			fmt.Printf("  %-15s %s\n", "Created At", u.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("  %-15s %t\n", "2FA Enabled", u.MFAEnabled)
			fmt.Printf("  %-15s %s\n", "Last Login", lastLogin)
			fmt.Printf("  %-15s %s\n", "Session Exp", authSession.ExpiresAt.Format("2006-01-02 15:04:05"))
			return nil
		},
	}
}
