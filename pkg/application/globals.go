package application

import (
	"io"
	"log/slog"
	"os"
	"reflect"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"git.flexabyte.io/flexabyte/go-slogd/pkg/slogd"

	"git.flexabyte.io/flexabyte/go-kit/pkg/slogd_colored"
)

var (
	appName            string
	appCmd             *cobra.Command
	persistentPreRunE  []func(cmd *cobra.Command, args []string) error // collection of PreRunE functions
	persistentPostRunE []func(cmd *cobra.Command, args []string) error // collection of PostRunE functions
	outWriter          io.Writer                                       = os.Stdout
	// version            semver.Version
)

func helpFuncE(cmd *cobra.Command, args []string) error {
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
		slogd.FromContext(cmd.Context()).LogAttrs(cmd.Context(), slogd.LevelTrace, "overriding command", slog.String("old_function", runtime.FuncForPC(reflect.ValueOf(cmd.RunE).Pointer()).Name()), slog.String("new_function", runtime.FuncForPC(reflect.ValueOf(versionRunFuncE).Pointer()).Name()))
		cmd.RunE = versionRunFuncE
		return nil
	}

	// Make sure that we show the app help if no commands or flags are passed
	if cmd.CalledAs() == appName && runtime.FuncForPC(reflect.ValueOf(cmd.RunE).Pointer()).Name() == runtime.FuncForPC(reflect.ValueOf(runFuncE).Pointer()).Name() {
		slogd.FromContext(cmd.Context()).LogAttrs(cmd.Context(), slogd.LevelTrace, "overriding command", slog.String("old_function", runtime.FuncForPC(reflect.ValueOf(cmd.RunE).Pointer()).Name()), slog.String("new_function", runtime.FuncForPC(reflect.ValueOf(helpFuncE).Pointer()).Name()))

		cmd.RunE = helpFuncE
		return nil
	}

	// TODO move to front??
	if quietFlag {
		slogd.FromContext(cmd.Context()).LogAttrs(cmd.Context(), slogd.LevelDebug, "activating quiet mode")
		outWriter = io.Discard
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

// appRunE is an empty catch function to allow overrides through persistentPreRunE
func runFuncE(cmd *cobra.Command, args []string) error {
	return nil
}
