package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"github.com/jantytgat/go-kit/pkg/application"
	"github.com/jantytgat/go-kit/pkg/semver"
	"github.com/jantytgat/go-kit/pkg/slogd"
)

var testCmd = &cobra.Command{
	Use:  "test",
	RunE: testFunc,
}

func testFunc(cmd *cobra.Command, args []string) error {
	out, mux := application.Output()
	mux.Lock()
	defer mux.Unlock()
	if _, err := fmt.Fprintln(out, "test"); err != nil {
		return err
	}
	slogd.FromContext(cmd.Context()).Log(cmd.Context(), slogd.LevelWarn.Level(), "test")
	return nil
}

func main() {
	var err error

	slogd.Init(slogd.LevelTrace.Level(), false)
	slogd.RegisterColoredTextHandler(os.Stderr, true)
	slogd.RegisterTextHandler(os.Stderr, false)
	slogd.RegisterJSONHandler(os.Stderr, false)

	// You should typically set this using build flags
	var version semver.Version
	if version, err = semver.Parse("0.1.0-alpha.0+metadata.20101112"); err != nil {
		panic(err)
	}

	application.New("example", "Example App", "", version)
	application.RegisterCommand(testCmd)
	ctx := slogd.WithContext(context.Background())

	if err = application.Run(ctx); err != nil {
		slogd.Logger().LogAttrs(ctx, slogd.LevelError.Level(), "error running application", slog.Any("error", err))
	}
	// slogd.Logger().LogAttrs(ctx, slogd.LevelDebug.Level(), "test coloured")
	// slogd.UseHandler(slogd.HandlerText)
	// slogd.Logger().LogAttrs(ctx, slogd.LevelDebug.Level(), "test after change to text")
	// slogd.SetLevel(slogd.LevelInfo)
	// slogd.UseHandler(slogd.HandlerJSON)
	// slogd.Logger().Log(ctx, slogd.LevelInfo.Level(), "test after level change to json")
	// slogd.UseHandler(slogd.HandlerColor)
	// slogd.Logger().Log(ctx, slogd.LevelInfo.Level(), "test after level change")
}
