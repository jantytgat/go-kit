package slogd

import (
	"log/slog"
	"sync"
)

func NewCustomHandlerOptions(level slog.Level, addSource bool, fn []func(groups []string, a slog.Attr) slog.Attr) *HandlerOptions {
	levelVar := new(slog.LevelVar)
	levelVar.Set(level)

	return &HandlerOptions{
		addSource:    addSource,
		levelVar:     levelVar,
		replaceAttrs: fn,
	}
}

func NewDefaultHandlerOptions(level slog.Level, addSource bool) *HandlerOptions {
	levelVar := new(slog.LevelVar)
	levelVar.Set(level)

	var replaceAttrsFuncs []func(groups []string, a slog.Attr) slog.Attr
	replaceAttrsFuncs = append(replaceAttrsFuncs, ReplaceLevelKey)

	return &HandlerOptions{
		addSource:    addSource,
		levelVar:     levelVar,
		replaceAttrs: replaceAttrsFuncs,
	}
}

type HandlerOptions struct {
	levelVar     *slog.LevelVar
	addSource    bool
	replaceAttrs []func(groups []string, a slog.Attr) slog.Attr
	mux          sync.Mutex
}

func (h *HandlerOptions) AddReplaceAttrsFunc(fn func(groups []string, a slog.Attr) slog.Attr) {
	h.mux.Lock()
	defer h.mux.Unlock()

	h.replaceAttrs = append(h.replaceAttrs, fn)
}

func (h *HandlerOptions) HandlerOptions() *slog.HandlerOptions {
	h.mux.Lock()
	defer h.mux.Unlock()

	return &slog.HandlerOptions{
		AddSource:   h.addSource,
		Level:       h.levelVar,
		ReplaceAttr: replaceAttrsFunc(h.replaceAttrs),
	}
}

func (h *HandlerOptions) SetLevel(level slog.Level) {
	h.mux.Lock()
	defer h.mux.Unlock()

	h.levelVar.Set(level)
}

func replaceAttrsFunc(fs []func(groups []string, a slog.Attr) slog.Attr) func(groups []string, a slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		var attr slog.Attr
		for _, f := range fs {
			attr = f(groups, a)
		}
		return attr
	}
}
