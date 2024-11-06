package slogd

import (
	"log/slog"
	"strings"

	slogformatter "github.com/samber/slog-formatter"
)

const (
	LevelTrace   slog.Level = -8
	LevelFatal   slog.Level = 12
	LevelDebug              = slog.LevelDebug
	LevelInfo               = slog.LevelInfo
	LevelWarn               = slog.LevelWarn
	LevelError              = slog.LevelError
	LevelDefault            = LevelInfo
)

var levelNames = map[slog.Leveler]string{
	LevelTrace: "TRACE",
	LevelDebug: "DEBUG",
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
	LevelFatal: "FATAL",
}

func LevelFormatter(key string) slogformatter.Formatter {
	return slogformatter.FormatByKey(
		key,
		func(v slog.Value) slog.Value {
			return slog.StringValue(LevelName(v.Any().(slog.Level)))
		})
}

func ReplaceAttrs(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.LevelKey {
		levelValue := a.Value.Any().(slog.Level)
		a.Value = slog.StringValue(LevelName(levelValue))
	}
	return a
}

func Level(l string) slog.Leveler {
	mux.Lock()
	defer mux.Unlock()
	for k, v := range levelNames {
		if strings.ToUpper(l) == v {
			return k
		}
	}
	return LevelDefault
}

func LevelName(l slog.Leveler) string {
	mux.Lock()
	defer mux.Unlock()
	for k, v := range levelNames {
		if k == l {
			return v
		}
	}
	return levelNames[LevelDefault]
}
