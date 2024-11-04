package application

import (
	"github.com/spf13/cobra"

	"github.com/jantytgat/go-kit/pkg/semver"
)

func New(c *cobra.Command, opts ...Option) (*Application, error) {
	var err error
	app := &Application{
		RootCmd: c,
	}

	if Version != "" {
		var vOpt Option
		if version, err = semver.Parse(Version); err != nil {
			return nil, err
		}
		vOpt = WithVersion(version)
		if err = vOpt(app); err != nil {
			return nil, err
		}
	}

	for _, opt := range opts {
		if err = opt(app); err != nil {
			return nil, err
		}
	}

	return app, nil
}

type Application struct {
	Banner  string
	RootCmd *cobra.Command
}

func (a *Application) RegisterCommand(c *cobra.Command) {
	if c != nil {
		a.RootCmd.AddCommand(c)
	}
}

func (a *Application) Run() error {
	return a.RootCmd.Execute()
}
