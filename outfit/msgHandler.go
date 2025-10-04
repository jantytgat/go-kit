package outfit

import (
	"log/slog"

	"github.com/nats-io/nats.go"
)

type MsgHandler interface {
	Handle(msg *nats.Msg)
	Handler() func(msg *nats.Msg)
	IsQueueHandler() bool
	MaxWorkers() int
	Name() string
	Queue() string
	Subject() string
	UpdateLogger(logger *slog.Logger)
}
