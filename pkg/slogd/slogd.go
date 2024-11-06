package slogd

import (
	"context"
	"io"
	"log/slog"
	"sync"

	slogformatter "github.com/samber/slog-formatter"
	slogmulti "github.com/samber/slog-multi"
)

const (
	HandlerText     string = "text"
	HandlerJSON     string = "json"
	HandlerColor    string = "color"
	handlerDisabled string = "disabled"
)

var (
	ctxKey = contextKey{}
)
var (
	handlers      = make(map[string]slog.Handler)
	activeHandler string
	level         = new(slog.LevelVar)
	formatters    = make([]slogformatter.Formatter, 0)
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
		AddSource: source,
		Level:     level,
		// ReplaceAttr: ReplaceAttrs,
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
	level.Set(l.Level())
}

func RegisterFormatter(f slogformatter.Formatter) {
	mux.Lock()
	defer mux.Unlock()

	formatters = append(formatters, f)
}

func RegisterHandler(name string, h slog.Handler, activate bool) {
	mux.Lock()
	defer mux.Unlock()

	handlers[name] = h

	if activate {
		logger = slog.New(slogmulti.
			Pipe(slogformatter.NewFormatterMiddleware(formatters...)).
			Handler(h))
		activeHandler = name
	}
}

func RegisterJSONHandler(w io.Writer, activate bool) {
	RegisterHandler(HandlerJSON, slog.NewJSONHandler(w, HandlerOptions()), activate)
}

func UseHandler(name string) {
	mux.Lock()
	defer mux.Unlock()
	if _, ok := handlers[name]; !ok {
		slog.LogAttrs(context.Background(), LevelError.Level(), "could not find handler", slog.String("name", name))
		return
	}

	logger = slog.New(slogmulti.
		Pipe(slogformatter.NewFormatterMiddleware(formatters...)).
		Handler(slogmulti.Fanout(handlers[name])))
	activeHandler = name
}

func WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, Key(), Logger())
}

func init() {
	RegisterFormatter(LevelFormatter(slog.LevelKey))
	registerDisabledHandler(true)
}

type contextKey struct{}
