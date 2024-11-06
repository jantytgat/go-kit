package pslog

import (
	"io"
	"log/slog"
)

func New(w io.Writer, l Level, f Format) *slog.Logger {
	switch f {
	case TextLogFormat:
		return slog.New(slog.NewTextHandler(
			w,
			&slog.HandlerOptions{
				AddSource:   false,
				Level:       l,
				ReplaceAttr: ReplaceAttrs,
			}))
	case JsonLogFormat:
		return slog.New(slog.NewJSONHandler(
			w,
			&slog.HandlerOptions{
				AddSource:   false,
				Level:       l,
				ReplaceAttr: ReplaceAttrs,
			}))
	case ColouredTextLogFormat:
		return slog.New(NewColouredTextHandler(
			w,
			&slog.HandlerOptions{
				AddSource:   false,
				Level:       l,
				ReplaceAttr: ReplaceAttrs,
			}))
	default:
		return slog.New(slog.NewTextHandler(
			w,
			&slog.HandlerOptions{
				AddSource:   false,
				Level:       l,
				ReplaceAttr: ReplaceAttrs,
			}))
	}
}
