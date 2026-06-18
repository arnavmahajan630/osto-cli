package repl

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/chzyer/readline"
	"osto-auth-cli/internal/commands"
	"osto-auth-cli/internal/state"
)

type REPL struct {
	state    *state.AppState
	registry *Registry
	rl       *readline.Instance
}

func NewREPL(s *state.AppState, r *Registry, rl *readline.Instance) *REPL {
	return &REPL{
		state:    s,
		registry: r,
		rl:       rl,
	}
}

func (r *REPL) Run() error {
	var completerItems []readline.PrefixCompleterInterface
	for _, cmd := range r.registry.All() {
		completerItems = append(completerItems, readline.PcItem(cmd.Name))
	}

	completer := readline.NewPrefixCompleter(completerItems...)
	r.rl.Config.AutoComplete = completer

	for {
		line, err := r.rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				continue
			} else if err == io.EOF {
				break
			}
			return err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		cmdName := parts[0]
		args := parts[1:]

		cmd, exists := r.registry.Get(cmdName)
		if !exists {
			fmt.Printf("[ERROR] Unknown command: %s\n", cmdName)
			continue
		}

		err = cmd.Handler(r.state, args)
		if err != nil {
			if errors.Is(err, commands.ErrExit) {
				break
			}
			// Command handlers should print their own user-facing errors (per CONTEXT),
			// but we can print a fallback here if they bubble up an unexpected error.
			fmt.Printf("[ERROR] %v\n", err)
		}
	}

	return nil
}
