package application

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"git.flexabyte.io/flexabyte/go-kit/semver"
)

const (
	versionName      = "version"
	versionShortHand = "V"
	versionUsage     = "Show version information"
)

var (
	version     semver.Version
	versionFlag bool
	versionCmd  = &cobra.Command{
		Use:   versionName,
		Short: versionUsage,
		RunE:  versionRunFuncE,
	}
)

func addVersionFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&versionFlag, versionName, versionShortHand, false, versionUsage)
}

func configureVersionFlag(cmd *cobra.Command, v semver.Version) {
	version = v
	cmd.AddCommand(versionCmd)
	addVersionFlag(cmd)
}

func printVersion(v semver.Version) string {
	var output string
	if !verboseFlag {
		output = v.String()
	}

	if jsonOutputFlag {
		var b []byte
		b, _ = json.Marshal(v)
		output = string(b)
	}

	if output != "" {
		return output
	}

	return fmt.Sprintf(
		"Full: %s\nVersion: %s\nChannel: %s\nCommit: %s\nDate: %s",
		v.String(),
		v.Number(),
		v.Release(),
		v.Commit(),
		v.Date(),
	)
}

func versionRunFuncE(cmd *cobra.Command, args []string) error {
	if _, err := fmt.Fprintln(outWriter, printVersion(version)); err != nil {
		return err
	}
	return nil
}
