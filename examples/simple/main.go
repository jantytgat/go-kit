package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"syscall"
	"time"

	"git.flexabyte.io/flexabyte/go-slogd/slogd"
	"github.com/spf13/cobra"

	"git.flexabyte.io/flexabyte/go-kit/application"
	"git.flexabyte.io/flexabyte/go-kit/httpd"
)

var (
	version    string = "0.1.0-alpha.0+metadata.20101112"
	branch     string = "0.1.0-dev"
	tag        string = "0.1.0-dev.0"
	commit     string = "aabbccddee"
	commitDate string = time.Now().String()
	buildDate  string = time.Now().String()

	major      string = "0"
	minor      string = "1"
	patch      string = "0"
	prerelease string = "dev"
)

func main() {
	var err error
	slogd.Init(application.GetLogLevelFromArgs(os.Args), false)
	slogd.RegisterSink(slogd.HandlerText, slog.NewTextHandler(os.Stdout, slogd.HandlerOptions()), true)
	ctx := slogd.WithContext(context.Background())

	config := application.Config{
		Name:   "main",
		Title:  "Main Test",
		Banner: "",
		// Version:                "0.1.0-alpha.0+metadata.20101112",
		Version: application.Version{
			Full:       version,
			Branch:     branch,
			Tag:        tag,
			Commit:     commit,
			CommitDate: commitDate,
			BuildDate:  buildDate,
			Major:      major,
			Minor:      minor,
			Patch:      patch,
			PreRelease: prerelease,
		},
		EnableGracefulShutdown: true,
		Logger:                 slogd.Logger(),
		OverrideRunE: func(cmd *cobra.Command, args []string) error {
			// fmt.Println("overrideRunE")
			// time.Sleep(1 * time.Second)
			// fmt.Println("overrideRunE done")
			mux := http.NewServeMux()
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				slogd.FromContext(r.Context()).LogAttrs(r.Context(), slogd.LevelInfo, "request received", slog.String("method", r.Method), slog.String("url", r.URL.String()), slog.String("user-agent", r.UserAgent()))
			})
			return httpd.RunHttpServer(cmd.Context(), slogd.FromContext(cmd.Context()), "127.0.0.1", 28000, mux, 1*time.Second)
		},
		PersistentPreRunE:        nil,
		PersistentPostRunE:       nil,
		ShutdownSignals:          []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT},
		ShutdownTimeout:          5 * time.Second,
		SubCommands:              nil,
		SubCommandInitializeFunc: nil,
		ValidArgs:                nil,
	}
	var app application.Application
	if app, err = application.New(config); err != nil {
		panic(err)
	}

	if err = app.ExecuteContext(ctx); err != nil {
		panic(err)
	}
}
