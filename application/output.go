package application

import "github.com/spf13/cobra"

var jsonOutputFlag bool
var noColorFlag bool
var quietFlag bool
var verboseFlag bool

func addJsonOutputFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&jsonOutputFlag, "json", "", false, "Enable JSON outWriter")
}

func addNoColorFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&noColorFlag, "no-color", "", false, "Disable color outWriter")
}

func addQuietFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&quietFlag, "quiet", "q", false, "Enable quiet mode")
}

func addVerboseFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "Enable verbose outWriter")
}

func configureOutputFlags(cmd *cobra.Command) {
	addJsonOutputFlag(cmd)
	addNoColorFlag(cmd)
	addVerboseFlag(cmd)
	addQuietFlag(cmd)

	cmd.MarkFlagsMutuallyExclusive("verbose", "quiet", "json")
	cmd.MarkFlagsMutuallyExclusive("json", "no-color")
	cmd.MarkFlagsMutuallyExclusive("quiet", "no-color")

}
