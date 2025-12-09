package application

import (
	"github.com/Oudwins/zog"
	"github.com/spf13/cobra"

	"github.com/jantytgat/go-kit/flagzog"
)

const (
	jsonOutputFlagDefault = false
	noColorFlagDefault    = false
	quietFlagDefault      = false
	quietFlagShortCode    = "q"
	verboseFlagDefault    = false
	verboseFlagShortCode  = "v"
)

var (
	jsonOutputFlag = flagzog.NewBoolFlag("json", zog.Bool(), "Set output to JSON")
	noColorFlag    = flagzog.NewBoolFlag("no-color", zog.Bool(), "Disable colored output")
	quietFlag      = flagzog.NewBoolFlag("quiet", zog.Bool(), "Suppress output")
	verboseFlag    = flagzog.NewBoolFlag("verbose", zog.Bool(), "Enable verbose output")
)

func addJsonOutputFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&jsonOutputFlag.Value, jsonOutputFlag.Name(), "", jsonOutputFlagDefault, jsonOutputFlag.Usage())
}

func addNoColorFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&noColorFlag.Value, noColorFlag.Name(), "", noColorFlagDefault, noColorFlag.Usage())
}

func addQuietFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&quietFlag.Value, quietFlag.Name(), quietFlagShortCode, quietFlagDefault, quietFlag.Usage())
}

func addVerboseFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&verboseFlag.Value, verboseFlag.Name(), verboseFlagShortCode, verboseFlagDefault, verboseFlag.Usage())
}
