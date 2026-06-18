package commands

import (
	"context"
	"fmt"
	"osto-auth-cli/internal/session"
	"osto-auth-cli/internal/state"
)

func NewLogoutCommand(sessionService session.SessionService) *Command {
	return &Command{
		Name:  "logout",
		Usage: "logout",
		Help:  "Log out of the current session",
		Handler: func(s *state.AppState, args []string) error {
			if !s.IsAuthenticated() {
				fmt.Println("[ERROR] Not authenticated.")
				return nil
			}
			
			if err := sessionService.Revoke(context.Background(), s.SessionToken); err != nil {
				fmt.Printf("[ERROR] Failed to revoke session: %v\n", err)
				return nil
			}

			s.Clear()
			fmt.Println("[OK] Logged out.")
			return nil
		},
	}
}
