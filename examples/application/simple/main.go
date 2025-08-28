package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/jantytgat/go-kit/application"
	"github.com/jantytgat/go-kit/httpd"
	"github.com/jantytgat/go-kit/slogd"
)

func main() {
	var err error
	slogd.Init(application.GetLogLevelFromArgs(os.Args), false)
	slogd.RegisterSink(slogd.HandlerText, slog.NewTextHandler(os.Stdout, slogd.HandlerOptions()), true)
	ctx := slogd.WithContext(context.Background())

	builder := application.Builder{
		Name:   "main",
		Title:  "Main Test",
		Banner: "",
		OverrideRunE: func(cmd *cobra.Command, args []string) error {
			mux := http.NewServeMux()
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				slogd.FromContext(r.Context()).LogAttrs(r.Context(), slogd.LevelInfo, "request received", slog.String("method", r.Method), slog.String("url", r.URL.String()), slog.String("user-agent", r.UserAgent()))
			})
			return httpd.RunHttpServer(cmd.Context(), slogd.Logger(), "127.0.0.1", 28000, mux, 5*time.Second)
		},
		PersistentPreRunE:  nil,
		PersistentPostRunE: nil,
		SubCommands:        nil,
		ValidArgs:          nil,
	}

	var cmd *cobra.Command
	if cmd, err = builder.Build(); err != nil {
		panic(err)
	}
	var app application.Application
	if app, err = application.New(cmd, application.NewDefaultQuitter(application.DefaultShutdownTimeout), slogd.Logger()); err != nil {
		panic(err)
	}

	if err = app.ExecuteContext(ctx); err != nil {
		panic(err)
	}
}
