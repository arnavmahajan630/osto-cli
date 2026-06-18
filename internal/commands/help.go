package commands

import (
	"fmt"
	"osto-auth-cli/internal/state"
)

func NewHelpCommand(getCommands func() []*Command) *Command {
	return &Command{
		Name:  "help",
		Usage: "help [command]",
		Help:  "Lists all available commands or prints detailed help for a specific command",
		Handler: func(state *state.AppState, args []string) error {
			cmds := getCommands()

			if len(args) == 0 {
				fmt.Println("Available commands:")
				for _, cmd := range cmds {
					fmt.Printf("  %-15s - %s\n", cmd.Name, cmd.Help)
				}
				return nil
			}

			target := args[0]
			for _, cmd := range cmds {
				if cmd.Name == target {
					fmt.Printf("Usage: %s\n", cmd.Usage)
					fmt.Printf("Help:  %s\n", cmd.Help)
					return nil
				}
			}

			fmt.Printf("[ERROR] Unknown command: %s\n", target)
			return nil
		},
	}
}
