package commands

import (
	"fmt"
	"osto-auth-cli/internal/state"
)

func NewRegisterCommand() *Command {
	return &Command{
		Name:  "register",
		Usage: "register",
		Help:  "Register a new user account",
		Handler: func(state *state.AppState, args []string) error {
			fmt.Println("[INFO] Not implemented yet.")
			return nil
		},
	}
}
