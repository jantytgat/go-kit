package application

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/Oudwins/zog"
	"github.com/spf13/cobra"

	"github.com/jantytgat/go-kit/flagzog"
)

const (
	versionFlagName      = "version"
	versionFlagShortCode = "V"
	versionFlagUsage     = "Show version information"
	versionFlagDefault   = false

	// https://semver.org/ && https://regex101.com/r/Ly7O1x/3/
	validSemVer = `^(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`
)

var (
	versionFlag = flagzog.NewBoolFlag(versionFlagName, zog.Bool(), versionFlagUsage)
	version     Version
	versionCmd  = &cobra.Command{
		Use:   versionFlagName,
		Short: versionFlagUsage,
		RunE:  versionRunFuncE,
	}

	regexSemver = regexp.MustCompile(validSemVer)
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

func (v Version) IsValid() bool {
	return regexSemver.MatchString(v.Full)
}

func addVersionFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&versionFlag.Value, versionFlag.Name(), versionFlagShortCode, versionFlagDefault, versionFlag.Usage())
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
