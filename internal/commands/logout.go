package commands

import (
	"fmt"
	"osto-auth-cli/internal/state"
)

func NewLogoutCommand() *Command {
	return &Command{
		Name:  "logout",
		Usage: "logout",
		Help:  "Log out of the current session",
		Handler: func(s *state.AppState, args []string) error {
			if !s.IsAuthenticated() {
				fmt.Println("[ERROR] Not authenticated.")
				return nil
			}
			s.Clear()
			fmt.Println("[OK] Logged out.")
			return nil
		},
	}
}
