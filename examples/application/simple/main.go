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

	slogd.All().WithDefaultFlow(
		slogd.NewFlow("stdout", slogd.FlowFanOut).
			WithHandler("stdout", slogd.NewDefaultTextHandler("stdout", os.Stdout, application.GetLogLevelFromArgs(os.Args, slogd.LevelDefault), false)))

	builder := application.Builder{
		Name:   "main",
		Title:  "Main Test",
		Banner: "################\n##### TEST #####\n################\n",
		//OverrideRunE: overrideRunFuncE,
		ConfigureRoot: func(cmd *cobra.Command) {
			cmd.Flags().StringP("username", "u", "", "username")
			cmd.Flags().StringP("password", "p", "", "password")
			cmd.Flags().StringP("test", "t", "", "test")
		},
		PersistentPreRunE: []func(cmd *cobra.Command, args []string) error{
			simplePersistentPreRunFuncE,
		},
		PersistentPostRunE: []func(cmd *cobra.Command, args []string) error{
			simplePersistentPostRunFuncE,
		},
		SubCommands:              nil,
		SubCommandsBannerEnabled: true,
		ParseArgsFromStdin:       true,
		ValidArgs:                nil,
		PersistentFlags:          application.PersistentFlagsDefault,
		EnableVersionCommand:     true,
	}

	var app application.Application
	if app, err = application.New(builder, application.NewDefaultQuitter(application.DefaultShutdownTimeout)); err != nil {
		panic(err)
	}
	// if app, err = application.New(builder, application.NewQuitter(nil, application.DefaultShutdownTimeout, false)); err != nil {
	// 	panic(err)
	// }
	if err = app.ExecuteContext(context.Background()); err != nil {
		slogd.GetDefaultLogger().Error("application exited with errors", slog.Any("error", err))
	}
}

func overrideRunFuncE(cmd *cobra.Command, args []string) error {
	// Set up OpenTelemetry.
	// var err error
	//
	// var traceExporter *otlptrace.Exporter
	// if traceExporter, err = otlptracehttp.NewFlow(cmd.Context(), otlptracehttp.WithEndpointURL("http://localhost:4318")); err != nil {
	// 	return err
	// }
	//
	// var metricExporter *otlpmetrichttp.Exporter
	// if metricExporter, err = otlpmetrichttp.NewFlow(cmd.Context(), otlpmetrichttp.WithEndpointURL("http://localhost:4318")); err != nil {
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

	fmt.Println("overrideRunFuncE called")

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		slogd.FromContext(r.Context()).Logger(slogd.GetDefaultFlowName()).LogAttrs(r.Context(), slogd.LevelInfo, "request received", slog.String("method", r.Method), slog.String("url", r.URL.String()), slog.String("user-agent", r.UserAgent()))
		fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])
	})
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		slogd.FromContext(r.Context()).Logger(slogd.GetDefaultFlowName()).LogAttrs(r.Context(), slogd.LevelInfo, "request received", slog.String("method", r.Method), slog.String("url", r.URL.String()), slog.String("user-agent", r.UserAgent()))
		fmt.Fprintf(w, "Hello World, %s!", r.URL.Path[1:])
	})

	httpCtx, httpCancel := context.WithTimeout(cmd.Context(), 10*time.Second)
	defer httpCancel()
	return httpd.RunHttpServer(httpCtx, slogd.FromContext(cmd.Context()).DefaultLogger(), "127.0.0.1", 28000, mux, 5*time.Second)

}

func simplePersistentPreRunFuncE(cmd *cobra.Command, args []string) error {
	slogd.All().Logger("stdout").LogAttrs(cmd.Context(), slogd.LevelDebug, "simplePersistentPreRunFuncE called", slog.String("command", cmd.CommandPath()))
	return nil
}

func simplePersistentPostRunFuncE(cmd *cobra.Command, args []string) error {
	slogd.All().Logger("stdout").LogAttrs(cmd.Context(), slogd.LevelDebug, "simplePersistentPostRunFuncE called", slog.String("command", cmd.CommandPath()))
	return nil
}
