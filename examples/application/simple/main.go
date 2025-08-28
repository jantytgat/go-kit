package main

import (
	"context"
	"fmt"
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
		Name:         "main",
		Title:        "Main Test",
		Banner:       "",
		OverrideRunE: overrideRunFuncE,
		PersistentPreRunE: []func(cmd *cobra.Command, args []string) error{
			simplePersistentPreRunFuncE,
		},
		PersistentPostRunE: []func(cmd *cobra.Command, args []string) error{
			simplePersistentPostRunFuncE,
		},
		SubCommands: nil,
		ValidArgs:   nil,
	}

	var app application.Application
	if app, err = application.New(builder, application.NewDefaultQuitter(application.DefaultShutdownTimeout), slogd.Logger()); err != nil {
		panic(err)
	}

	if err = app.ExecuteContext(ctx); err != nil {
		panic(err)
	}
}

func overrideRunFuncE(cmd *cobra.Command, args []string) error {
	// Set up OpenTelemetry.
	// var err error
	//
	// var traceExporter *otlptrace.Exporter
	// if traceExporter, err = otlptracehttp.New(cmd.Context(), otlptracehttp.WithEndpointURL("http://localhost:4318")); err != nil {
	// 	return err
	// }
	//
	// var metricExporter *otlpmetrichttp.Exporter
	// if metricExporter, err = otlpmetrichttp.New(cmd.Context(), otlpmetrichttp.WithEndpointURL("http://localhost:4318")); err != nil {
	// 	return err
	// }
	//
	// var otelShutdown func(context.Context) error
	// if otelShutdown, err = oteld.SetupOTelSDK(cmd.Context(), traceExporter, metricExporter); err != nil {
	// 	return err
	// }
	// // Handle shutdown properly so nothing leaks.
	// defer func() {
	// 	err = errors.Join(err, otelShutdown(context.Background()))
	// }()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		slogd.FromContext(r.Context()).LogAttrs(r.Context(), slogd.LevelInfo, "request received", slog.String("method", r.Method), slog.String("url", r.URL.String()), slog.String("user-agent", r.UserAgent()))
		fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])
	})
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		slogd.FromContext(r.Context()).LogAttrs(r.Context(), slogd.LevelInfo, "request received", slog.String("method", r.Method), slog.String("url", r.URL.String()), slog.String("user-agent", r.UserAgent()))
		fmt.Fprintf(w, "Hello World, %s!", r.URL.Path[1:])
	})
	// return httpd.RunHttpServer(cmd.Context(), slogd.Logger(), "127.0.0.1", 28000, oteld.EmbedHttpHandler(mux, "/"), 5*time.Second)
	return httpd.RunHttpServer(cmd.Context(), slogd.Logger(), "127.0.0.1", 28000, mux, 5*time.Second)
}

func simplePersistentPreRunFuncE(cmd *cobra.Command, args []string) error {
	slogd.FromContext(cmd.Context()).LogAttrs(cmd.Context(), slogd.LevelDebug, "simplePersistentPreRunFuncE called")
	return nil
}

func simplePersistentPostRunFuncE(cmd *cobra.Command, args []string) error {
	slogd.FromContext(cmd.Context()).LogAttrs(cmd.Context(), slogd.LevelDebug, "simplePersistentPostRunFuncE called")
	return nil
}
