package application

import (
	"log/slog"

	"git.flexabyte.io/flexabyte/go-slogd/slogd"
	"github.com/spf13/cobra"
)

const (
	LogOutputStdOut = "stdout"
	LogOutputStdErr = "stderr"
	LogOutputFile   = "file"
)

var (
	logLevelFlagName  string = "log-level"
	logLevelFlag      string
	logOutputFlagName string = "log-output"
	logOutputFlag     string
	logTypeFlagName   = "log-type"
	logTypeFlag       string
)

func addLogLevelFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&logLevelFlag, logLevelFlagName, "", "info", "Set log level (trace, debug, info, warn, error, fatal)")
}

func addLogOutputFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&logOutputFlag, logOutputFlagName, "", "stderr", "Set log output (stdout, stderr, file)")
}

func addLogTypeFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&logTypeFlag, logTypeFlagName, "", "text", "Set log type (text, json, color)")
}

func configureLoggingFlags(cmd *cobra.Command) {
	addLogLevelFlag(cmd)
	addLogOutputFlag(cmd)
	addLogTypeFlag(cmd)

	cmd.MarkFlagsMutuallyExclusive("no-color", "log-type")
}

func GetLogLevelFromArgs(args []string) slog.Level {
	for i, arg := range args {
		if arg == "--log-level" && i+1 < len(args) {
			return slogd.Level(args[i+1])
		}
	}
	return slogd.LevelDefault
}
