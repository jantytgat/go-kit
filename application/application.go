package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/samber/oops"
	"github.com/spf13/cobra"

	"github.com/jantytgat/go-kit/slogd"
)

type Application interface {
	ExecuteContext(ctx context.Context) error
}

func New(builder Builder, quitter Quitter) (Application, error) {
	var err error
	if err = builder.Validate(); err != nil {
		return nil, oops.In("application").Wrapf(err, "builder validation failed")
	}

	oopsBuilder := oops.
		In("application").
		Tags(builder.Name)

	if quitter == nil {
		return nil, oopsBuilder.New("quitter is required")
	}

	var cmd *cobra.Command
	if cmd, err = builder.buildCommand(); err != nil {
		return nil, oopsBuilder.Wrapf(err, "application command build failed")
	}

	// var logger *slog.Logger
	// if logger, err = builder.buildLogger(); err != nil {
	// 	return nil, oopsBuilder.Wrapf(err, "application logger build failed")
	// }

	return &application{
		cmd: cmd,
		// logger:  logger,
		quitter: quitter,
		oops:    oopsBuilder,
	}, nil
}

type application struct {
	cmd *cobra.Command
	// logger  *slog.Logger
	quitter Quitter
	oops    oops.OopsErrorBuilder
}

func (a *application) ExecuteContext(ctx context.Context) error {
	// Make the oopsBuilder available through context
	appCtx := oops.WithBuilder(ctx, a.oops)
	signals := a.quitter.ShutdownSignals()

	if signals == nil {
		// a.logger.LogAttrs(appCtx, slogd.LevelTrace, "executing application context without shutdown signals")
		slogd.GetDefaultLogger().LogAttrs(appCtx, slogd.LevelTrace, "executing application context without shutdown signals")
		return a.oops.Wrap(a.cmd.ExecuteContext(appCtx))
	}

	// a.logger.LogAttrs(appCtx, slogd.LevelTrace, "configuring application shutdown signals", slog.Any("signals", signals))
	slogd.GetDefaultLogger().LogAttrs(appCtx, slogd.LevelTrace, "configuring application shutdown signals", slog.Any("signals", signals))
	sigCtx, sigCancel := signal.NotifyContext(appCtx, signals...)
	defer sigCancel() // Ensure that this gets called.

	// Result channel for command output
	chExe := make(chan error)

	// Run the application command using the signal context and output channel
	// a.logger.LogAttrs(appCtx, slogd.LevelTrace, "executing application context with shutdown signals", slog.Any("signals", a.quitter.ShutdownSignals()))
	slogd.GetDefaultLogger().LogAttrs(appCtx, slogd.LevelTrace, "executing application context with shutdown signals", slog.Any("signals", a.quitter.ShutdownSignals()))
	go func(ctx context.Context, chErr chan error) {
		chErr <- a.cmd.ExecuteContext(ctx)
	}(sigCtx, chExe)

	// Wait for command output or a shutdown signal
	select {
	case <-sigCtx.Done(): // sigCtx.Done() returns a channel that will have a message when the context is canceled.
		sigCancel()
		return a.oops.Wrap(a.handleShutdownSignal(appCtx))
	case err := <-chExe: // Alternatively, chExe will receive the response from the execution context if the application finishes.
		// a.logger.LogAttrs(appCtx, slogd.LevelTrace, "application terminated successfully")
		slogd.GetDefaultLogger().LogAttrs(appCtx, slogd.LevelTrace, "application terminated successfully")
		return a.oops.Wrap(err)
	}
}

func (a *application) gracefulShutdown(ctx context.Context) error {
	fmt.Printf("waiting %s for graceful application shutdown... PRESS CTRL+C again to quit now!\n", a.quitter.Timeout())

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, a.quitter.Timeout())
	defer shutdownCancel()

	sig := make(chan os.Signal, 1)
	defer close(sig)

	signal.Notify(sig, a.quitter.ShutdownSignals()...)

	select {
	case <-shutdownCtx.Done(): // Timeout exceeded
		return oops.FromContext(ctx).Wrap(shutdownCtx.Err())
	case s := <-sig: // Additional shutdown signal received
		signal.Stop(sig)
		return oops.FromContext(ctx).With("signal", s).New("process killed manually")
	}
}

func (a *application) handleGracefulShutdown(ctx context.Context) error {
	// a.logger.LogAttrs(ctx, slogd.LevelTrace, "gracefully shutting down application")
	slogd.GetDefaultLogger().LogAttrs(ctx, slogd.LevelTrace, "gracefully shutting down application")

	var err error
	if err = a.gracefulShutdown(ctx); !errors.Is(err, context.DeadlineExceeded) {
		// a.logger.LogAttrs(ctx, slogd.LevelWarn, "graceful shutdown failed", slog.Any("error", err))
		slogd.GetDefaultLogger().LogAttrs(ctx, slogd.LevelWarn, "graceful shutdown failed", slog.Any("error", err))
		return nil
	}

	// a.logger.LogAttrs(ctx, slogd.LevelTrace, "graceful shutdown completed")
	slogd.GetDefaultLogger().LogAttrs(ctx, slogd.LevelTrace, "graceful shutdown completed")
	return nil
}

func (a *application) handleShutdownSignal(ctx context.Context) error {
	if a.quitter == nil {
		return oops.FromContext(ctx).New("no quitter configured")
	}
	// Adapt the shutdown scenario if a graceful shutdown period is configured
	switch a.quitter.IsGraceful() && a.quitter.Timeout() > 0 {
	case true:
		return a.handleGracefulShutdown(ctx)
	case false:
		// a.logger.LogAttrs(ctx, slogd.LevelTrace, "immediately shutting down application")
		slogd.GetDefaultLogger().LogAttrs(ctx, slogd.LevelTrace, "immediately shutting down application")
		return nil
	default:
		panic("cannot handle shutdown signal")
	}
}
