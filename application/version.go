package application

import (
	"encoding/json"
	"fmt"

	"github.com/Oudwins/zog"
	"github.com/spf13/cobra"

	"git.flexabyte.io/flexabyte/go-kit/flagzog"
)

const (
	versionName          = "version"
	versionFlagShortCode = "V"
	versionUsage         = "Show version information"
)

var (
	versionFlag = flagzog.NewBoolFlag(versionName, zog.Bool(), versionUsage)
	version     Version
	versionCmd  = &cobra.Command{
		Use:   versionName,
		Short: versionUsage,
		RunE:  versionRunFuncE,
	}
)

type Version struct {
	Full       string
	Branch     string
	Tag        string
	Commit     string
	CommitDate string
	BuildDate  string
	Major      string
	Minor      string
	Patch      string
	PreRelease string
}

func addVersionFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&versionFlag.Value, versionFlag.Name(), versionFlagShortCode, false, versionFlag.Usage())
}

func configureVersionFlag(cmd *cobra.Command, v Version) {
	version = v
	cmd.AddCommand(versionCmd)
	addVersionFlag(cmd)
}

func printVersion(v Version) string {
	var output string
	if !verboseFlag.Value {
		output = v.Full
	}

	if jsonOutputFlag.Value {
		var b []byte
		b, _ = json.Marshal(v)
		output = string(b)
	}

	if output != "" {
		return output
	}
	return fmt.Sprintf(
		"Full: %s\nBranch: %s\nTag: %s\nCommit: %s\nCommit date: %s\nBuild date: %s\nMajor: %s\nMinor: %s\nPatch: %s\nPreRelease: %s\n",
		v.Full,
		v.Branch,
		v.Tag,
		v.Commit,
		v.CommitDate,
		v.BuildDate,
		v.Major,
		v.Minor,
		v.Patch,
		v.PreRelease,
	)
}

func versionRunFuncE(cmd *cobra.Command, args []string) error {
	_, err := fmt.Fprintln(outWriter, printVersion(version))
	return err
}
