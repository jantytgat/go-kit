package application

import (
	"github.com/spf13/cobra"
)

const (
	LogOutputStdOut = "stdout"
	LogOutputStdErr = "stderr"
	LogOutputFile   = "file"
)

var logLevelFlag string
var logOutputFlag string
var logTypeFlag string

func addLogLevelFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&logLevelFlag, "log-level", "", "info", "Set log level (trace, debug, info, warn, error, fatal)")
}

func addLogOutputFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&logOutputFlag, "log-outWriter", "", "stderr", "Set log outWriter (stdout, stderr, file)")
}

func addLogTypeFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&logTypeFlag, "log-type", "", "text", "Set log type (text, json, color)")
}

func configureLoggingFlags(cmd *cobra.Command) {
	addLogLevelFlag(cmd)
	addLogOutputFlag(cmd)
	addLogTypeFlag(cmd)

	cmd.MarkFlagsMutuallyExclusive("no-color", "log-type")
}
