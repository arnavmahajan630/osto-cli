package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/chzyer/readline"
	"osto-auth-cli/internal/auth"
	"osto-auth-cli/internal/session"
	"osto-auth-cli/internal/state"
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
					fmt.Println("[ERROR] Not authenticated.")
				} else if errors.Is(err, session.ErrSessionExpired) {
					fmt.Println("[ERROR] Session expired.")
				} else {
					fmt.Printf("[ERROR] Failed to verify session: %v\n", err)
				}
				return nil
			}

			if authSession.User.MFAEnabled {
				fmt.Println("[ERROR] 2FA is already enabled.")
				return nil
			}

			secret, qr, err := totpService.GenerateSecret(authSession.User.Username)
			if err != nil {
				fmt.Printf("[ERROR] Failed to generate 2FA secret: %v\n", err)
				return nil
			}

			fmt.Println("\nScan the QR code below with your authenticator app:")
			fmt.Println(qr)
			fmt.Println("\nOr enter this code manually:")
			fmt.Printf("Secret: %s\n\n", secret)

			oldPrompt := rl.Config.Prompt
			rl.SetPrompt("Enter the current 6-digit code to confirm setup: ")
			code, err := rl.Readline()
			rl.SetPrompt(oldPrompt)

			if err != nil {
				if err == readline.ErrInterrupt {
					fmt.Println("\n[ERROR] Setup aborted.")
				}
				return err
			}

			code = strings.TrimSpace(code)

			err = enrollmentService.ConfirmEnrollment(ctx, authSession.User.ID, s.SessionToken, secret, code)
			if err != nil {
				if errors.Is(err, auth.ErrInvalidTOTP) {
					fmt.Println("[ERROR] Invalid code. Setup aborted.")
				} else {
					fmt.Printf("[ERROR] Failed to enable 2FA: %v\n", err)
				}
				return nil
			}

			s.Clear()
			fmt.Println("[OK] 2FA enabled successfully. Please log in again.")
			return nil
		},
	}
}
