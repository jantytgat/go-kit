package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"git.flexabyte.io/flexabyte/go-slogd/slogd"

	"git.flexabyte.io/flexabyte/go-kit/application"
)

func main() {
	var err error
	slogd.Init(slogd.LevelTrace, true)

	config := application.Config{
		Name:                   "main",
		Title:                  "Main Test",
		Banner:                 "",
		Version:                "0.1.0-alpha.0+metadata.20101112",
		EnableGracefulShutdown: true,
		OverrideRunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("overrideRunE")
			time.Sleep(5 * time.Second)
			fmt.Println("overrideRunE done")
			return nil
		},
		PersistentPreRunE:  nil,
		PersistentPostRunE: nil,
		// ShutdownSignals:          []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT},
		ShutdownTimeout:          1 * time.Second,
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
