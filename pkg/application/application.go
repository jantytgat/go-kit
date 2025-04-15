package application

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

type Application interface {
	Start(ctx context.Context) error
	Shutdown() error
}

func New(c Config) (Application, error) {
	var cmd *cobra.Command
	var err error
	if cmd, err = c.getRootCommand(); err != nil {
		return nil, err
	}

	return &application{
		cmd: cmd,
	}, nil
}

type application struct {
	cmd *cobra.Command
}

func (a *application) Start(ctx context.Context) error {
	return a.cmd.ExecuteContext(ctx)
}

func (a *application) Shutdown() error {
	fmt.Println("Shutdown")
	return nil
}
