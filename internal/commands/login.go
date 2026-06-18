package commands

import (
	"context"
	"strings"

	"github.com/chzyer/readline"
	"osto-auth-cli/internal/auth"
	"osto-auth-cli/internal/state"
	"osto-auth-cli/internal/style"
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
				return loginErr
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
					return totpErr
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
