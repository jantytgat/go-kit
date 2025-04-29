package httpd

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"git.flexabyte.io/flexabyte/go-slogd/slogd"
)

func RunHttpServer(ctx context.Context, log *slog.Logger, listenAddress string, port int, h http.Handler, shutdownTimeout time.Duration) error {
	s := &http.Server{
		Addr:    listenAddress + ":" + strconv.Itoa(port),
		Handler: h}

	return run(ctx, s, log, shutdownTimeout)
}

func RunSocketHttpServer(ctx context.Context, log *slog.Logger, socketPath string, h http.Handler, shutdownTimeout time.Duration) error {
	s := &http.Server{
		Handler: h}

	return run(ctx, s, log, shutdownTimeout)
}

func run(ctx context.Context, s *http.Server, log *slog.Logger, shutdownTimeout time.Duration) error {
	log.LogAttrs(ctx, slogd.LevelTrace, "starting http server", slog.String("listenAddress", s.Addr))
	idleConnectionsClosed := make(chan struct{})

	runCtx, runCancel := context.WithCancel(ctx)
	defer runCancel()

	go func(ctx context.Context) {
		log.LogAttrs(ctx, slogd.LevelTrace, "awaiting shutdown signal for http server", slog.String("listenAddress", s.Addr))
		<-ctx.Done()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer shutdownCancel()

		log.LogAttrs(shutdownCtx, slogd.LevelTrace, "shutting down http server", slog.String("listenAddress", s.Addr))
		// We received an interrupt signal, shut down.
		if err := s.Shutdown(shutdownCtx); err != nil {
			// Error from closing listeners, or context timeout:
			log.LogAttrs(ctx, slogd.LevelTrace, "shutting down http server failed", slog.String("listenAddress", s.Addr), slog.Any("error", err))
		}
		log.LogAttrs(shutdownCtx, slogd.LevelTrace, "shutting down http server completed", slog.String("listenAddress", s.Addr))
		close(idleConnectionsClosed)
	}(runCtx)

	var err error
	if err = s.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		// Error starting or closing listener:
		log.LogAttrs(ctx, slogd.LevelError, "http server start failed", slog.String("error", err.Error()))
		return err
	}

	<-idleConnectionsClosed
	return err
}
