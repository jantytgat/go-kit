package application

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jantytgat/go-kit/pkg/slogd"
)

func NewHttpServer(listenAddress string, port int, h http.Handler) *HttpServer {
	return &HttpServer{
		Server: &http.Server{
			Addr:    listenAddress + ":" + strconv.Itoa(port),
			Handler: h},
		chIdleConnections: make(chan struct{}),
		chWaitForShutdown: make(chan struct{}),
		listenAddress:     strings.Join([]string{listenAddress, strconv.Itoa(port)}, ":"),
	}
}

func NewSocketHttpServer(path string, h http.Handler) *HttpServer {
	return &HttpServer{
		Server: &http.Server{
			Handler: h},
		chIdleConnections: make(chan struct{}),
		chWaitForShutdown: make(chan struct{}),
		socketPath:        path,
	}

}

type HttpServer struct {
	*http.Server
	chIdleConnections chan struct{}
	chWaitForShutdown chan struct{}
	listenAddress     string
	socketPath        string
}

func (s *HttpServer) Run(ctx context.Context) {
	slogd.FromContext(ctx).LogAttrs(ctx, slogd.LevelDebug, "starting http server")

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
				slogd.FromContext(ctx).LogAttrs(ctx, slogd.LevelError, "shutting down http server failed", slog.String("address", "http://"+s.Addr), slog.String("error", shutdownCtx.Err().Error()))
			}
		}(shutdownCtx)

		// Trigger graceful shutdown
		s.chWaitForShutdown <- struct{}{}
		slogd.FromContext(ctx).LogAttrs(ctx, slogd.LevelDebug, "shutting down http server", slog.String("address", "http://"+s.Addr))
		// chExe <- s.Server.Shutdown(shutdownCtx)
		if err := s.Server.Shutdown(shutdownCtx); err != nil {
			slogd.FromContext(ctx).LogAttrs(ctx, slogd.LevelError, "shutting down http server failed", slog.String("error", shutdownCtx.Err().Error()))
		}
		// chExe <- nil
		close(chIdle)
	}(runCtx, chExe, s.chIdleConnections)

	if s.listenAddress != "" {
		go func(ctx context.Context, chExe chan error) {
			slogd.FromContext(ctx).LogAttrs(ctx, slogd.LevelDebug, "http server started", slog.String("address", "http://"+s.Addr))
			if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				slogd.FromContext(ctx).LogAttrs(ctx, slogd.LevelError, "http server start failed", slog.String("error", err.Error()))
				chExe <- err
			}
		}(runCtx, chExe)
	} else if s.socketPath != "" {
		go func(ctx context.Context, chExe chan error) {
			var err error
			var config = new(net.ListenConfig)
			var socket net.Listener

			if socket, err = config.Listen(ctx, "unix", s.socketPath); err != nil {
				slogd.FromContext(ctx).LogAttrs(ctx, slogd.LevelError, "failed to listen on socket", slog.String("error", err.Error()))
				chExe <- err
				return
			}

			slogd.FromContext(ctx).LogAttrs(ctx, slogd.LevelDebug, "http server started", slog.String("socket", s.socketPath))
			if err = s.Serve(socket); err != nil && !errors.Is(err, http.ErrServerClosed) {
				slogd.FromContext(ctx).LogAttrs(ctx, slogd.LevelError, "http server start failed", slog.String("error", err.Error()))
				chExe <- err
			}
		}(runCtx, chExe)
	}

	for {
		select {
		case <-s.chWaitForShutdown:
			slogd.FromContext(ctx).LogAttrs(ctx, slogd.LevelDebug, "waiting for idle connections to close")
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
