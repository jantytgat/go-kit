package application

import (
	"os"
	"time"
)

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
