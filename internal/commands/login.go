package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/chzyer/readline"
	"osto-auth-cli/internal/auth"
	"osto-auth-cli/internal/state"
)

func NewLoginCommand(rl *readline.Instance, authService auth.AuthService) *Command {
	return &Command{
		Name:  "login",
		Usage: "login",
		Help:  "Log in to an existing account",
		Handler: func(s *state.AppState, args []string) error {
			oldPrompt := rl.Config.Prompt

			rl.SetPrompt("Username: ")
			username, err := rl.Readline()
			if err != nil {
				rl.SetPrompt(oldPrompt)
				return err
			}
			username = strings.TrimSpace(username)

			passBytes, err := rl.ReadPassword("Password: ")
			if err != nil {
				rl.SetPrompt(oldPrompt)
				return err
			}
			password := string(passBytes)

			rl.SetPrompt(oldPrompt)

			result, err := authService.Login(context.Background(), username, password)
			if err != nil {
				if errors.Is(err, auth.ErrInvalidCredentials) {
					fmt.Println("[ERROR] Invalid credentials.")
				} else {
					fmt.Printf("[ERROR] Login failed: %v\n", err)
				}
				return nil
			}

			s.SessionToken = result.SessionToken
			s.CurrentUser = result.User

			fmt.Printf("[OK] Logged in as %s.\n", s.CurrentUser.Username)
			return nil
		},
	}
}
