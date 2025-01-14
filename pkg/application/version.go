package application

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jantytgat/go-kit/pkg/semver"
)

const (
	versionName      = "version"
	versionShortHand = "V"
	versionUsage     = "Show version information"
)

var version semver.Version
var versionFlag bool
var versionCmd = &cobra.Command{
	Use:   versionName,
	Short: versionUsage,
	RunE:  versionRunE,
}

func addVersionFlag(c *cobra.Command) {
	c.PersistentFlags().BoolVarP(&versionFlag, versionName, versionShortHand, false, versionUsage)
}

func configureVersion(v semver.Version) {
	version = v
	app.AddCommand(versionCmd)
	addVersionFlag(app)
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

func versionRunE(cmd *cobra.Command, args []string) error {
	if _, err := fmt.Fprintln(out, printVersion(version)); err != nil {
		return err
	}
	return nil
}
