package application

import (
	"errors"

	"github.com/spf13/cobra"
)

type Builder struct {
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

func (b Builder) Build() (*cobra.Command, error) {
	var err error
	if err = b.Validate(); err != nil {
		return nil, err
	}

	appName = b.Name

	var long string
	if b.Banner != "" {
		long = b.Banner + "\n" + b.Title
		banner = b.Banner
	} else {
		long = b.Title
	}

	cmd := &cobra.Command{
		Use:                b.Name,
		Short:              b.Title,
		Long:               long,
		PersistentPreRunE:  persistentPreRunFuncE,
		PersistentPostRunE: persistentPostRunFuncE,
		RunE:               HelpFuncE,
		SilenceErrors:      true,
		SilenceUsage:       true,
	}

	if b.OverrideRunE != nil {
		cmd.RunE = b.OverrideRunE
	}

	for _, preFuncE := range b.PersistentPreRunE {
		b.RegisterPersistentPreRunE(preFuncE)
	}

	for _, postFuncE := range b.PersistentPostRunE {
		b.RegisterPersistentPostRunE(postFuncE)
	}

	for _, subcommand := range b.SubCommands {
		cmd.AddCommand(subcommand.Initialize(b.SubCommandInitializers))
	}

	configureVersionFlag(cmd)                             // Configure app for version information
	configureOutputFlags(cmd)                             // Configure verbosity
	configureLoggingFlags(cmd)                            // Configure logging
	cmd.PersistentFlags().SetNormalizeFunc(normalizeFunc) // normalize persistent flags

	return cmd, nil
}

func (b Builder) RegisterCommand(cmd Commander) {
	b.SubCommands = append(b.SubCommands, cmd)
}

func (b Builder) RegisterCommands(cmds []Commander) {
	b.SubCommands = append(b.SubCommands, cmds...)
}

func (b Builder) RegisterPersistentPreRunE(f func(cmd *cobra.Command, args []string) error) {
	persistentPreRunE = append(persistentPreRunE, f)
}

func (b Builder) RegisterPersistentPostRunE(f func(cmd *cobra.Command, args []string) error) {
	persistentPostRunE = append(persistentPostRunE, f)
}

func (b Builder) Validate() error {
	if b.Name == "" {
		return errors.New("name is required")
	}
	if b.Title == "" {
		return errors.New("title is required")
	}
	return nil
}
