package application

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"reflect"
	"runtime"
	"sync"

	"github.com/spf13/cobra"

	"github.com/jantytgat/go-kit/pkg/pslog"
	"github.com/jantytgat/go-kit/pkg/semver"
)

var appName string
var app = &cobra.Command{
	PersistentPreRunE:  persistentPreRunFuncE,
	PersistentPostRunE: persistentPostRunFuncE,
	RunE:               appRunE,
}

var persistentPreRunE []func(cmd *cobra.Command, args []string) error
var persistentPostRunE []func(cmd *cobra.Command, args []string) error

var logger *slog.Logger
var out io.Writer = os.Stdout
var muxOut = &sync.Mutex{}

func New(name, title, banner string, v semver.Version, l *slog.Logger) error {
	var err error
	if err = configureApp(name, title, banner); err != nil {
		return err
	}

	// Set logger
	if l != nil {
		logger = l
	}

	// Configure app for version information
	configureVersion(v)

	// Configure verbosity
	configureVerbosity()

	// Configure logging
	configureLogging()

	return nil
}

func Output() (io.Writer, *sync.Mutex) {
	return out, muxOut
}

func RegisterCommand(c *cobra.Command) {
	app.AddCommand(c)
}

func RegisterPreRunE(f func(cmd *cobra.Command, args []string) error) {
	persistentPreRunE = append(persistentPreRunE, f)
}

func Run(ctx context.Context) error {
	return app.ExecuteContext(ctx)
}

// appRunE is an empty catch function to allow overrides through persistentPreRunE
func appRunE(cmd *cobra.Command, args []string) error {
	return nil
}

func configureApp(name, title, banner string) error {
	if name == "" {
		return fmt.Errorf("application name is empty")
	}
	appName = name

	// Configure app
	app.Use = name

	if title == "" {
		return fmt.Errorf("application title is empty")
	}
	app.Short = title

	if banner != "" {
		app.Long = banner + "\n" + title
	} else {
		app.Long = title
	}
	return nil
}

func helpE(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}

func persistentPreRunFuncE(cmd *cobra.Command, args []string) error {
	logger.Log(cmd.Context(), pslog.LevelTrace.Level(), "executing PersistentPreRun")

	// Make sure we can always get the version
	if versionFlag || cmd.Use == versionName {
		logger.LogAttrs(cmd.Context(), pslog.LevelTrace.Level(), "overriding command", slog.String("old_function", runtime.FuncForPC(reflect.ValueOf(cmd.RunE).Pointer()).Name()), slog.String("new_function", runtime.FuncForPC(reflect.ValueOf(versionRunE).Pointer()).Name()))
		cmd.RunE = versionRunE
		return nil
	}

	// Make sure that we show the app help if no commands or flags are passed
	if cmd.CalledAs() == appName {
		logger.LogAttrs(cmd.Context(), pslog.LevelTrace.Level(), "overriding command", slog.String("old_function", runtime.FuncForPC(reflect.ValueOf(cmd.RunE).Pointer()).Name()), slog.String("new_function", runtime.FuncForPC(reflect.ValueOf(helpE).Pointer()).Name()))
		cmd.RunE = helpE
		return nil
	}

	if quietFlag {
		logger.LogAttrs(cmd.Context(), pslog.LevelDebug.Level(), "activating quiet mode")
		out = io.Discard
	}

	if persistentPreRunE == nil {
		return nil
	}

	var err error
	for _, preRun := range persistentPreRunE {
		logger.Log(cmd.Context(), pslog.LevelTrace.Level(), "executing PersistentPreRun function", slog.String("function", runtime.FuncForPC(reflect.ValueOf(preRun).Pointer()).Name()))
		if err = preRun(cmd, args); err != nil {
			return err
		}
	}
	return nil
}

func persistentPostRunFuncE(cmd *cobra.Command, args []string) error {
	logger.Log(cmd.Context(), pslog.LevelTrace.Level(), "executing PersistentPostRunE")

	if persistentPostRunE == nil {
		return nil
	}

	var err error
	for _, preRun := range persistentPostRunE {
		logger.Log(cmd.Context(), pslog.LevelTrace.Level(), "executing PersistentPostRun function", slog.String("function", runtime.FuncForPC(reflect.ValueOf(preRun).Pointer()).Name()))
		if err = preRun(cmd, args); err != nil {
			return err
		}
	}
	return nil
}
