package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jantytgat/go-kit/pkg/application"
)

var rootCmd = &cobra.Command{
	Use: "example",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(cmd.UsageString())
		return nil
	},
}
var banner = "\nExample app\n-----------\n"

func main() {
	var err error
	var app *application.Application

	// You should typically set this using build flags
	application.Version = "0.1.0-alpha.0+metadata.20101112"

	if app, err = application.New(rootCmd, application.WithBanner(banner)); err != nil {
		panic(err)
	}

	if err = app.Run(); err != nil {
		panic(err)
	}
}
