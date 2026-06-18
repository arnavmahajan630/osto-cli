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

func NewEnable2FACommand(
	rl *readline.Instance,
	authGuard session.AuthGuard,
	totpService totp.TOTPService,
	enrollmentService *totp.EnrollmentService,
) *Command {
	return &Command{
		Name:  "enable-2fa",
		Usage: "enable-2fa",
		Help:  "Set up TOTP two-factor authentication",
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

			if authSession.User.MFAEnabled {
				style.Error("2FA is already enabled.")
				return nil
			}

			secret, qr, err := totpService.GenerateSecret(authSession.User.Username)
			if err != nil {
				style.Error("Failed to generate 2FA secret: %v", err)
				return nil
			}

			style.Info("Scan the QR code below with your authenticator app:\n\n%s\n", qr)
			style.Info("Or enter this code manually: %s\n", secret)

			var success bool
			err = PromptWithRetries(rl, "Enter the current 6-digit code to confirm setup: ", false, func(code string) error {
				code = strings.TrimSpace(code)
				confirmErr := enrollmentService.ConfirmEnrollment(ctx, authSession.User.ID, s.SessionToken, secret, code)
				if confirmErr != nil {
					if errors.Is(confirmErr, totp.ErrInvalidTOTP) {
						return errors.New("Invalid code")
					}
					return fmt.Errorf("Failed to enable 2FA: %v", confirmErr)
				}
				success = true
				return nil
			})

			if err != nil || !success {
				style.Warn("Setup aborted.")
				return nil
			}

			s.Clear()
			style.Separator()
			style.OK("2FA enabled successfully. Please log in again.")
			return nil
		},
	}
}
