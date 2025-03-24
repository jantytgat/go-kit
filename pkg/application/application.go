package application

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/jantytgat/go-kit/pkg/semver"
	"github.com/jantytgat/go-kit/pkg/slogd"
	"github.com/jantytgat/go-kit/pkg/slogd-colored"
)

const (
	shutdownMaxSeconds        = 5
	shutdownTimeOut           = shutdownMaxSeconds * time.Second
	shutdownCountdownInterval = 1 * time.Second
)

var (
	persistentPreRunE  []func(cmd *cobra.Command, args []string) error // collection of PreRunE functions
	persistentPostRunE []func(cmd *cobra.Command, args []string) error // collection of PostRunE functions
	appName            string                                          // name of the application
	app                = &cobra.Command{
		PersistentPreRunE:  persistentPreRunFuncE,
		PersistentPostRunE: persistentPostRunFuncE,
		RunE:               appRunE,
		SilenceUsage:       true,
		SilenceErrors:      true,
	}
)

// var logger *slog.Logger
var out io.Writer = os.Stdout

func RegisterFlag(f func(*cobra.Command)) {
	f(app)
}

func EnableTraverseRunHooks() {
	cobra.EnableTraverseRunHooks = true
}

func New(name, title, banner string, v semver.Version) {
	var err error
	if err = configureApp(name, title, banner); err != nil {
		panic(err)
	}

	configureVersion(v)                                   // Configure app for version information
	configureVerbosity()                                  // Configure verbosity
	configureLogging()                                    // Configure logging
	app.PersistentFlags().SetNormalizeFunc(normalizeFunc) // normalize persistent flags
}

func RegisterCommand(cmd Commander, f func(*cobra.Command)) {
	app.AddCommand(cmd.Initialize(f))
}

func RegisterCommands(cmds []Commander, f func(*cobra.Command)) {
	for _, cmd := range cmds {
		app.AddCommand(cmd.Initialize(f))
	}
}

func RegisterPersistentPreRunE(f func(cmd *cobra.Command, args []string) error) {
	persistentPreRunE = append(persistentPreRunE, f)
}

func RegisterPersistentPostRunE(f func(cmd *cobra.Command, args []string) error) {
	persistentPostRunE = append(persistentPostRunE, f)
}

func RegisterValidArgs(args []string) {
	app.ValidArgs = args
}

func Run(ctx context.Context) error {
	// Result channel from command execution
	chErr := make(chan error)

	exeCtx, exeCancel := context.WithCancel(ctx)
	defer exeCancel()

	go gracefulShutdown(ctx, exeCancel, chErr) // Goroutine to handle graceful shutdown
	go executeCommand(exeCtx, chErr)           // Execute command
	return <-chErr                             // Wait for result of command execution or graceful shutdown
}

// appRunE is an empty catch function to allow overrides through persistentPreRunE
func appRunE(cmd *cobra.Command, args []string) error {
	return nil
}

func configureApp(name, title, banner string) error {
	if name == "" {
		return fmt.Errorf("application name is empty")
	}
	appName = name // Configure app name
	app.Use = name // Configure root command name

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

func executeCommand(ctx context.Context, chErr chan error) {
	if err := app.ExecuteContext(ctx); err != nil {
		chErr <- err
	}
	chErr <- nil
}

func gracefulShutdown(ctx context.Context, cancel context.CancelFunc, chErr chan error) {
	// Signal Handling
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Wait for signal
	s := <-sig
	slogd.Logger().LogAttrs(ctx, slogd.LevelWarn, "received shutdown signal", slog.String("signal", s.String()))

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, shutdownTimeOut)
	defer shutdownCancel()

	go gracefulShutdownCounter(shutdownCtx)
	cancel()
	fmt.Println("Press Ctrl+C again to force shutdown")
	select {
	case <-shutdownCtx.Done():
		slogd.Logger().LogAttrs(ctx, slogd.LevelError, "graceful shutdown failed")
		fmt.Println("graceful shutdown failed... exiting")
		chErr <- fmt.Errorf("graceful shutdown failed")
		return
	case s2 := <-sig:
		slogd.Logger().LogAttrs(ctx, slogd.LevelWarn, "received forced shutdown signal", slog.String("signal", s.String()))
		chErr <- fmt.Errorf("received signal %q", s2)
		return
	}
}

