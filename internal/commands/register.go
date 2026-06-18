package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/chzyer/readline"
	"osto-auth-cli/internal/auth"
	"osto-auth-cli/internal/state"
	"osto-auth-cli/internal/validation"
)

func NewRegisterCommand(rl *readline.Instance, authService auth.AuthService) *Command {
	return &Command{
		Name:  "register",
		Usage: "register",
		Help:  "Create a new user account.",
		Handler: func(s *state.AppState, args []string) error {
			// Save old prompt
			oldPrompt := rl.Config.Prompt

			rl.SetPrompt("Username: ")
			username, err := rl.Readline()
			if err != nil {
				rl.SetPrompt(oldPrompt)
				return err
			}
			username = strings.TrimSpace(username)

			rl.SetPrompt("Name: ")
			nameStr, err := rl.Readline()
			if err != nil {
				rl.SetPrompt(oldPrompt)
				return err
			}
			nameStr = strings.TrimSpace(nameStr)
			var name *string
			if nameStr != "" {
				name = &nameStr
			}

			rl.SetPrompt("Birth Date (YYYY-MM-DD, optional): ")
			bdStr, err := rl.Readline()
			if err != nil {
				rl.SetPrompt(oldPrompt)
				return err
			}
			bdStr = strings.TrimSpace(bdStr)
			var birthDate *string
			if bdStr != "" {
				birthDate = &bdStr
			}

			passBytes, err := rl.ReadPassword("Password: ")
			if err != nil {
				rl.SetPrompt(oldPrompt)
				return err
			}
			password := string(passBytes)

			confirmBytes, err := rl.ReadPassword("Confirm Password: ")
			if err != nil {
				rl.SetPrompt(oldPrompt)
				return err
			}
			confirm := string(confirmBytes)

			// Restore old prompt
			rl.SetPrompt(oldPrompt)

			if password != confirm {
				fmt.Println("[ERROR] Passwords do not match.")
				return nil
			}

			input := auth.RegisterInput{
				Username:  username,
				Password:  password,
				Name:      name,
				BirthDate: birthDate,
			}

			err = authService.Register(context.Background(), input)
			if err != nil {
				var valErr *validation.ValidationError
				if errors.Is(err, auth.ErrUserExists) {
					fmt.Println("[ERROR] Username already exists.")
				} else if errors.As(err, &valErr) {
					fmt.Printf("[ERROR] %s\n", valErr.Message)
				} else {
					fmt.Println("[ERROR] Internal server error. Please try again.")
				}
				return nil
			}

			fmt.Println("[OK] Registered. You can now log in.")
			return nil
		},
	}
}
