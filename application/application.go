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

func New(builder Builder, quitter Quitter) (Application, error) {
	var err error
	if err = builder.Validate(); err != nil {
		return nil, oops.In("application").Wrapf(err, "builder validation failed")
	}

	if quitter == nil {
		return nil, oops.In("application").New("quitter is required")
	}

	var cmd *cobra.Command
	if cmd, err = builder.buildCommand(); err != nil {
		return nil, oops.In("application").Wrapf(err, "application command build failed")
	}

	return &application{
		cmd:     cmd,
		quitter: quitter,
		chCmd:   make(chan error, 1),
		chOut:   make(chan error, 1),
		chSig:   make(chan os.Signal, 1),
	}, nil
}

type Application interface {
	ExecuteContext(ctx context.Context) error
}

type application struct {
	cmd     *cobra.Command
	quitter Quitter
	oops    oops.OopsErrorBuilder
	chCmd   chan error
	chOut   chan error
	chSig   chan os.Signal
}

func (a *application) ExecuteContext(ctx context.Context) error {
	// Make the oopsBuilder available through context and create cancellable context for application execution
	a.oops = oops.
		In("application").
		Tags(a.cmd.Name()).
		With("version", version)
	oopsCtx := oops.WithBuilder(ctx, a.oops)

	// Create cancellable context for application execution
	appCtx, appCancel := context.WithCancel(oopsCtx)
	defer appCancel()

	// Run the application command using the signal context and output channel
	go a.processOutput(oopsCtx, appCancel) // Process output using original context, as appCancel is called in processOutput, cancelling the context
	go a.launch(appCtx)                    // Launch the Cobra command using the cancellable context

	return <-a.chOut
}

func (a *application) launch(ctx context.Context) {
	slogd.GetDefaultLogger().Log(ctx, slogd.LevelTrace, "starting cobra command")
	a.chCmd <- a.cmd.ExecuteContext(ctx)
}

func (a *application) processOutput(ctx context.Context, appCancel context.CancelFunc) {
	var err error

	if !a.quitter.HasSignals() {
		slogd.GetDefaultLogger().LogAttrs(ctx, slogd.LevelTrace, "process output without shutdown signals")
		err = <-a.chCmd
	} else {
		slogd.GetDefaultLogger().LogAttrs(ctx, slogd.LevelTrace, "process output with shutdown signals", slog.Any("signals", a.quitter.ShutdownSignals()))
		err = a.processOutputWithSignals(ctx, appCancel)
	}
	a.chOut <- err
}

func (a *application) processOutputWithSignals(ctx context.Context, appCancel context.CancelFunc) error {
	var err error
	// Configure application shutdown signals
	signals := a.quitter.ShutdownSignals()
	signal.Notify(a.chSig, signals...)

	chShutdown := make(chan error)
	shutdownCtx, shutdownCancel := context.WithCancel(ctx)
	defer shutdownCancel()

	// Wait for command output or a shutdown signal
	select {
	case sig := <-a.chSig: // sigCtx.Done() returns a channel that will have a message when the context is canceled.
		slogd.GetDefaultLogger().LogAttrs(ctx, slogd.LevelTrace, "received shutdown signal", slog.Any("signal", sig))

		go a.handleShutdownSignal(shutdownCtx, chShutdown)
		appCancel()

		select {
		case err = <-a.chCmd:
			slogd.GetDefaultLogger().LogAttrs(ctx, slogd.LevelTrace, "cobra command finished successfully before graceful shutdown deadline")
			shutdownCancel()
		case err = <-chShutdown:
			slogd.GetDefaultLogger().LogAttrs(ctx, slogd.LevelTrace, "application shutdown signal processed")
		}
	case err = <-a.chCmd: // Alternatively, chCmd will receive the response from the execution context if the application finishes.
		slogd.GetDefaultLogger().LogAttrs(ctx, slogd.LevelTrace, "cobra command finished successfully")
	}
	return err
}

func (a *application) handleShutdownSignal(ctx context.Context, ch chan error) {
	if a.quitter == nil {
		ch <- oops.FromContext(ctx).New("no quitter configured")
	}
	// Adapt the shutdown scenario if a graceful shutdown period is configured
	switch a.quitter.IsGraceful() && a.quitter.Timeout() > 0 {
	case true:
		slogd.GetDefaultLogger().LogAttrs(ctx, slogd.LevelTrace, "shutting down application gracefully")
		select {
		case <-ctx.Done():
		case ch <- a.startGracefulShutdown(ctx):
		}
	case false:
		slogd.GetDefaultLogger().LogAttrs(ctx, slogd.LevelTrace, "shutting down application immediately")
		ch <- nil
	default:
		panic("cannot handle shutdown signal")
	}
}

func (a *application) startGracefulShutdown(ctx context.Context) error {
	var err error
	if err = a.waitForGracefulShutdown(ctx); err != nil && !errors.Is(err, context.DeadlineExceeded) {
		slogd.GetDefaultLogger().LogAttrs(ctx, slogd.LevelWarn, "graceful shutdown failed", slog.Any("error", err))
		return oops.FromContext(ctx).Wrap(err)
	} else if err != nil && errors.Is(err, context.DeadlineExceeded) {
		slogd.GetDefaultLogger().LogAttrs(ctx, slogd.LevelWarn, "graceful shutdown deadline exceeded")
		return oops.FromContext(ctx).Wrap(err)
	}
	return nil
}

func (a *application) waitForGracefulShutdown(ctx context.Context) error {
	fmt.Printf("\nwaiting %s for graceful application shutdown... PRESS CTRL+C again to quit now!\n\n", a.quitter.Timeout())

	gracefulCtx, gracefulCancel := context.WithTimeout(ctx, a.quitter.Timeout())
	defer gracefulCancel()

	sig := make(chan os.Signal, 1)
	defer close(sig)

	signal.Notify(sig, a.quitter.ShutdownSignals()...)
	defer signal.Stop(sig)

	select {
	case <-gracefulCtx.Done(): // Timeout exceeded
		return oops.FromContext(ctx).Wrap(gracefulCtx.Err())
	case s := <-sig: // Additional shutdown signal received
		slogd.GetDefaultLogger().LogAttrs(ctx, slogd.LevelWarn, "graceful application shutdown override", slog.Any("signal", s))
		return nil
	}
}
