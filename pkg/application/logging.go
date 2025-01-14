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

func addLogLevelFlag(c *cobra.Command) {
	c.PersistentFlags().StringVarP(&logLevelFlag, "log-level", "", "info", "Set log level (trace, debug, info, warn, error, fatal)")
}

func addLogOutputFlag(c *cobra.Command) {
	c.PersistentFlags().StringVarP(&logOutputFlag, "log-output", "", "stderr", "Set log output (stdout, stderr, file)")
}

func addLogTypeFlag(c *cobra.Command) {
	c.PersistentFlags().StringVarP(&logTypeFlag, "log-type", "", "text", "Set log type (text, json, color)")
}

func configureLogging() {
	addLogLevelFlag(app)
	addLogOutputFlag(app)
	addLogTypeFlag(app)

	app.MarkFlagsMutuallyExclusive("no-color", "log-type")
}
