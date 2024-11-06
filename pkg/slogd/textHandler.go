package slogd

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
)

type TextHandler struct {
	handler slog.Handler
	w       io.Writer
	mux     *sync.Mutex
}

func (h *TextHandler) Handle(ctx context.Context, r slog.Record) error {
	var err error
	levelName := levelNames[r.Level]
	fields := make(map[string]interface{}, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()

		return true
	})

	var fieldStrings []string
	for k, v := range fields {
		fieldStrings = append(fieldStrings, k+"="+fmt.Sprintf("%v", v))
	}

	timeStr := r.Time.Format("[15:05:05.0000]")

	h.mux.Lock()
	defer h.mux.Unlock()
	_, err = h.w.Write([]byte(strings.Join([]string{timeStr, levelName, r.Message, lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Render(strings.Join(fieldStrings, " "))}, " ") + "\n"))
	return err
}

func (h *TextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *TextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h.handler.WithAttrs(attrs)
}

func (h *TextHandler) WithGroup(group string) slog.Handler {
	return h.handler.WithGroup(group)
}

func NewTextHandler(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
	return &TextHandler{
		handler: slog.NewTextHandler(w, opts),
		w:       w,
		mux:     &sync.Mutex{},
	}
}

func RegisterTextHandler(w io.Writer, activate bool) {
	RegisterHandler(HandlerText, NewTextHandler(w, HandlerOptions()), activate)
}
