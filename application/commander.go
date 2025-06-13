package application

import "github.com/spf13/cobra"

type Commander interface {
	Initialize(f func(c *cobra.Command)) *cobra.Command
}

type Command struct {
	Command     *cobra.Command
	SubCommands []Commander
	Configure   func(c *cobra.Command)
}

func (c Command) Initialize(f func(cmd *cobra.Command)) *cobra.Command {
	if f != nil {
		f(c.Command)
	}

	if c.Configure != nil {
		c.Configure(c.Command)
	}

	for _, sub := range c.SubCommands {
		c.Command.AddCommand(sub.Initialize(f))
	}
	return c.Command
}
