package application

import (
	"context"
	"fmt"
	"os/signal"

	"github.com/spf13/cobra"
)

type Application interface {
	Start(ctx context.Context) error
	Shutdown() error
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

func (a *application) Start(ctx context.Context) error {

	sigCtx, sigStop := signal.NotifyContext(ctx, a.config.ShutdownSignals...)
	defer sigStop() // Ensure that this gets called.

	// Result channel for command
	chExe := make(chan error)
	go a.execute(sigCtx, chExe)

	select {
	// sigCtx.Done() returns a channel that will have a message
	// when the context is cancelled. We wait for that signal, which means
	// we received the signal, or our context was cancelled for some other reason.
	case <-sigCtx.Done():
		sigStop()
		return a.Shutdown()
	case err := <-chExe:
		return err
	}
}

func (a *application) Shutdown() error {
	switch a.config.EnableGracefulShutdown {
	case true:
		return a.gracefulShutdown()
	case false:
		return a.shutdown()
		// default:
		// 	return a.shutdown()
	}
	return nil
}

func (a *application) execute(ctx context.Context, chErr chan error) {
	chErr <- a.cmd.ExecuteContext(ctx)
}

func (a *application) gracefulShutdown() error {
	fmt.Println("graceful shutdown")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), a.config.ShutdownTimeout)
	defer shutdownCancel()
	<-shutdownCtx.Done()
	return nil
}

func (a *application) shutdown() error {
	fmt.Println("shutdown")
	return nil
}
