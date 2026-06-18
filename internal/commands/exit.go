package commands

import (
	"errors"
	"osto-auth-cli/internal/state"
)

// ErrExit is returned by the exit command to signal the REPL to shut down.
var ErrExit = errors.New("exit")

func NewExitCommand() *Command {
	return &Command{
		Name:  "exit",
		Usage: "exit",
		Help:  "Exits the application cleanly",
		Handler: func(state *state.AppState, args []string) error {
			return ErrExit
		},
	}
}
