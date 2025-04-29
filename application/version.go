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
	version     Version
	versionFlag bool
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
	cmd.PersistentFlags().BoolVarP(&versionFlag, versionName, versionShortHand, false, versionUsage)
}

func configureVersionFlag(cmd *cobra.Command, v Version) {
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
		"Full: %s\nBranch: %s\nTag: %s\nCommit: %s\nCommit date: %s\nBuild date: %s\nMajor: %s\nMinor: %s\nPatch: %s\nPreRelease: %s\n",
		version.Full,
		version.Branch,
		version.Tag,
		version.Commit,
		version.CommitDate,
		version.BuildDate,
		version.Major,
		version.Minor,
		version.Patch,
		version.PreRelease,
	)
}

func versionRunFuncE(cmd *cobra.Command, args []string) error {
	var v semver.Version
	var err error
	if v, err = semver.Parse(version.Full); err != nil {
		return err
	}
	if _, err = fmt.Fprintln(outWriter, printVersion(v)); err != nil {
		return err
	}
	return nil
}
