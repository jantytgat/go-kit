package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/signal"

	"github.com/spf13/cobra"

	"git.flexabyte.io/flexabyte/go-kit/slogd"
)

type Application interface {
	ExecuteContext(ctx context.Context) error
}

func New(c Config) (Application, error) {
	var cmd *cobra.Command
	var err error
	if cmd, err = c.getRootCommand(); err != nil {
		return nil, err
	}

	return &application{
		cmd:    cmd,
		config: c,
	}, nil
}

type application struct {
	cmd    *cobra.Command
	config Config
}

func (a *application) ExecuteContext(ctx context.Context) error {
	a.config.Logger.LogAttrs(ctx, slogd.LevelTrace, "configuring application shutdown signals", slog.Any("signals", a.config.ShutdownSignals))

	sigCtx, sigCancel := signal.NotifyContext(ctx, a.config.ShutdownSignals...)
	defer sigCancel() // Ensure that this gets called.

	// Result channel for command output
	chExe := make(chan error)

	// Run the application command
	a.config.Logger.LogAttrs(ctx, slogd.LevelTrace, "executing application context")
	go func(ctx context.Context, chErr chan error) {
		chErr <- a.cmd.ExecuteContext(ctx)
	}(sigCtx, chExe)

	// Wait for command output or a shutdown signal
	var err error
	select {
	// sigCtx.Done() returns a channel that will have a message when the context is canceled.
	// Alternatively, chExe will receive the response from the execution context if the application finishes.
	case <-sigCtx.Done():
		sigCancel()
		return a.handleShutdownSignal(ctx)
	case err = <-chExe:
		a.config.Logger.LogAttrs(ctx, slogd.LevelTrace, "application terminated successfully")
		return err
	}
}

func (a *application) handleShutdownSignal(ctx context.Context) error {
	var err error
	// Adapt the shutdown scenario if a graceful shutdown period is configured
	switch a.config.EnableGracefulShutdown && a.config.ShutdownTimeout > 0 {
	case true:
		a.config.Logger.LogAttrs(ctx, slogd.LevelTrace, "gracefully shutting down application")
		if err = a.gracefulShutdown(ctx); !errors.Is(err, context.DeadlineExceeded) {
			a.config.Logger.LogAttrs(ctx, slogd.LevelTrace, "graceful shutdown failed", slog.Any("error", err))
			return nil
		}
		a.config.Logger.LogAttrs(ctx, slogd.LevelTrace, "graceful shutdown completed")
		return nil
	case false:
		a.config.Logger.LogAttrs(ctx, slogd.LevelTrace, "immediately shutting down application")
		return nil
	default:
		panic("cannot handle shutdown signal")
	}
}

func (a *application) gracefulShutdown(ctx context.Context) error {
	fmt.Printf("waiting %s for graceful application shutdown... PRESS CTRL+C again to quit now!\n", a.config.ShutdownTimeout)

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, a.config.ShutdownTimeout)
	defer shutdownCancel()

	// Wait for the shutdown timeout or a hard exit signal
	sigCtx, sigCancel := signal.NotifyContext(shutdownCtx, a.config.ShutdownSignals...)
	defer sigCancel() // Ensure that this gets called.

	select {
	case <-shutdownCtx.Done():
		return shutdownCtx.Err()
	case <-sigCtx.Done():
		fmt.Println("exiting...")
		sigCancel()
		return fmt.Errorf("process killed")
	}
}
