package httpd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"git.flexabyte.io/flexabyte/go-kit/slogd"
)

func RunHttpServer(ctx context.Context, log *slog.Logger, listenAddress string, port int, h http.Handler, shutdownTimeout time.Duration) error {
	s := &http.Server{
		Addr:    listenAddress + ":" + strconv.Itoa(port),
		Handler: h}

	log.LogAttrs(ctx, slogd.LevelTrace, "starting http server", slog.String("listenAddress", fmt.Sprintf("http://%s", s.Addr)))

	shutdownCtx, shutdownCancel := context.WithCancel(ctx)
	defer shutdownCancel()

	// Run goroutine to handle graceful shutdown
	idleConnectionsClosed := make(chan struct{})
	go shutdown(shutdownCtx, log, s, shutdownTimeout, idleConnectionsClosed)

	var err error
	if err = s.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		// Error starting or closing listener:
		log.LogAttrs(ctx, slogd.LevelError, "http server start failed", slog.String("error", err.Error()))
		return err
	}

	<-idleConnectionsClosed
	return err
}

func RunSocketHttpServer(ctx context.Context, log *slog.Logger, socketPath string, h http.Handler, shutdownTimeout time.Duration) error {
	s := &http.Server{
		Handler: h}

	log.LogAttrs(ctx, slogd.LevelTrace, "starting http server", slog.String("socket", s.Addr))

	shutdownCtx, shutdownCancel := context.WithCancel(ctx)
	defer shutdownCancel()

	// Run goroutine to handle graceful shutdown
	idleConnectionsClosed := make(chan struct{})
	go shutdown(shutdownCtx, log, s, shutdownTimeout, idleConnectionsClosed)

	var err error
	var config = new(net.ListenConfig)
	var socket net.Listener

	if socket, err = config.Listen(ctx, "unix", socketPath); err != nil {
		log.LogAttrs(ctx, slogd.LevelError, "failed to listen on socket", slog.String("error", err.Error()))
		return err
	}

	if err = s.Serve(socket); err != nil && !errors.Is(err, http.ErrServerClosed) {
		// Error starting or closing listener:
		log.LogAttrs(ctx, slogd.LevelError, "http server start failed", slog.String("error", err.Error()))
		return err
	}

	<-idleConnectionsClosed
	return err
}

func shutdown(ctx context.Context, log *slog.Logger, s *http.Server, shutdownTimeout time.Duration, idleConnectionsClosed chan struct{}) {
	log.LogAttrs(ctx, slogd.LevelTrace, "awaiting shutdown signal for http server", slog.String("listenAddress", s.Addr))
	<-ctx.Done()

	// When shutdown signal is received, create a new context with the configured shutdown timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	log.LogAttrs(shutdownCtx, slogd.LevelTrace, "shutdown signal received for http server", slog.String("listenAddress", s.Addr))
	time.Sleep(2 * time.Second)
	// We received an interrupt signal, shut down.
	if err := s.Shutdown(shutdownCtx); err != nil {
		// Error from closing listeners, or context timeout:
		log.LogAttrs(ctx, slogd.LevelTrace, "shutdown for http server failed", slog.String("listenAddress", s.Addr), slog.Any("error", err))
	}
	log.LogAttrs(shutdownCtx, slogd.LevelTrace, "shutdown for http server completed", slog.String("listenAddress", s.Addr))
	close(idleConnectionsClosed)
}
