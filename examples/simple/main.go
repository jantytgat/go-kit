package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"git.flexabyte.io/flexabyte/go-slogd/pkg/slogd"

	"git.flexabyte.io/flexabyte/go-kit/pkg/application"
)

func main() {
	var err error
	slogd.Init(slogd.LevelTrace, true)

	config := application.Config{
		Name:                   "main",
		Title:                  "Main Test",
		Banner:                 "",
		Version:                "0.1.0-alpha.0+metadata.20101112",
		EnableGracefulShutdown: false,
		OverrideRunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("overrideRunE")
			return nil
		},
		PersistentPreRunE:        nil,
		PersistentPostRunE:       nil,
		ShutdownSignals:          nil,
		ShutdownTimeout:          0,
		SubCommands:              nil,
		SubCommandInitializeFunc: nil,
		ValidArgs:                nil,
	}

	var app application.Application
	if app, err = application.New(config); err != nil {
		panic(err)
	}

	if err = app.Start(context.Background()); err != nil {
		panic(err)
	}
}
