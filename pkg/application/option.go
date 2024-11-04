package application

import (
	"github.com/spf13/cobra"

	"github.com/jantytgat/go-kit/pkg/semver"
)

type Option func(*Application) error

func WithBanner(banner string) Option {
	return func(a *Application) error {
		a.Banner = banner
		return nil
	}
}

func WithVersion(v semver.Version) Option {
	return func(a *Application) error {
		a.RegisterCommand(&cobra.Command{
			Use:   "version",
			Short: "Show version information",
			RunE: func(cmd *cobra.Command, args []string) error {
				PrintVersion(a.Banner, v)
				return nil
			},
		})
		return nil
	}
}
