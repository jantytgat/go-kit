package slogd

import (
	"context"
	"log/slog"
	"sync"
)

type contextKey struct{}

var (
	defaultCtxKey = contextKey{}
	// formatters    []slogformatter.Formatter
	// middlewares   []slogmulti.Middleware
	logSet *LogSet
	mux    = &sync.Mutex{}
)

func All() *LogSet {
	mux.Lock()
	defer mux.Unlock()

	if logSet != nil {
		return logSet
	}
	logSet = newDefaultLogSet()
	return logSet
}

func convertHandlerToSlogHandler(handlers []*Handler) []slog.Handler {
	slogHandlers := make([]slog.Handler, len(handlers))

	for i, h := range handlers {
		slogHandlers[i] = h.Handler()
	}

	return slogHandlers
}

func FromContext(ctx context.Context) *LogSet {
	if ls, ok := ctx.Value(defaultCtxKey).(*LogSet); ok {
		return ls
	}
	return All()
}

func GetDefaultFlow() *Flow {
	mux.Lock()
	defer mux.Unlock()

	return logSet.flows[logSet.defaultFlow]
}

func GetDefaultLogger() *slog.Logger {
	return GetDefaultFlow().Logger()
}

func GetFlow(name string) *Flow {
	mux.Lock()
	defer mux.Unlock()

	if l, ok := logSet.flows[name]; ok {
		return l
	}
	return logSet.flows[logSet.defaultFlow]
}

func GetLogger(name string) *slog.Logger {
	return GetFlow(name).Logger()
}

func SetLevel(flow string, level slog.Level) {
	GetFlow(flow).SetLevel(level)
}

func WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, defaultCtxKey, All())
}

//
// func RegisterFormatter(f slogformatter.Formatter) {
// 	mux.Lock()
// 	defer mux.Unlock()
// 	formatters = append(formatters, f)
// }
//
// func RegisterMiddleware(h slogmulti.Middleware) {
// 	mux.Lock()
// 	defer mux.Unlock()
// 	middlewares = append(middlewares, h)
// }
//

//
// func UseHandler(name string) {
// 	mux.Lock()
// 	defer mux.Unlock()
// 	if _, ok := handlers[name]; !ok {
// 		Flow().LogAttrs(context.Background(), LevelError, "could not find handler", slog.String("name", name))
// 		return
// 	}
//
// 	formatterPipe := slogformatter.NewFormatterMiddleware(formatters...)
// 	pipe := slogmulti.Pipe(middlewares...).Pipe(formatterPipe)
// 	handler := slogmulti.Fanout(handlers[name])
//
// 	logSet = slog.NewFlow(pipe.Handler(handler))
// 	activeHandler = name
// }
