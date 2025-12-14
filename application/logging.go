package application

import (
	"fmt"
	"log/slog"

	"github.com/Oudwins/zog"
	"github.com/spf13/cobra"

	"github.com/jantytgat/go-kit/flagzog"
	"github.com/jantytgat/go-kit/slogd"
)

const (
	LogLevelTrace   LogLevel  = "trace"
	LogLevelDebug   LogLevel  = "debug"
	LogLevelInfo    LogLevel  = "info"
	LogLevelWarn    LogLevel  = "warn"
	LogLevelError   LogLevel  = "error"
	LogLevelFatal   LogLevel  = "fatal"
	LogOutputStdout LogOutput = "stdout"
	LogOutputStderr LogOutput = "stderr"
	LogOutputFile   LogOutput = "file"
	LogFormatText   LogFormat = "text"
	LogFormatJson   LogFormat = "json"
	LogFormatColor  LogFormat = "color"
)

var (
	logLevelFlag = flagzog.NewStringFlag(
		"log-level",
		zog.String().OneOf([]string{
			string(LogLevelTrace),
			string(LogLevelDebug),
			string(LogLevelInfo),
			string(LogLevelWarn),
			string(LogLevelError),
			string(LogLevelFatal)}),
		fmt.Sprintf("Set log level (%s, %s, %s, %s, %s, %s)", LogLevelTrace, LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError, LogLevelFatal))
	logOutputFlag = flagzog.NewStringFlag(
		"log-output",
		zog.String().OneOf([]string{
			string(LogOutputStdout),
			string(LogOutputStderr),
			string(LogOutputFile)}),
		fmt.Sprintf("Set log output (%s, %s, %s)", LogOutputStdout, LogOutputStderr, LogOutputFile))
	logDestinationFlag = flagzog.NewStringFlag(
		"log-destination",
		zog.String(),
		fmt.Sprintf("Set log destination (%s, %s, %s)", LogOutputStderr, LogOutputStdout, "filepath"))
	logFormatFlag = flagzog.NewStringFlag(
		"log-type",
		zog.String().OneOf([]string{
			string(LogFormatText),
			string(LogFormatJson),
			string(LogFormatColor)}),
		fmt.Sprintf("Set log type (%s, %s, %s)", LogFormatText, LogFormatJson, LogFormatColor))
)

type LogOutput string
type LogDestination string
type LogFormat string
type LogLevel string

func addLogLevelFlag(cmd *cobra.Command, l LogLevel) {
	cmd.PersistentFlags().StringVarP(&logLevelFlag.Value, logLevelFlag.Name(), "", string(l), logLevelFlag.Usage())
}

func addLogOutputFlag(cmd *cobra.Command, o LogOutput) {
	cmd.PersistentFlags().StringVarP(&logOutputFlag.Value, logOutputFlag.Name(), "", string(o), logOutputFlag.Usage())
}

func addLogDestinationFlag(cmd *cobra.Command, d LogDestination) {
	cmd.PersistentFlags().StringVarP(&logDestinationFlag.Value, logDestinationFlag.Name(), "", string(d), logDestinationFlag.Usage())
}

func addLogFormatFlag(cmd *cobra.Command, t LogFormat) {
	cmd.PersistentFlags().StringVarP(&logFormatFlag.Value, logFormatFlag.Name(), "", string(t), logFormatFlag.Usage())
}

func GetLogLevelFromArgs(args []string, defaultLevel slog.Level) slog.Level {
	for i, arg := range args {
		if arg == "--log-level" && i+1 < len(args) {
			return slogd.GetLevelFromString(args[i+1])
		}
	}
	return defaultLevel
}