func gracefulShutdownCounter(ctx context.Context) {
	slogd.Logger().LogAttrs(ctx, slogd.LevelDebug, "starting graceful shutdown", slog.String("limit", fmt.Sprintf("%ds", shutdownMaxSeconds)))
	counter := 0
	for {
		counter++
		slogd.Logger().LogAttrs(ctx, slogd.LevelTrace, "waiting for graceful shutdown to complete", slog.String("limit", fmt.Sprintf("%ds", shutdownMaxSeconds-counter)))
		time.Sleep(shutdownCountdownInterval)
	}
}

func helpE(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}

func normalizeFunc(f *pflag.FlagSet, name string) pflag.NormalizedName {
	// switch name {
	// case "no-color":
	// 	name = "log-type"
	// 	break
	// }
	return pflag.NormalizedName(name)
}

func persistentPreRunFuncE(cmd *cobra.Command, args []string) error {
	slogd.SetLevel(slogd.Level(logLevelFlag))
	if slogd.ActiveHandler() == slogd_colored.HandlerColor && noColorFlag {
		slogd.UseHandler(slogd.HandlerText)
		cmd.SetContext(slogd.WithContext(cmd.Context()))
	}

	slogd.FromContext(cmd.Context()).Log(cmd.Context(), slogd.LevelTrace, "starting application", slog.String("command", cmd.CommandPath()))
	slogd.FromContext(cmd.Context()).Log(cmd.Context(), slogd.LevelTrace, "executing PersistentPreRun")

	// Make sure we can always get the version
	if versionFlag || cmd.Use == versionName {
		slogd.FromContext(cmd.Context()).LogAttrs(cmd.Context(), slogd.LevelTrace, "overriding command", slog.String("old_function", runtime.FuncForPC(reflect.ValueOf(cmd.RunE).Pointer()).Name()), slog.String("new_function", runtime.FuncForPC(reflect.ValueOf(versionRunE).Pointer()).Name()))
		cmd.RunE = versionRunE
		return nil
	}

	// Make sure that we show the app help if no commands or flags are passed
	if cmd.CalledAs() == appName {
		slogd.FromContext(cmd.Context()).LogAttrs(cmd.Context(), slogd.LevelTrace, "overriding command", slog.String("old_function", runtime.FuncForPC(reflect.ValueOf(cmd.RunE).Pointer()).Name()), slog.String("new_function", runtime.FuncForPC(reflect.ValueOf(helpE).Pointer()).Name()))

		cmd.RunE = helpE
		return nil
	}

	if quietFlag {
		slogd.FromContext(cmd.Context()).LogAttrs(cmd.Context(), slogd.LevelDebug, "activating quiet mode")
		out = io.Discard
	}

	if persistentPreRunE == nil {
		return nil
	}

	var err error
	for _, preRun := range persistentPreRunE {
		slogd.FromContext(cmd.Context()).Log(cmd.Context(), slogd.LevelTrace, "executing PersistentPreRun function", slog.String("function", runtime.FuncForPC(reflect.ValueOf(preRun).Pointer()).Name()))
		if err = preRun(cmd, args); err != nil {
			return err
		}
	}
	return nil
}

func persistentPostRunFuncE(cmd *cobra.Command, args []string) error {
	defer slogd.FromContext(cmd.Context()).Log(cmd.Context(), slogd.LevelTrace, "stopping application", slog.String("command", cmd.CommandPath()))
	slogd.FromContext(cmd.Context()).Log(cmd.Context(), slogd.LevelTrace, "executing PersistentPostRunE")

	if persistentPostRunE == nil {
		return nil
	}

	var err error
	for _, postRun := range persistentPostRunE {
		slogd.FromContext(cmd.Context()).Log(cmd.Context(), slogd.LevelTrace, "executing PersistentPostRun function", slog.String("function", runtime.FuncForPC(reflect.ValueOf(postRun).Pointer()).Name()))
		if err = postRun(cmd, args); err != nil {
			return err
		}
	}
	return nil
}
