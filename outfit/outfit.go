package outfit

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/nats-io/nats.go"
)

type NatsChanHandler interface {
	Subject(prefix string) string
	Handler() chan *nats.Msg
	NatsWorker
}

type NatsQueueChanHandler interface {
	Queue() string
	NatsChanHandler
}

type NatsWorker interface {
	MaxWorkers() int
	Start(ctx context.Context)
	Shutdown()
	Handle(ctx context.Context, chMsg chan *nats.Msg)
}

type ModuleOption interface {
	Configure(m *Module) error
}

func NewModule(name string, option ...ModuleOption) (*Module, error) {
	m := &Module{
		name:     name,
		handlers: nil,
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
	name     string
	handlers []NatsChanHandler

	mux sync.Mutex
}

func (m *Module) Name() string {
	return m.name
}

func (m *Module) AddNatsChanHandler(prefix string, handler NatsChanHandler) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	for _, h := range m.handlers {
		if h.Subject(prefix) == handler.Subject(prefix) {
			return errors.New("nats channel handler already exists")
		}
	}
	m.handlers = append(m.handlers, handler)
	return nil
}

func (m *Module) RemoveNatsChanHandler(prefix string, handler NatsChanHandler) {
	m.mux.Lock()
	defer m.mux.Unlock()
	for i, h := range m.handlers {
		if h.Subject(prefix) == handler.Subject(prefix) {
			m.handlers = append(m.handlers[:i], m.handlers[i+1:]...)
		}
	}
}

func (m *Module) Subject(prefix string) string {
	if prefix != "" {
		return strings.Join([]string{prefix, m.Name()}, ".")
	}
	return m.Name()
}

func (m *Module) SubscribeAll(prefix string, nc *nats.Conn) ([]*nats.Subscription, error) {
	var err error
	var subs []*nats.Subscription

	for _, h := range m.handlers {
		var sub *nats.Subscription
		switch h.(type) {
		case NatsChanHandler:
			if sub, err = nc.ChanSubscribe(h.Subject(m.Subject(prefix)), h.Handler()); err != nil {
				return nil, err
			}
		case NatsQueueChanHandler:
			if sub, err = nc.ChanQueueSubscribe(h.Subject(m.Subject(prefix)), h.(NatsQueueChanHandler).Queue(), h.Handler()); err != nil {
				return nil, err
			}
		}
		subs = append(subs, sub)
	}
	return subs, nil
}
