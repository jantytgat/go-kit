package application

import (
	"github.com/Oudwins/zog"
	"github.com/spf13/cobra"

	"git.flexabyte.io/flexabyte/go-kit/flagzog"
)

const (
	quietFlagShortCode   = "q"
	verboseFlagShortCode = "v"
)

var (
	jsonOutputFlag = flagzog.NewBoolFlag("json", zog.Bool(), "Enable JSON output")
	noColorFlag    = flagzog.NewBoolFlag("no-color", zog.Bool(), "Disable colored output")
	quietFlag      = flagzog.NewBoolFlag("quiet", zog.Bool(), "Suppress output")
	verboseFlag    = flagzog.NewBoolFlag("verbose", zog.Bool(), "Enable verbose output")
)

func addJsonOutputFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&jsonOutputFlag.Value, jsonOutputFlag.Name(), "", false, jsonOutputFlag.Usage())
}

func addNoColorFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&noColorFlag.Value, noColorFlag.Name(), "", false, noColorFlag.Usage())
}

func addQuietFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&quietFlag.Value, quietFlag.Name(), quietFlagShortCode, false, quietFlag.Usage())
}

func addVerboseFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&verboseFlag.Value, verboseFlag.Name(), verboseFlagShortCode, false, verboseFlag.Usage())
}

func configureOutputFlags(cmd *cobra.Command) {
	addJsonOutputFlag(cmd)
	addNoColorFlag(cmd)
	addVerboseFlag(cmd)
	addQuietFlag(cmd)

	cmd.MarkFlagsMutuallyExclusive(verboseFlag.Name(), quietFlag.Name(), jsonOutputFlag.Name())
	cmd.MarkFlagsMutuallyExclusive(jsonOutputFlag.Name(), noColorFlag.Name())
	cmd.MarkFlagsMutuallyExclusive(quietFlag.Name(), noColorFlag.Name())
}
