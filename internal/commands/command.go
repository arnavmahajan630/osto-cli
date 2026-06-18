package commands

import "osto-auth-cli/internal/state"

type Command struct {
	Name    string
	Usage   string
	Help    string
	Handler func(state *state.AppState, args []string) error
}
