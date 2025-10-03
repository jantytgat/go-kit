package outfit

import (
	"errors"
	"strings"
	"sync"

	"github.com/nats-io/nats.go"
)

type Component struct {
	Name    string
	Modules []Module
}

type Module struct {
	Name         string
	httpHandlers []HttpHandler
	natsHandlers []NatsHandler

	mux sync.Mutex
}

func (m *Module) getFullSubject(prefix string) string {
	if prefix != "" {
		return strings.Join([]string{prefix, m.Name}, ".")
	}
	return m.Name
}

func (m *Module) AddHandler(httpHandler HttpHandler, natsHandler NatsHandler) {
	m.mux.Lock()
	defer m.mux.Unlock()

	m.httpHandlers = append(m.httpHandlers, httpHandler)
	m.natsHandlers = append(m.natsHandlers, natsHandler)
}

func (m *Module) RegisterAllNatsHandlers(prefix string, nc *nats.Conn) ([]*nats.Subscription, error) {
	m.mux.Lock()
	defer m.mux.Unlock()

	var err error
	var subs []*nats.Subscription

	for _, h := range m.natsHandlers {
		var sub *nats.Subscription
		if sub, err = h.Register(m.getFullSubject(prefix), nc); err != nil {
			return subs, err
		}
		subs = append(subs, sub)
	}
	return subs, nil
}

func (m *Module) RegisterNatsHandler(prefix, subject string, nc *nats.Conn) (*nats.Subscription, error) {
	m.mux.Lock()
	defer m.mux.Unlock()

	for _, h := range m.natsHandlers {
		if h.Subject == subject {
			return h.Register(m.getFullSubject(prefix), nc)
		}
	}
	return nil, errors.New("nats handler not found")
}

type HttpHandler struct {
	Resource string
}

type SubscriptionType int

const (
	InvalidSubscription SubscriptionType = iota
	AsyncSubscription
	QueueSubscription
)

type NatsHandler struct {
	Subject string
	Queue   string
	Type    SubscriptionType
	Handler func(msg *nats.Msg)
}

// getFullSubject Get Fully Subject Name
func (h *NatsHandler) getFullSubject(prefix string) string {
	if prefix != "" {
		return strings.Join([]string{prefix, h.Subject}, ".")
	}
	return h.Subject
}

func (h *NatsHandler) Register(prefix string, nc *nats.Conn) (*nats.Subscription, error) {

	switch h.Type {
	case AsyncSubscription:
		return nc.Subscribe(h.getFullSubject(prefix), h.Handler)
	case QueueSubscription:
		return nc.QueueSubscribe(h.getFullSubject(prefix), h.Queue, h.Handler)
	default:
		return nil, errors.New("invalid subscription type")
	}
}
