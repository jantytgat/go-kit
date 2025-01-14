package slogd

import (
	"io"
	"log/slog"
)

func RegisterJSONHandler(w io.Writer, activate bool) {
	RegisterSink(HandlerJSON, slog.NewJSONHandler(w, HandlerOptions()), activate)
}
