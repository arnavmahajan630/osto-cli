package commands

import (
	"context"
	"osto-auth-cli/internal/session"
	"osto-auth-cli/internal/state"
	"osto-auth-cli/internal/style"
)

func NewLogoutCommand(sessionService session.SessionService) *Command {
	return &Command{
		Name:  "logout",
		Usage: "logout",
		Help:  "Log out of the current session",
		Handler: func(s *state.AppState, args []string) error {
			if !s.IsAuthenticated() {
				style.Error("Not authenticated.")
				return nil
			}
			
			if err := sessionService.Revoke(context.Background(), s.SessionToken); err != nil {
				style.Error("Failed to revoke session: %v", err)
				return nil
			}

			s.Clear()
			style.Separator()
			style.OK("Logged out.")
			return nil
		},
	}
}
