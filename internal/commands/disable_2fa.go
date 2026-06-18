package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/chzyer/readline"
	"osto-auth-cli/internal/session"
	"osto-auth-cli/internal/state"
	"osto-auth-cli/internal/style"
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
					style.Error("Not authenticated.")
				} else if errors.Is(err, session.ErrSessionExpired) {
					style.Error("Session expired.")
				} else {
					style.Error("Failed to verify session: %v", err)
				}
				return nil
			}

			if !authSession.User.MFAEnabled {
				style.Error("2FA is not enabled.")
				return nil
			}

			var success bool
			err = PromptWithRetries(rl, "Enter the current 6-digit code to disable 2FA: ", false, func(code string) error {
				code = strings.TrimSpace(code)
				disableErr := enrollmentService.Disable(ctx, authSession.User.ID, s.SessionToken, authSession.User.MFASecretEnc, code)
				if disableErr != nil {
					if errors.Is(disableErr, totp.ErrInvalidTOTP) {
						return errors.New("Invalid code")
					}
					return fmt.Errorf("Failed to disable 2FA: %v", disableErr)
				}
				success = true
				return nil
			})

			if err != nil || !success {
				style.Warn("Disablement aborted.")
				return nil
			}

			s.Clear()
			style.OK("2FA disabled successfully. Please log in again.")
			return nil
		},
	}
}
