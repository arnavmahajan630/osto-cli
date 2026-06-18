package repl

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/chzyer/readline"
	"osto-auth-cli/internal/commands"
	"osto-auth-cli/internal/state"
	"osto-auth-cli/internal/style"
)

type REPL struct {
	state     *state.AppState
	preLogin  *Registry
	postLogin *Registry
	rl        *readline.Instance
}

func NewREPL(s *state.AppState, preLogin *Registry, postLogin *Registry, rl *readline.Instance) *REPL {
	return &REPL{
		state:     s,
		preLogin:  preLogin,
		postLogin: postLogin,
		rl:        rl,
	}
}

func (r *REPL) getActiveRegistry() *Registry {
	if r.state.IsAuthenticated() {
		return r.postLogin
	}
	return r.preLogin
}

func (r *REPL) updateAutoComplete() {
	var completerItems []readline.PrefixCompleterInterface
	for _, cmd := range r.getActiveRegistry().All() {
		completerItems = append(completerItems, readline.PcItem(cmd.Name))
	}
	completer := readline.NewPrefixCompleter(completerItems...)
	r.rl.Config.AutoComplete = completer
}

func (r *REPL) Run(sessionRevoker func()) error {
	consecutiveInterrupts := 0

	for {
		r.updateAutoComplete()

		if r.state.IsAuthenticated() {
			r.rl.SetPrompt(fmt.Sprintf("osto(%s)> ", r.state.CurrentUser.Username))
		} else {
			r.rl.SetPrompt("osto> ")
		}

		line, err := r.rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				consecutiveInterrupts++
				if consecutiveInterrupts == 1 {
					style.Info("Type 'exit' or press Ctrl+C again to quit.")
					continue
				} else {
					if r.state.IsAuthenticated() && sessionRevoker != nil {
						sessionRevoker()
					}
					return nil
				}
			} else if err == io.EOF {
				if r.state.IsAuthenticated() && sessionRevoker != nil {
					sessionRevoker()
				}
				break
			}
			return err
		}
		consecutiveInterrupts = 0

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		cmdName := parts[0]
		args := parts[1:]

		// Handle Aliases
		switch cmdName {
		case "?":
			cmdName = "help"
		case "q", "quit":
			cmdName = "exit"
		}

		activeRegistry := r.getActiveRegistry()
		cmd, exists := activeRegistry.Get(cmdName)
		if !exists {
			style.Error("Unknown command: %s", cmdName)
			continue
		}

		err = cmd.Handler(r.state, args)
		if err != nil {
			if errors.Is(err, commands.ErrExit) {
				if r.state.IsAuthenticated() && sessionRevoker != nil {
					sessionRevoker()
				}
				break
			}
			style.Error("%v", err)
		}
	}

	return nil
}
