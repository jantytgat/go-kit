package application

import (
	"bufio"
	"errors"
	"os"

	"github.com/spf13/cobra"

	"github.com/jantytgat/go-kit/shellquote"
)

type Builder struct {
	Name                   string
	Title                  string
	Banner                 string
	OverrideRunE           func(cmd *cobra.Command, args []string) error
	ConfigureRoot          func(cmd *cobra.Command)
	PersistentPreRunE      []func(cmd *cobra.Command, args []string) error // collection of PreRunE functions
	PersistentPostRunE     []func(cmd *cobra.Command, args []string) error // collection of PostRunE functions
	SubCommands            []Commander
	SubCommandInitializers []func(cmd *cobra.Command)
	ParseArgsFromStdin     bool
	ValidArgs              []string
}

func (b Builder) updateArgsFromStdin() error {
	var err error
	var fi os.FileInfo
	if fi, err = os.Stdin.Stat(); err != nil {
		return err
	}
	if fi.Mode()&os.ModeNamedPipe != 0 {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			var extraArgs []string
			if extraArgs, err = shellquote.Split(scanner.Text()); err != nil {
				return err
			}
			os.Args = append(os.Args, extraArgs...)
		}
		return nil
	}
	return nil
}

func (b Builder) build() (*cobra.Command, error) {
	var err error
	if err = b.Validate(); err != nil {
		return nil, err
	}

	// Update arguments with input on os.Stdin
	if b.ParseArgsFromStdin {
		if err = b.updateArgsFromStdin(); err != nil {
			return nil, err
		}
	}

	appName = b.Name

	var long string
	if b.Banner != "" {
		long = b.Banner + "\n" + b.Title
		banner = b.Banner
	} else {
		long = b.Title
	}

	// Create default cobra.Command, then proceed with configuration of the command
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
	if b.ConfigureRoot != nil {
		b.ConfigureRoot(cmd)
	}

	// Override the default HelpFuncE if the root command should actually do something
	if b.OverrideRunE != nil {
		cmd.RunE = b.OverrideRunE
	}

	// Add PersistentPreRunE functions to the root command
	for _, preFuncE := range b.PersistentPreRunE {
		b.RegisterPersistentPreRunE(preFuncE)
	}

	// Add PersistentPostRunE functions to the root command
	for _, postFuncE := range b.PersistentPostRunE {
		b.RegisterPersistentPostRunE(postFuncE)
	}

	// Configure subcommands
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
