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

type ColouredTextHandler struct {
	handler slog.Handler
	w       io.Writer
	mux     *sync.Mutex
}

func (h *ColouredTextHandler) Handle(ctx context.Context, r slog.Record) error {
	var err error
	levelName := levelNames[r.Level]

	switch r.Level {
	case LevelTrace:
		levelName = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF7F50")).Render(levelName) // coral
	case LevelDebug:
		levelName = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FFFF")).Render(levelName) // cyan
	case LevelInfo:
		levelName = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF00")).Render(levelName) // green
	case LevelNotice:
		levelName = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFF00")).Render(levelName) // yellow
	case LevelWarn:
		levelName = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFA500")).Render(levelName) // orange
	case LevelError:
		levelName = lipgloss.NewStyle().Blink(true).Bold(true).Foreground(lipgloss.Color("#FF0000")).Render(levelName) // red
	case LevelFatal:
		levelName = lipgloss.NewStyle().Blink(true).Bold(true).Foreground(lipgloss.Color("#FF00FF")).Render(levelName) // magenta

	}
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
	msg := lipgloss.NewStyle().Foreground(lipgloss.NoColor{}).Render(r.Message) // white

	h.mux.Lock()
	defer h.mux.Unlock()
	_, err = h.w.Write([]byte(strings.Join([]string{timeStr, levelName, msg, lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Render(strings.Join(fieldStrings, " "))}, " ") + "\n"))
	return err
}

func (h *ColouredTextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *ColouredTextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h.handler.WithAttrs(attrs)
}

func (h *ColouredTextHandler) WithGroup(group string) slog.Handler {
	return h.handler.WithGroup(group)
}

func NewColoredTextHandler(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
	return &ColouredTextHandler{
		handler: slog.NewTextHandler(w, opts),
		w:       w,
		mux:     &sync.Mutex{},
	}
}

func RegisterColoredTextHandler(w io.Writer, activate bool) {
	RegisterSink(HandlerColor, NewColoredTextHandler(w, HandlerOptions()), activate)
}
