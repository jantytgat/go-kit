package outfit

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/jantytgat/go-kit/slogd"
)

func NewHelloHandler(ctx context.Context, subject, queue string, maxWorkers int, logger *slog.Logger) *HelloHandler {
	return &HelloHandler{
		ctx:        ctx,
		name:       "hello",
		subject:    subject,
		queue:      queue,
		maxWorkers: maxWorkers,
		logger:     logger,
	}
}

type HelloHandler struct {
	ctx        context.Context
	name       string
	maxWorkers int
	queue      string
	subject    string
	logger     *slog.Logger
}

func (h *HelloHandler) Handler() func(msg *nats.Msg) {
	return h.Handle
}

func (h *HelloHandler) IsQueueHandler() bool {
	return h.queue != ""
}

func (h *HelloHandler) MaxWorkers() int {
	return h.maxWorkers
}

func (h *HelloHandler) Name() string {
	return h.name
}

func (h *HelloHandler) Queue() string {
	return h.queue
}

func (h *HelloHandler) Subject() string {
	return h.subject
}

func (h *HelloHandler) UpdateLogger(logger *slog.Logger) {
	h.logger = logger
}

func (h *HelloHandler) Handle(msg *nats.Msg) {
	h.logger.LogAttrs(h.ctx, slogd.LevelTrace, "handling message")
	fmt.Println("Hello", string(msg.Data))
	time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
}
