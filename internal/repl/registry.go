package repl

import (
	"sort"

	"osto-auth-cli/internal/commands"
)

type Registry struct {
	commands map[string]*commands.Command
}

func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]*commands.Command),
	}
}

func (r *Registry) Register(cmd *commands.Command) {
	r.commands[cmd.Name] = cmd
}

func (r *Registry) Get(name string) (*commands.Command, bool) {
	cmd, exists := r.commands[name]
	return cmd, exists
}

func (r *Registry) All() []*commands.Command {
	cmds := make([]*commands.Command, 0, len(r.commands))
	for _, cmd := range r.commands {
		cmds = append(cmds, cmd)
	}

	sort.Slice(cmds, func(i, j int) bool {
		return cmds[i].Name < cmds[j].Name
	})

	return cmds
}
