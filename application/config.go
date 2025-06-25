package application

import (
	"errors"

	"github.com/spf13/cobra"
)

type Configuration interface {
	BuildCommand() (*cobra.Command, error)
}
type Config struct {
	Name                   string
	Title                  string
	Banner                 string
	OverrideRunE           func(cmd *cobra.Command, args []string) error
	PersistentPreRunE      []func(cmd *cobra.Command, args []string) error // collection of PreRunE functions
	PersistentPostRunE     []func(cmd *cobra.Command, args []string) error // collection of PostRunE functions
	SubCommands            []Commander
	SubCommandInitializers []func(cmd *cobra.Command)
	ValidArgs              []string
}

func (c Config) BuildCommand() (*cobra.Command, error) {
	var err error
	if err = c.Validate(); err != nil {
		return nil, err
	}

	appName = c.Name

	var long string
	if c.Banner != "" {
		long = c.Banner + "\n" + c.Title
		banner = c.Banner
	} else {
		long = c.Title
	}

	cmd := &cobra.Command{
		Use:                c.Name,
		Short:              c.Title,
		Long:               long,
		PersistentPreRunE:  persistentPreRunFuncE,
		PersistentPostRunE: persistentPostRunFuncE,
		RunE:               HelpFuncE,
		SilenceErrors:      true,
		SilenceUsage:       true,
	}

	if c.OverrideRunE != nil {
		cmd.RunE = c.OverrideRunE
	}

	for _, subcommand := range c.SubCommands {
		cmd.AddCommand(subcommand.Initialize(c.SubCommandInitializers))
	}

	configureVersionFlag(cmd)                             // Configure app for version information
	configureOutputFlags(cmd)                             // Configure verbosity
	configureLoggingFlags(cmd)                            // Configure logging
	cmd.PersistentFlags().SetNormalizeFunc(normalizeFunc) // normalize persistent flags

	return cmd, nil
}

func (c Config) RegisterCommand(cmd Commander) {
	c.SubCommands = append(c.SubCommands, cmd)
}

func (c Config) RegisterCommands(cmds []Commander) {
	c.SubCommands = append(c.SubCommands, cmds...)
}

func (c Config) RegisterPersistentPreRunE(f func(cmd *cobra.Command, args []string) error) {
	persistentPreRunE = append(persistentPreRunE, f)
}

func (c Config) RegisterPersistentPostRunE(f func(cmd *cobra.Command, args []string) error) {
	persistentPostRunE = append(persistentPostRunE, f)
}

func (c Config) Validate() error {
	if c.Name == "" {
		return errors.New("name is required")
	}
	if c.Title == "" {
		return errors.New("title is required")
	}
	return nil
}
