package pslog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
)

type PrettyHandlerOptions struct {
	SlogOpts slog.HandlerOptions
}

type ColouredTextHandler struct {
	slog.Handler
	l   io.Writer
	mux *sync.Mutex
}

func (h *ColouredTextHandler) Handle(ctx context.Context, r slog.Record) error {
	var err error
	level := Level(r.Level).String()

	switch Level(r.Level) {
	case LevelTrace:
		level = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF00FF")).Render(level) // magenta
	case LevelDebug:
		level = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FFFF")).Render(level) // cyan
	case LevelInfo:
		level = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF00")).Render(level) // green
	case LevelWarn:
		level = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFF00")).Render(level) // yellow
	case LevelError:
		level = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF0000")).Render(level) // red
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
	_, err = h.l.Write([]byte(strings.Join([]string{timeStr, level, msg, lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Render(strings.Join(fieldStrings, " "))}, " ") + "\n"))
	return err
}

func NewColouredTextHandler(out io.Writer, opts *slog.HandlerOptions) *ColouredTextHandler {
	return &ColouredTextHandler{
		Handler: slog.NewTextHandler(out, opts),
		l:       out,
		mux:     &sync.Mutex{},
	}
}
