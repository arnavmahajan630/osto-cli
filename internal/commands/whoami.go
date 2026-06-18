package commands

import (
	"fmt"
	"osto-auth-cli/internal/state"
)

func NewWhoamiCommand() *Command {
	return &Command{
		Name:  "whoami",
		Usage: "whoami",
		Help:  "Display current authenticated user",
		Handler: func(s *state.AppState, args []string) error {
			if !s.IsAuthenticated() {
				fmt.Println("[ERROR] Not authenticated.")
				return nil
			}
			fmt.Printf("[INFO] Current user: %s (ID: %d)\n", s.CurrentUser.Username, s.CurrentUser.ID)
			return nil
		},
	}
}
