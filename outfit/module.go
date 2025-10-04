package outfit

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/nats-io/nats.go"

	"github.com/jantytgat/go-kit/slogd"
)

type ModuleOption interface {
	Configure(m *Module) error
}

func NewModule(ctx context.Context, name string, logger *slog.Logger, option ...ModuleOption) (*Module, error) {
	m := &Module{
		ctx:      ctx,
		name:     name,
		handlers: nil,
		logger:   logger.With(slog.String("module", name)),
		mux:      sync.Mutex{},
	}

	var err error
	for _, opt := range option {
		if opt == nil {
			continue
		}
		if err = opt.Configure(m); err != nil {
			return nil, err
		}
	}

	return m, nil
}

type Module struct {
	ctx           context.Context
	name          string
	handlers      []*Handler
	connection    *nats.Conn
	subscriptions []*nats.Subscription
	logger        *slog.Logger
	mux           sync.Mutex
}

func (m *Module) Name() string {
	return m.name
}

func (m *Module) AddHandler(handler *Handler) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	for _, h := range m.handlers {
		if h.Subject() == handler.Subject() {
			return errors.New("subject handler already exists in module")
		}
	}
	handler.updateLogger(m.logger)
	m.handlers = append(m.handlers, handler)
	return nil
}

func (m *Module) DeleteHandler(handler *Handler) {
	m.mux.Lock()
	defer m.mux.Unlock()
	for i, h := range m.handlers {
		if h.Subject() == handler.Subject() {
			m.handlers = append(m.handlers[:i], m.handlers[i+1:]...)
		}
	}
}

func (m *Module) Start(ctx context.Context) {
	for _, h := range m.handlers {
		h.Start(ctx)
	}
}

func (m *Module) Shutdown() {
	for _, h := range m.handlers {
		h.Shutdown()
		m.logger.LogAttrs(m.ctx, slogd.LevelInfo, "module shutdown completed", slog.Uint64("messages-in", m.connection.InMsgs), slog.Uint64("messages-out", m.connection.OutMsgs))
	}
}

func (m *Module) Subscribe(nc *nats.Conn) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	var err error
	for _, h := range m.handlers {
		var sub *nats.Subscription
		if h.handler.IsQueueHandler() {
			m.logger.LogAttrs(m.ctx, slogd.LevelTrace, "subscribing nats queue handler", slog.String("handler", h.handler.Name()), slog.String("queue", h.handler.Queue()))
			if sub, err = nc.QueueSubscribe(h.Subject(), h.handler.Queue(), h.Process); err != nil {
				return err
			}
		} else {
			m.logger.LogAttrs(m.ctx, slogd.LevelTrace, "subscribing nats handler", slog.String("handler", h.handler.Name()))
			if sub, err = nc.Subscribe(h.Subject(), h.Process); err != nil {
				return err
			}
		}
		m.subscriptions = append(m.subscriptions, sub)
	}
	m.connection = nc
	return nil
}
