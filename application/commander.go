package application

import "github.com/spf13/cobra"

type Commander interface {
	Initialize(f []func(c *cobra.Command)) *cobra.Command
}

type Command struct {
	Command     *cobra.Command
	SubCommands []Commander
	Configure   func(c *cobra.Command)
}

func (c Command) Initialize(f []func(cmd *cobra.Command)) *cobra.Command {
	// Run the initialization function passed through Initialize
	if f != nil {
		for _, init := range f {
			init(c.Command)
		}
	}

	// Run the Command Configuration function
	if c.Configure != nil {
		c.Configure(c.Command)
	}

	// Pass the initialization function to the subcommands
	for _, sub := range c.SubCommands {
		c.Command.AddCommand(sub.Initialize(f))
	}
	return c.Command
}
