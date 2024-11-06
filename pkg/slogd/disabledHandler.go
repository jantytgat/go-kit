package slogd

import (
	"context"
	"log/slog"
)

func newDisabledHandler() slog.Handler {
	return &disabledHandler{}
}

func registerDisabledHandler(activate bool) {
	RegisterHandler(handlerDisabled, newDisabledHandler(), activate)
}

type disabledHandler struct{}

func (h *disabledHandler) Handle(ctx context.Context, r slog.Record) error {
	return nil
}

func (h *disabledHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return false
}

func (h *disabledHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *disabledHandler) WithGroup(group string) slog.Handler {
	return h
}
