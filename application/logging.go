package application

import (
	"fmt"
	"log/slog"

	"github.com/Oudwins/zog"
	"github.com/spf13/cobra"

	"git.flexabyte.io/flexabyte/go-kit/flagzog"
	"git.flexabyte.io/flexabyte/go-kit/slogd"
)

const (
	logLevelTrace   = "trace"
	logLevelDebug   = "debug"
	logLevelInfo    = "info"
	logLevelWarn    = "warn"
	logLevelError   = "error"
	logLevelFatal   = "fatal"
	logOutputStdout = "stdout"
	logOutputStderr = "stderr"
	logOutputFile   = "file"
	logTypeText     = "text"
	logTypeJson     = "json"
	logTypeColor    = "color"
)

var (
	logLevelFlag  = flagzog.NewStringFlag("log-level", zog.String().OneOf([]string{logLevelTrace, logLevelDebug, logLevelInfo, logLevelWarn, logLevelError, logLevelFatal}), fmt.Sprintf("Set log level (%s, %s, %s, %s, %s, %s)", logLevelTrace, logLevelDebug, logLevelInfo, logLevelWarn, logLevelError, logLevelFatal))
	logOutputFlag = flagzog.NewStringFlag("log-output", zog.String().OneOf([]string{logOutputStdout, logOutputStderr, logOutputFile}), fmt.Sprintf("Set log output (%s, %s, %s)", logOutputStdout, logOutputStderr, logOutputFile))
	logTypeFlag   = flagzog.NewStringFlag("log-type", zog.String().OneOf([]string{logTypeText, logTypeJson, logTypeColor}), fmt.Sprintf("Set log type (%s, %s, %s)", logTypeText, logTypeJson, logTypeColor))
)

func addLogLevelFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&logLevelFlag.Value, logLevelFlag.Name(), "", logLevelInfo, logLevelFlag.Usage())
}

func addLogOutputFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&logOutputFlag.Value, logOutputFlag.Name(), "", logOutputStderr, logOutputFlag.Usage())
}

func addLogTypeFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&logTypeFlag.Value, logTypeFlag.Name(), "", logTypeText, logTypeFlag.Usage())
}

func configureLoggingFlags(cmd *cobra.Command) {
	addLogLevelFlag(cmd)
	addLogOutputFlag(cmd)
	addLogTypeFlag(cmd)

	cmd.MarkFlagsMutuallyExclusive("no-color", logTypeFlag.Name())
}

func GetLogLevelFromArgs(args []string) slog.Level {
	for i, arg := range args {
		if arg == "--log-level" && i+1 < len(args) {
			return slogd.Level(args[i+1])
		}
	}
	return slogd.LevelDefault
}
