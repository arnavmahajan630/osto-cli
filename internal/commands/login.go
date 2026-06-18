package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/chzyer/readline"
	"osto-auth-cli/internal/auth"
	"osto-auth-cli/internal/state"
	"osto-auth-cli/internal/style"
	"osto-auth-cli/internal/totp"
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

			var result *auth.LoginResult
			err = PromptWithRetries(rl, "Password: ", true, func(pass string) error {
				var loginErr error
				result, loginErr = authService.Login(context.Background(), username, pass)
				if loginErr != nil {
					var lockedErr *auth.ErrorAccountLocked
					if errors.As(loginErr, &lockedErr) {
						return loginErr // Propagate to breakout
					} else if errors.Is(loginErr, auth.ErrInvalidCredentials) {
						return errors.New("Invalid credentials.")
					}
					return fmt.Errorf("Login failed: %v", loginErr)
				}
				return nil
			})

			if err != nil {
				return nil // Aborted or locked
			}
			if result == nil {
				return nil // Exhausted retries
			}

			if result.RequiresTOTP {
				var token string
				err = PromptWithRetries(rl, "Enter 6-digit TOTP code: ", false, func(code string) error {
					code = strings.TrimSpace(code)
					var totpErr error
					token, totpErr = authService.VerifyTOTPAndCreateSession(context.Background(), result.User.ID, code)
					if totpErr != nil {
						var lockedErr *auth.ErrorAccountLocked
						if errors.As(totpErr, &lockedErr) {
							return totpErr
						} else if errors.Is(totpErr, totp.ErrInvalidTOTP) {
							return errors.New("Invalid TOTP code.")
						}
						return fmt.Errorf("TOTP verification failed: %v", totpErr)
					}
					return nil
				})
				
				if err != nil || token == "" {
					return nil // Aborted, locked, or exhausted
				}
				result.SessionToken = token
			}

			s.SessionToken = result.SessionToken
			s.CurrentUser = result.User

			style.OK("Logged in as %s.", s.CurrentUser.Username)
			style.Separator()
			return nil
		},
	}
}
