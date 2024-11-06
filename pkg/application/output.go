package application

import "github.com/spf13/cobra"

var jsonOutputFlag bool
var noColorFlag bool
var quietFlag bool
var verboseFlag bool

func addJsonOutputFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&jsonOutputFlag, "json", "", false, "Enable JSON output")
}

func addNoColorFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&noColorFlag, "no-color", "", false, "Disable color output")
}

func addQuietFlag(c *cobra.Command) {
	c.PersistentFlags().BoolVarP(&quietFlag, "quiet", "q", false, "Enable quiet mode")
}

func addVerboseFlag(c *cobra.Command) {
	c.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "Enable verbose output")
}

func configureVerbosity() {
	addJsonOutputFlag(app)
	addNoColorFlag(app)
	addVerboseFlag(app)
	addQuietFlag(app)

	app.MarkFlagsMutuallyExclusive("verbose", "quiet", "json")
	app.MarkFlagsMutuallyExclusive("json", "no-color")
	app.MarkFlagsMutuallyExclusive("quiet", "no-color")

}
