package commands

import (
	"fmt"
	"osto-auth-cli/internal/state"
)

func NewClearCommand() *Command {
	return &Command{
		Name:  "clear",
		Usage: "clear",
		Help:  "Clears the terminal screen",
		Handler: func(s *state.AppState, args []string) error {
			fmt.Print("\033[H\033[2J\033[3J")
			return nil
		},
	}
}
