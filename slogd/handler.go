package slogd

import (
	"context"
	"io"
	"log/slog"
	"sync"
)

func NewHandler(name string, handler slog.Handler, opts *HandlerOptions) *Handler {
	return &Handler{
		name:           name,
		handler:        handler,
		handlerOptions: opts,
	}
}

func NewDisabledHandler() *Handler {
	return &Handler{
		name:              "disabled",
		handler:           &disabledHandler{},
		handlerOptions:    nil,
		failoverOrder:     0,
		routingPredicates: nil,
	}
}

func NewDefaultJsonHandler(name string, w io.Writer, level slog.Level, addSource bool) *Handler {
	opts := NewDefaultHandlerOptions(level, addSource)
	return &Handler{
		name:              name,
		handler:           slog.NewJSONHandler(w, opts.HandlerOptions()),
		handlerOptions:    opts,
		failoverOrder:     0,
		routingPredicates: nil,
	}
}

func NewDefaultTextHandler(name string, w io.Writer, level slog.Level, addSource bool) *Handler {
	opts := NewDefaultHandlerOptions(level, addSource)
	return &Handler{
		name:              name,
		handler:           slog.NewTextHandler(w, opts.HandlerOptions()),
		handlerOptions:    opts,
		failoverOrder:     0,
		routingPredicates: nil,
	}
}

type Handler struct {
	name              string
	handler           slog.Handler
	handlerOptions    *HandlerOptions
	failoverOrder     int
	routingPredicates []func(ctx context.Context, r slog.Record) bool
	mux               sync.Mutex
}

func (h *Handler) AddRoutingPredicate(predicate func(ctx context.Context, r slog.Record) bool) {
	h.mux.Lock()
	defer h.mux.Unlock()

	h.routingPredicates = append(h.routingPredicates, predicate)
}

func (h *Handler) Handler() slog.Handler {
	h.mux.Lock()
	defer h.mux.Unlock()

	return h.handler
}

func (h *Handler) HandlerOptions() *HandlerOptions {
	h.mux.Lock()
	defer h.mux.Unlock()

	return h.handlerOptions
}

func (h *Handler) Name() string {
	return h.name
}

func (h *Handler) RoutingPredicates() []func(ctx context.Context, r slog.Record) bool {
	h.mux.Lock()
	defer h.mux.Unlock()

	return h.routingPredicates
}

func (h *Handler) GetFailoverOrder() int {
	h.mux.Lock()
	defer h.mux.Unlock()

	return h.failoverOrder
}

func (h *Handler) SetFailoverOrder(failoverOrder int) {
	h.mux.Lock()
	defer h.mux.Unlock()

	h.failoverOrder = failoverOrder
}

func (h *Handler) SetLevel(level slog.Level) {
	h.mux.Lock()
	defer h.mux.Unlock()

	h.handlerOptions.SetLevel(level)
}
