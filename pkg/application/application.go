package application

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/jantytgat/go-kit/pkg/semver"
	"github.com/jantytgat/go-kit/pkg/slogd"
)

const (
	shutdownTimeOut       = 30 * time.Second
	shutdownSleepInterval = 1 * time.Second
)

var appName string
var app = &cobra.Command{
	PersistentPreRunE:  persistentPreRunFuncE,
	PersistentPostRunE: persistentPostRunFuncE,
	RunE:               appRunE,
	SilenceUsage:       true,
	SilenceErrors:      true,
}

var persistentPreRunE []func(cmd *cobra.Command, args []string) error
var persistentPostRunE []func(cmd *cobra.Command, args []string) error

// var logger *slog.Logger
var out io.Writer = os.Stdout
var muxOut = &sync.Mutex{}

func New(name, title, banner string, v semver.Version) {
	var err error
	if err = configureApp(name, title, banner); err != nil {
		panic(err)
	}

	// Configure app for version information
	configureVersion(v)

	// Configure verbosity
	configureVerbosity()

	// Configure logging
	configureLogging()
	app.PersistentFlags().SetNormalizeFunc(normalizeFunc)
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
	runCtx, runCancel := context.WithCancel(context.WithValue(ctx, "run", "application run"))
	defer runCancel()

	// Signal Handling
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Result channel from command execution
	chExe := make(chan error)

	exeCtx, exeCancel := context.WithCancel(context.WithValue(runCtx, "exe", "application exe"))
	defer exeCancel()

	// Goroutine to handle graceful shutdown
	go func(ctx context.Context, cancel context.CancelFunc, sig chan os.Signal, chExe chan error) {
		<-sig
		cancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(ctx, shutdownTimeOut)
		defer shutdownCancel()

		slogd.FromContext(ctx).LogAttrs(ctx, slogd.LevelDebug.Level(), "cleaning up for shutdown")
		for {
			select {
			case <-shutdownCtx.Done():
				if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
					slogd.FromContext(ctx).LogAttrs(ctx, slogd.LevelError.Level(), "timeout while waiting for cleanup to complete")
				}
				chExe <- shutdownCtx.Err()
				return
			default:
				slogd.FromContext(ctx).LogAttrs(ctx, slogd.LevelTrace.Level(), "waiting for cleanup to complete")
				time.Sleep(shutdownSleepInterval)
			}
		}
	}(runCtx, exeCancel, sig, chExe)

	// Execute command
	go func(ctx context.Context, chExe chan error) {
		chExe <- app.ExecuteContext(ctx)
	}(exeCtx, chExe)

	// Wait for result of command execution or graceful shutdown
	return <-chExe
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

func normalizeFunc(f *pflag.FlagSet, name string) pflag.NormalizedName {
	// switch name {
	// case "no-color":
	// 	name = "log-type"
	// 	break
	// }
	return pflag.NormalizedName(name)
}

func persistentPreRunFuncE(cmd *cobra.Command, args []string) error {
	slogd.SetLevel(slogd.Level(logLevelFlag).Level())
	if slogd.ActiveHandler() == slogd.HandlerColor && noColorFlag {
		slogd.UseHandler(slogd.HandlerText)
		cmd.SetContext(slogd.WithContext(cmd.Context()))
	}
	slogd.FromContext(cmd.Context()).Log(cmd.Context(), slogd.LevelTrace.Level(), "starting application")
	slogd.FromContext(cmd.Context()).Log(cmd.Context(), slogd.LevelTrace.Level(), "executing PersistentPreRun")

	// Make sure we can always get the version
	if versionFlag || cmd.Use == versionName {
		slogd.FromContext(cmd.Context()).LogAttrs(cmd.Context(), slogd.LevelTrace.Level(), "overriding command", slog.String("old_function", runtime.FuncForPC(reflect.ValueOf(cmd.RunE).Pointer()).Name()), slog.String("new_function", runtime.FuncForPC(reflect.ValueOf(versionRunE).Pointer()).Name()))
		cmd.RunE = versionRunE
		return nil
	}

	// Make sure that we show the app help if no commands or flags are passed
	if cmd.CalledAs() == appName {
		slogd.FromContext(cmd.Context()).LogAttrs(cmd.Context(), slogd.LevelTrace.Level(), "overriding command", slog.String("old_function", runtime.FuncForPC(reflect.ValueOf(cmd.RunE).Pointer()).Name()), slog.String("new_function", runtime.FuncForPC(reflect.ValueOf(helpE).Pointer()).Name()))

		cmd.RunE = helpE
		return nil
	}

	if quietFlag {
		slogd.FromContext(cmd.Context()).LogAttrs(cmd.Context(), slogd.LevelDebug.Level(), "activating quiet mode")
		out = io.Discard
	}

	if persistentPreRunE == nil {
		return nil
	}

	var err error
	for _, preRun := range persistentPreRunE {
		slogd.FromContext(cmd.Context()).Log(cmd.Context(), slogd.LevelTrace.Level(), "executing PersistentPreRun function", slog.String("function", runtime.FuncForPC(reflect.ValueOf(preRun).Pointer()).Name()))
		if err = preRun(cmd, args); err != nil {
			return err
		}
	}
	return nil
}

func persistentPostRunFuncE(cmd *cobra.Command, args []string) error {
	defer slogd.FromContext(cmd.Context()).Log(cmd.Context(), slogd.LevelTrace.Level(), "stopping application")
	slogd.FromContext(cmd.Context()).Log(cmd.Context(), slogd.LevelTrace.Level(), "executing PersistentPostRunE")

	if persistentPostRunE == nil {
		return nil
	}

	var err error
	for _, preRun := range persistentPostRunE {
		slogd.FromContext(cmd.Context()).Log(cmd.Context(), slogd.LevelTrace.Level(), "executing PersistentPostRun function", slog.String("function", runtime.FuncForPC(reflect.ValueOf(preRun).Pointer()).Name()))
		if err = preRun(cmd, args); err != nil {
			return err
		}
	}
	return nil
}
