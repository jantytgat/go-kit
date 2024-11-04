package application

import (
	"fmt"

	"github.com/jantytgat/go-kit/pkg/semver"
)

var Version string
var version semver.Version

func PrintVersion(banner string, v semver.Version) {
	if banner != "" {
		fmt.Printf(
			"%sFull: %s\nVersion: %s\nChannel: %s\nCommit: %s\nDate: %s\n",
			banner,
			v.String(),
			v.Number(),
			v.Release(),
			v.Commit(),
			v.Date(),
		)
	} else {
		fmt.Printf(
			"Full: %s\nVersion: %s\nChannel: %s\nCommit: %s\nDate: %s\n",
			v.String(),
			v.Number(),
			v.Release(),
			v.Commit(),
			v.Date(),
		)
	}
}
