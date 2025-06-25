package application

import (
	"errors"

	"github.com/spf13/cobra"
)

var (
	versionFull       string = "0.0.0-RUN"
	versionBranch     string
	versionTag        string
	versionCommit     string
	versionCommitDate string
	versionBuildDate  string
	versionMajor      string
	versionMinor      string
	versionPatch      string
	versionPrerelease string
)

var (
	banner string
)

type Configuration interface {
	GetRootCommand() (*cobra.Command, error)
}
type Config struct {
	Name                     string
	Title                    string
	Banner                   string
	OverrideRunE             func(cmd *cobra.Command, args []string) error
	PersistentPreRunE        []func(cmd *cobra.Command, args []string) error // collection of PreRunE functions
	PersistentPostRunE       []func(cmd *cobra.Command, args []string) error // collection of PostRunE functions
	SubCommands              []Commander
	SubCommandInitializeFunc func(cmd *cobra.Command)
	ValidArgs                []string
}

func (c Config) GetRootCommand() (*cobra.Command, error) {
	var err error
	if err = c.Validate(); err != nil {
		return nil, err
	}

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
		cmd.AddCommand(subcommand.Initialize(c.SubCommandInitializeFunc))
	}

	configureVersionFlag(cmd)                             // Configure app for version information
	configureOutputFlags(cmd)                             // Configure verbosity
	configureLoggingFlags(cmd)                            // Configure logging
	cmd.PersistentFlags().SetNormalizeFunc(normalizeFunc) // normalize persistent flags

	return cmd, nil
}

func (c Config) RegisterCommand(cmd Commander, f func(*cobra.Command)) {
	appCmd.AddCommand(cmd.Initialize(f))
}

func (c Config) RegisterCommands(cmds []Commander, f func(*cobra.Command)) {
	for _, cmd := range cmds {
		appCmd.AddCommand(cmd.Initialize(f))
	}
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
