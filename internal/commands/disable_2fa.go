package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/chzyer/readline"
	"osto-auth-cli/internal/session"
	"osto-auth-cli/internal/state"
	"osto-auth-cli/internal/totp"
)

func NewDisable2FACommand(
	rl *readline.Instance,
	authGuard session.AuthGuard,
	enrollmentService *totp.EnrollmentService,
) *Command {
	return &Command{
		Name:  "disable-2fa",
		Usage: "disable-2fa",
		Help:  "Disable TOTP two-factor authentication",
		Handler: func(s *state.AppState, args []string) error {
			ctx := context.Background()
			authSession, err := authGuard.Require(ctx, s)
			if err != nil {
				if errors.Is(err, session.ErrNotAuthenticated) {
					fmt.Println("[ERROR] Not authenticated.")
				} else if errors.Is(err, session.ErrSessionExpired) {
					fmt.Println("[ERROR] Session expired.")
				} else {
					fmt.Printf("[ERROR] Failed to verify session: %v\n", err)
				}
				return nil
			}

			if !authSession.User.MFAEnabled {
				fmt.Println("[ERROR] 2FA is not enabled.")
				return nil
			}

			oldPrompt := rl.Config.Prompt
			rl.SetPrompt("Enter the current 6-digit code to disable 2FA: ")
			code, err := rl.Readline()
			rl.SetPrompt(oldPrompt)

			if err != nil {
				if err == readline.ErrInterrupt {
					fmt.Println("\n[ERROR] Disablement aborted.")
				}
				return err
			}

			code = strings.TrimSpace(code)

			err = enrollmentService.Disable(ctx, authSession.User.ID, s.SessionToken, authSession.User.MFASecretEnc, code)
			if err != nil {
				if errors.Is(err, totp.ErrInvalidTOTP) {
					fmt.Println("[ERROR] Invalid code. Disablement aborted.")
				} else {
					fmt.Printf("[ERROR] Failed to disable 2FA: %v\n", err)
				}
				return nil
			}

			s.Clear()
			fmt.Println("[OK] 2FA disabled successfully. Please log in again.")
			return nil
		},
	}
}
