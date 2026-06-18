package commands

import (
	"fmt"
	"osto-auth-cli/internal/state"
)

func NewLoginCommand() *Command {
	return &Command{
		Name:  "login",
		Usage: "login",
		Help:  "Log in to an existing account",
		Handler: func(state *state.AppState, args []string) error {
			fmt.Println("[INFO] Not implemented yet.")
			return nil
		},
	}
}
