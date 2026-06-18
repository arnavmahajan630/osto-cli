package commands

import (
	"errors"

	"github.com/chzyer/readline"
	"osto-auth-cli/internal/auth"
	"osto-auth-cli/internal/style"
)

// PromptWithRetries prompts the user up to 3 times.
// If isPassword is true, it masks the input.
func PromptWithRetries(
	rl *readline.Instance,
	prompt string,
	isPassword bool,
	verifyFn func(input string) error,
) error {
	const maxAttempts = 3
	oldPrompt := rl.Config.Prompt

	for attempts := 1; attempts <= maxAttempts; attempts++ {
		var input string
		var err error

		if isPassword {
			var passBytes []byte
			passBytes, err = rl.ReadPassword(prompt)
			input = string(passBytes)
		} else {
			rl.SetPrompt(prompt)
			input, err = rl.Readline()
			rl.SetPrompt(oldPrompt)
		}

		if err != nil {
			if err == readline.ErrInterrupt {
				style.Error("Input aborted.")
			}
			return err
		}

		err = verifyFn(input)
		if err == nil {
			return nil
		}

		// Check for lockout
		var lockedErr *auth.ErrorAccountLocked
		if errors.As(err, &lockedErr) {
			style.Error("Account locked until %s", lockedErr.Until.Format("15:04"))
			return nil // gracefully stop retries, the error has been displayed
		}

		remaining := maxAttempts - attempts
		if remaining > 0 {
			// Instead of a generic error, we rely on the specific verify function
			// returning the error string, or we just print the error they returned
			style.Error("%v (%d attempt(s) remaining)", err, remaining)
		} else {
			style.Error("%v", err)
		}
	}

	return nil
}
