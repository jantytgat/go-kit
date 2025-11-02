package slogd

import (
	"log/slog"
	"sort"
	"sync"

	slogmulti "github.com/samber/slog-multi"
)

const (
	FlowFanOut FlowType = iota
	FlowPipeline
	FlowRouting
	FlowFailOver
	FlowLoadBalancing
)

type FlowType int

func NewFlow(name string, flow FlowType) *Flow {
	return &Flow{
		name:     name,
		handlers: make(map[string]*Handler),
		flow:     flow,
	}
}

type Flow struct {
	name     string
	handlers map[string]*Handler
	flow     FlowType
	mux      sync.Mutex
	logger   *slog.Logger
}

func (l *Flow) build() slog.Handler {
	switch l.flow {
	case FlowFanOut:
		return slogmulti.Fanout(l.getSlogHandlers()...)
	case FlowRouting:
		router := slogmulti.Router()
		for _, h := range l.handlers {
			router.Add(h.handler, h.RoutingPredicates()...)
		}
		return router.Handler()
	case FlowFailOver:
		handlers := l.getHandlers()
		sort.Sort(FailoverHandlerSorter(handlers))
		return slogmulti.Failover()(l.getFailoverSortedSlogHandlers()...)
	case FlowPipeline:
		break
	case FlowLoadBalancing:
		return slogmulti.Pool()(l.getSlogHandlers()...)
	}
	return nil
}

func (l *Flow) getFailoverSortedHandlers() []*Handler {
	handlers := l.getHandlers()
	sort.Sort(FailoverHandlerSorter(handlers))
	return handlers
}

func (l *Flow) getFailoverSortedSlogHandlers() []slog.Handler {
	return convertHandlerToSlogHandler(l.getFailoverSortedHandlers())
}

func (l *Flow) getHandlers() []*Handler {
	handlers := make([]*Handler, 0)

	for _, h := range l.handlers {
		handlers = append(handlers, h)
	}

	return handlers
}

func (l *Flow) getSlogHandlers() []slog.Handler {
	return convertHandlerToSlogHandler(l.getHandlers())
}

func (f *Flow) Logger() *slog.Logger {
	f.mux.Lock()
	defer f.mux.Unlock()

	if f.logger != nil {
		return f.logger
	}

	f.logger = slog.New(f.build())
	return f.logger
}

func (f *Flow) SetLevel(level slog.Level) {
	f.mux.Lock()
	defer f.mux.Unlock()

	for _, h := range f.handlers {
		h.SetLevel(level)
	}

	f.logger = slog.New(f.build())
}

func (l *Flow) WithHandler(name string, handler *Handler) *Flow {
	l.mux.Lock()
	defer l.mux.Unlock()

	l.handlers[name] = handler
	return l
}
