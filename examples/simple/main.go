package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/jantytgat/go-kit/pkg/application"
	"github.com/jantytgat/go-kit/pkg/semver"
	"github.com/jantytgat/go-kit/pkg/slogd"
)

var appTestCmd = application.Command{
	Command:     testCmd,
	SubCommands: nil,
	Configure:   nil,
}
var testCmd = &cobra.Command{
	Use:  "test",
	RunE: testFunc,
}

func testFunc(cmd *cobra.Command, args []string) error {
	slogd.SetLevel(slogd.LevelTrace)
	mux := http.NewServeMux() // Create sample handler to returns 404
	mux.Handle("/", http.RedirectHandler("https://jantytgat.com", 302))
	server := application.NewHttpServer("127.0.0.1", 5600, mux)
	serverCtx, serverCancel := context.WithCancel(cmd.Context())
	defer serverCancel()

	go server.Run(serverCtx)

	var exit bool
	for i := 0; i < 50; i++ {
		if exit {
			break
		}
		select {
		case <-cmd.Context().Done():
			slogd.FromContext(cmd.Context()).Log(cmd.Context(), slogd.LevelWarn, "test context cancelled")
			serverCancel()
			time.Sleep(10 * time.Second)
			// return cmd.Context().Err()
			return nil
		default:
			slogd.FromContext(cmd.Context()).Log(cmd.Context(), slogd.LevelInfo, "test sleeping")
			time.Sleep(100 * time.Millisecond)
		}
	}
	return nil
}

var appTestCmd2 = application.Command{
	Command: testCmd2,
}

var testCmd2 = &cobra.Command{
	Use:  "test2",
	RunE: testFunc2,
}

func testFunc2(cmd *cobra.Command, args []string) error {
	mux := http.NewServeMux() // Create sample handler to returns 404
	mux.Handle("/", http.RedirectHandler("https://jantytgat.com", 302))
	server := application.NewSocketHttpServer("./main.sock", mux)
	serverCtx, serverCancel := context.WithCancel(cmd.Context())
	defer serverCancel()

	server.Run(serverCtx)
	return nil
}

func main() {
	var err error

	slogd.Init(slogd.LevelTrace, false)
	slogd.RegisterColoredTextHandler(os.Stderr, true)
	slogd.RegisterTextHandler(os.Stderr, false)
	slogd.RegisterJSONHandler(os.Stderr, false)

	// You should typically set this using build flags
	var version semver.Version
	if version, err = semver.Parse("0.1.0-alpha.0+metadata.20101112"); err != nil {
		panic(err)
	}

	application.New("example", "Example App", "", version)
	application.RegisterCommands([]application.Commander{appTestCmd, appTestCmd2}, nil)
	ctx := slogd.WithContext(context.Background())

	if err = application.Run(ctx); err != nil {
		slogd.Logger().LogAttrs(ctx, slogd.LevelError, "error running application", slog.Any("error", err))
		os.Exit(1)
	}
	os.Exit(0)
}
