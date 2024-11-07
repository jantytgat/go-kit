package application

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/jantytgat/go-kit/pkg/slogd"
)

type HttpServer struct {
	*http.Server
	chIdleConnections chan struct{}
	chWaitForShutdown chan struct{}
}

func NewHttpServer(listenAddress string, port int, h http.Handler) *HttpServer {
	return &HttpServer{
		Server: &http.Server{
			Addr:    listenAddress + ":" + strconv.Itoa(port),
			Handler: h},
		chIdleConnections: make(chan struct{}),
		chWaitForShutdown: make(chan struct{}),
	}
}

func (s *HttpServer) Run(ctx context.Context) {
	slogd.FromContext(ctx).LogAttrs(ctx, slogd.LevelDebug.Level(), "starting http server")

	// HttpServer run context
	runCtx, runCancel := context.WithCancel(ctx)
	defer runCancel()

	chExe := make(chan error)

	// Goroutine to handle graceful shutdown
	go func(ctx context.Context, chExe chan error, chIdle chan struct{}) {
		<-ctx.Done()

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 30*time.Second)
		defer shutdownCancel()

		go func(ctx context.Context) {
			<-ctx.Done()
			if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
				slogd.FromContext(ctx).LogAttrs(ctx, slogd.LevelError.Level(), "shutting down http server failed", slog.String("address", "http://"+s.Addr), slog.String("error", shutdownCtx.Err().Error()))
			}
		}(shutdownCtx)

		// Trigger graceful shutdown
		s.chWaitForShutdown <- struct{}{}
		slogd.FromContext(ctx).LogAttrs(ctx, slogd.LevelDebug.Level(), "shutting down http server", slog.String("address", "http://"+s.Addr))
		// chExe <- s.Server.Shutdown(shutdownCtx)
		if err := s.Server.Shutdown(shutdownCtx); err != nil {
			slogd.FromContext(ctx).LogAttrs(ctx, slogd.LevelError.Level(), "shutting down http server failed", slog.String("error", shutdownCtx.Err().Error()))
		}
		// chExe <- nil
		close(chIdle)
	}(runCtx, chExe, s.chIdleConnections)

	go func(ctx context.Context, chExe chan error) {
		slogd.FromContext(ctx).LogAttrs(ctx, slogd.LevelDebug.Level(), "http server started", slog.String("address", "http://"+s.Addr))
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slogd.FromContext(ctx).LogAttrs(ctx, slogd.LevelError.Level(), "http server start failed", slog.String("error", err.Error()))
			chExe <- err
		}
	}(runCtx, chExe)

	for {
		select {
		case <-s.chWaitForShutdown:
			slogd.FromContext(ctx).LogAttrs(ctx, slogd.LevelDebug.Level(), "waiting for idle connections to close")
			select {
			case <-s.chIdleConnections:
				return
			}
		case <-chExe:
			return
		}
	}
}

func (s *HttpServer) shutdown(c context.Context, cancelFunc context.CancelFunc) {

	// Shutdown signal with grace period of 30 seconds
	shutdownCtx, cancel := context.WithTimeout(c, 30*time.Second)
	defer cancel()

	go func() {
		<-shutdownCtx.Done()
		fmt.Printf("\nHttpServer shutdown started\n")
		if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
			log.Fatal("graceful shutdown timed out.. forcing exit.")
		}
	}()

	// Trigger graceful shutdown
	err := s.Server.Shutdown(shutdownCtx)
	if err != nil {
		log.Fatal(err)
	}

	cancelFunc()
}
