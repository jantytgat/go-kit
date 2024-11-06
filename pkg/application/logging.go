package application

import "github.com/spf13/cobra"

var logLevelFlag string
var logTargetFlag string

func addLogLevelFlag(c *cobra.Command) {
	c.PersistentFlags().StringVarP(&logLevelFlag, "log-level", "", "", "Set log level (trace, debug, info, warn, error, fatal)")
}

func addLogOutputFlag(c *cobra.Command) {
	c.PersistentFlags().StringVarP(&logTargetFlag, "log-output", "", "stderr", "Set log output (stdout, stderr, filename)")
}

func configureLogging() {
	addLogLevelFlag(app)
	addLogOutputFlag(app)
}
