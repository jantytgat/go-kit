package slogd

import (
	"log/slog"
	"sync"
)

func newDefaultLogSet() *LogSet {
	defaultLogger := &LogSet{
		flows: make(map[string]*Flow),
	}

	return defaultLogger.WithDefaultFlow(
		NewFlow("disabled", FlowFanOut).
			WithHandler("disabled", NewDisabledHandler()))
}

type LogSet struct {
	flows       map[string]*Flow
	defaultFlow string
	mux         sync.Mutex
}

func (l *LogSet) Logger(name string) *slog.Logger {
	l.mux.Lock()
	defer l.mux.Unlock()

	if flow, ok := l.flows[name]; ok {
		return flow.Logger()
	}
	return l.flows["disabled"].Logger()
}

func (l *LogSet) WithDefaultFlow(flow *Flow) *LogSet {
	l.mux.Lock()
	defer l.mux.Unlock()

	l.flows[flow.name] = flow
	l.defaultFlow = flow.name

	return l
}
func (l *LogSet) WithFlow(flow *Flow) *LogSet {
	l.mux.Lock()
	defer l.mux.Unlock()

	l.flows[flow.name] = flow
	return l
}
