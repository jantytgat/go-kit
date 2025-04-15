package application

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"git.flexabyte.io/flexabyte/go-kit/pkg/semver"
)

type Config struct {
	Name                     string
	Title                    string
	Banner                   string
	Version                  string
	EnableGracefulShutdown   bool
	OverrideRunE             func(cmd *cobra.Command, args []string) error
	PersistentPreRunE        []func(cmd *cobra.Command, args []string) error // collection of PreRunE functions
	PersistentPostRunE       []func(cmd *cobra.Command, args []string) error // collection of PostRunE functions
	ShutdownSignals          []os.Signal
	ShutdownTimeout          time.Duration
	SubCommands              []Command
	SubCommandInitializeFunc func(cmd *cobra.Command)
	ValidArgs                []string
}

func (c Config) getRootCommand() (*cobra.Command, error) {
	var err error
	if err = c.Validate(); err != nil {
		return nil, err
	}

	var long string
	if c.Banner != "" {
		long = c.Banner + "\n" + c.Title
	} else {
		long = c.Title
	}

	cmd := &cobra.Command{
		Use:                c.Name,
		Short:              c.Title,
		Long:               long,
		PersistentPreRunE:  persistentPreRunFuncE,
		PersistentPostRunE: persistentPostRunFuncE,
		RunE:               runFuncE,
		SilenceErrors:      true,
		SilenceUsage:       true,
	}

	if c.OverrideRunE != nil {
		cmd.RunE = c.OverrideRunE
	}

	for _, subcommand := range c.SubCommands {
		cmd.AddCommand(subcommand.Initialize(c.SubCommandInitializeFunc))
	}

	var v semver.Version
	if v, err = c.ParseVersion(); err != nil {
		return nil, err
	}

	configureVersionFlag(cmd, v)                          // Configure app for version information
	configureOutputFlags(cmd)                             // Configure verbosity
	configureLoggingFlags(cmd)                            // Configure logging
	cmd.PersistentFlags().SetNormalizeFunc(normalizeFunc) // normalize persistent flags

	return cmd, nil
}

func (c Config) ParseVersion() (semver.Version, error) {
	return semver.Parse(c.Version)
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

	var err error
	if _, err = semver.Parse(c.Version); err != nil {
		return fmt.Errorf("invalid version: %s", c.Version)
	}
	return nil
}
