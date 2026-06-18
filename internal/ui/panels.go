package ui

import (
	"fmt"
	"strings"
	"time"

	"osto-auth-cli/internal/models"
)

type CommandDesc struct {
	Name     string
	Usage    string
	Help     string
	Category string
	Aliases  []string
}

// FormHeader prints a boxed header for guided flows.
func FormHeader(title string) {
	fmt.Println("┌─ " + title + " " + strings.Repeat("─", 36-len(title)) + "┐")
}

// FormFooter prints the bottom of the form box.
func FormFooter() {
	fmt.Println("└──────────────────────────────────────┘")
}

// SuccessBox prints a green success box.
func SuccessBox(title string, lines ...string) {
	fmt.Println("┌─ " + title + " " + strings.Repeat("─", 36-len(title)) + "┐")
	for _, line := range lines {
		fmt.Printf("│ %-36s │\n", line)
	}
	fmt.Println("└──────────────────────────────────────┘")
}

// ErrorBox prints a red error box.
func ErrorBox(title string, lines ...string) {
	fmt.Println("┌─ " + title + " " + strings.Repeat("─", 36-len(title)) + "┐")
	for _, line := range lines {
		fmt.Printf("│ %-36s │\n", line)
	}
	fmt.Println("└──────────────────────────────────────┘")
}

// DashboardPanel prints the welcome dashboard.
func DashboardPanel(user *models.User, expiresAt time.Time) {
	fmt.Println("┌─ Account Overview ───────────────────┐")
	fmt.Printf("│ User: %-30s │\n", user.Username)
	mfaStr := "Disabled"
	if user.MFAEnabled {
		mfaStr = "Enabled"
	}
	fmt.Printf("│ MFA: %-31s │\n", mfaStr)
	var lastLogin string
	if user.LastLoginAt != nil {
		lastLogin = user.LastLoginAt.Format("2006-01-02 15:04")
	} else {
		lastLogin = "Never"
	}
	fmt.Printf("│ Last Login: %-24s │\n", lastLogin)

	remaining := time.Until(expiresAt)
	if remaining < 0 {
		remaining = 0
	}
	fmt.Printf("│ Session Expires: %-19s │\n", formatDuration(remaining))
	fmt.Println("└──────────────────────────────────────┘")
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	m := d / time.Minute
	s := (d % time.Minute) / time.Second
	return fmt.Sprintf("%dm %ds", m, s)
}

// CommandDiscoveryPanel groups and displays commands by category.
func CommandDiscoveryPanel(cmds []CommandDesc) {
	fmt.Println("\nAvailable Commands")

	categories := make(map[string][]CommandDesc)
	for _, cmd := range cmds {
		cat := cmd.Category
		if cat == "" {
			cat = "General"
		}
		categories[cat] = append(categories[cat], cmd)
	}

	// Print in a stable order
	order := []string{"Authentication", "Account", "General"}
	for _, cat := range order {
		if catCmds, ok := categories[cat]; ok && len(catCmds) > 0 {
			fmt.Println("\n" + cat)
			for _, cmd := range catCmds {
				fmt.Printf("  %s\n", cmd.Name)
			}
		}
	}
	fmt.Println()
}

// HelpPanel displays detailed help for a specific command.
func HelpPanel(cmd CommandDesc) {
	fmt.Printf("Command      %s\n", cmd.Name)
	fmt.Printf("Description  %s\n", cmd.Help)
	aliases := strings.Join(cmd.Aliases, ", ")
	if aliases == "" {
		aliases = "None"
	}
	fmt.Printf("Aliases      %s\n", aliases)
	fmt.Printf("Usage        %s\n", cmd.Usage)
}
