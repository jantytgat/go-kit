package slogd

import (
	"context"
	"log/slog"
	"sync"

	slogformatter "github.com/samber/slog-formatter"
	slogmulti "github.com/samber/slog-multi"
)

const (
	HandlerText     string = "text"
	HandlerJSON     string = "json"
	handlerDisabled string = "disabled"
)

const (
	FlowFanOut Flow = iota
	FlowPipeline
	FlowRouting
	FlowFailOver
	FlowLoadBalancing
)

type Flow int

var (
	ctxKey = contextKey{}
)
var (
	handlers      = make(map[string]slog.Handler)
	activeHandler string
	level         = new(slog.LevelVar)
	formatters    []slogformatter.Formatter
	middlewares   []slogmulti.Middleware
	source        bool
	logger        *slog.Logger
	mux           = &sync.Mutex{}
)

func ActiveHandler() string {
	mux.Lock()
	defer mux.Unlock()
	return activeHandler
}

func Key() contextKey {
	return ctxKey
}

func Disable() {
	UseHandler(handlerDisabled)
}

func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(Key()).(*slog.Logger); ok {
		return l
	}
	return Logger()
}

func HandlerOptions() *slog.HandlerOptions {
	return &slog.HandlerOptions{
		AddSource:   source,
		Level:       level,
		ReplaceAttr: ReplaceAttrs,
	}
}

func Init(l slog.Level, addSource bool) {
	level.Set(l)
	source = addSource
}

func Logger() *slog.Logger {
	mux.Lock()
	defer mux.Unlock()

	if logger == nil {
		logger = slog.New(handlers[handlerDisabled])
	}
	return logger
}

func SetLevel(l slog.Level) {
	mux.Lock()
	defer mux.Unlock()
	level.Set(l)
}

func RegisterFormatter(f slogformatter.Formatter) {
	mux.Lock()
	defer mux.Unlock()
	formatters = append(formatters, f)
}

func RegisterMiddleware(h slogmulti.Middleware) {
	mux.Lock()
	defer mux.Unlock()
	middlewares = append(middlewares, h)
}

func RegisterSink(name string, h slog.Handler, activate bool) {
	mux.Lock()
	handlers[name] = h
	mux.Unlock()

	if activate {
		UseHandler(name)
	}
}

func UseHandler(name string) {
	mux.Lock()
	defer mux.Unlock()
	if _, ok := handlers[name]; !ok {
		Logger().LogAttrs(context.Background(), LevelError, "could not find handler", slog.String("name", name))
		return
	}

	formatterPipe := slogformatter.NewFormatterMiddleware(formatters...)
	pipe := slogmulti.Pipe(middlewares...).Pipe(formatterPipe)
	handler := slogmulti.Fanout(handlers[name])

	logger = slog.New(pipe.Handler(handler))
	activeHandler = name
}

func WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, Key(), Logger())
}

func init() {
	// RegisterFormatter(LevelFormatter())
	// RegisterMiddleware(NewLevelMiddleware())
	registerDisabledHandler(true)
}

type contextKey struct{}
