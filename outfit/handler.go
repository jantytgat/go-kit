package outfit

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go"

	"github.com/jantytgat/go-kit/slogd"
)

func NewHandler(name string, handler MsgHandler, logger *slog.Logger) *Handler {
	chMsg := make(chan *nats.Msg, handler.MaxWorkers())
	fullName := fmt.Sprintf("%s-%s", name, handler.Name())
	h := &Handler{
		name:    fullName,
		handler: handler,
		chMsg:   chMsg,
		logger:  logger,
		pool:    NewWorkerPool(handler.Subject(), chMsg, handler, logger),
	}
	// h.updateLogger(logger)
	return h
}

type Handler struct {
	name string

	handler MsgHandler
	pool    *WorkerPool
	chMsg   chan *nats.Msg

	logger *slog.Logger
}

func (h *Handler) Process(msg *nats.Msg) {
	h.logger.LogAttrs(h.pool.workerCtx, slogd.LevelTrace, "handler received message")
	h.chMsg <- msg
}

func (h *Handler) Handle() func(msg *nats.Msg) {
	return h.handler.Handler()
}

func (h *Handler) Start(ctx context.Context) {
	h.pool.Start(ctx)
}

func (h *Handler) Shutdown() {
	h.pool.Shutdown()
}

func (h *Handler) Subject() string {
	return h.handler.Subject()
}

func (h *Handler) updateLogger(logger *slog.Logger) {
	newLogger := logger.With(slog.String("handler", h.handler.Name()), slog.String("subject", h.handler.Subject()))
	h.logger = newLogger
	h.pool.updateLogger(newLogger)
	h.handler.UpdateLogger(newLogger)
}
