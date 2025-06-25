package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"

	"github.com/jantytgat/go-kit/slogd"
)

type Application interface {
	ExecuteContext(ctx context.Context) error
}

type Configuration interface {
	GetRootCommand() (*cobra.Command, error)
}

type Quitter interface {
	ShutdownSignals() []os.Signal
	Timeout() time.Duration
	IsGraceful() bool
}

func NewDefaultQuitter(timeout time.Duration) Quitter {
	return quitter{
		signals:  DefaultShutdownSignals,
		timeout:  timeout,
		graceful: true,
	}
}

func NewQuitter(signals []os.Signal, timeout time.Duration, graceful bool) Quitter {
	return quitter{
		signals:  signals,
		timeout:  timeout,
		graceful: graceful,
	}
}

type quitter struct {
	signals  []os.Signal
	timeout  time.Duration
	graceful bool
}

func (q quitter) ShutdownSignals() []os.Signal {
	return q.signals
}

func (q quitter) Timeout() time.Duration {
	return q.timeout
}

func (q quitter) IsGraceful() bool {
	return q.graceful
}

func New(cmd *cobra.Command, quitter Quitter, logger *slog.Logger) (Application, error) {
	if logger == nil {
		return nil, errors.New("logger is required")
	}

	return &application{
		cmd:    cmd,
		logger: logger,
		// config:  config,
		quitter: quitter,
	}, nil
}

type application struct {
	cmd    *cobra.Command
	logger *slog.Logger
	// config  Configuration
	quitter Quitter
}

func (a *application) ExecuteContext(ctx context.Context) error {
	signals := a.quitter.ShutdownSignals()

	if signals == nil {
		a.logger.LogAttrs(ctx, slogd.LevelTrace, "executing application context without shutdown signals")
		return a.cmd.ExecuteContext(ctx)
	}

	a.logger.LogAttrs(ctx, slogd.LevelTrace, "configuring application shutdown signals", slog.Any("signals", signals))
	sigCtx, sigCancel := signal.NotifyContext(ctx, signals...)
	defer sigCancel() // Ensure that this gets called.

	// Result channel for command output
	chExe := make(chan error)

	// Run the application command using the signal context and output channel
	a.logger.LogAttrs(ctx, slogd.LevelTrace, "executing application context with shutdown signals", slog.Any("signals", a.quitter.ShutdownSignals()))
	go func(ctx context.Context, chErr chan error) {
		chErr <- a.cmd.ExecuteContext(ctx)
	}(sigCtx, chExe)

	// Wait for command output or a shutdown signal
	select {
	case <-sigCtx.Done(): // sigCtx.Done() returns a channel that will have a message when the context is canceled.
		sigCancel()
		return a.handleShutdownSignal(ctx)
	case err := <-chExe: // Alternatively, chExe will receive the response from the execution context if the application finishes.
		a.logger.LogAttrs(ctx, slogd.LevelTrace, "application terminated successfully")
		return err
	}
}

func (a *application) gracefulShutdown(ctx context.Context) error {
	fmt.Printf("waiting %s for graceful application shutdown... PRESS CTRL+C again to quit now!\n", a.quitter.Timeout())

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, a.quitter.Timeout())
	defer shutdownCancel()

	// Wait for the shutdown timeout or a hard exit signal
	sigCtx, sigCancel := signal.NotifyContext(shutdownCtx, a.quitter.ShutdownSignals()...)
	defer sigCancel() // Ensure that this gets called.

	select {
	case <-shutdownCtx.Done(): // Timeout exceeded
		return shutdownCtx.Err()
	case <-sigCtx.Done(): // Received additional shutdown signal to forcefully exit
		fmt.Println("exiting...")
		sigCancel()
		return fmt.Errorf("process killed")
	}
}

func (a *application) handleGracefulShutdown(ctx context.Context) error {
	a.logger.LogAttrs(ctx, slogd.LevelTrace, "gracefully shutting down application")

	var err error
	if err = a.gracefulShutdown(ctx); !errors.Is(err, context.DeadlineExceeded) {
		a.logger.LogAttrs(ctx, slogd.LevelWarn, "graceful shutdown failed", slog.Any("error", err))
		return nil
	}

	a.logger.LogAttrs(ctx, slogd.LevelTrace, "graceful shutdown completed")
	return nil
}

func (a *application) handleShutdownSignal(ctx context.Context) error {
	if a.quitter == nil {
		return fmt.Errorf("no quitter configured")
	}
	// Adapt the shutdown scenario if a graceful shutdown period is configured
	switch a.quitter.IsGraceful() && a.quitter.Timeout() > 0 {
	case true:
		return a.handleGracefulShutdown(ctx)
	case false:
		a.logger.LogAttrs(ctx, slogd.LevelTrace, "immediately shutting down application")
		return nil
	default:
		panic("cannot handle shutdown signal")
	}
}
