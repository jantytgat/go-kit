package slogd

import (
	"io"
	"log/slog"
)

func RegisterTextHandler(w io.Writer, activate bool) {
	RegisterSink(HandlerText, slog.NewTextHandler(w, HandlerOptions()), activate)
}
