package slogd

import (
	"log/slog"
	"strings"
)

const (
	LevelTrace   = slog.Level(-8)
	LevelDebug   = slog.LevelDebug
	LevelInfo    = slog.LevelInfo
	LevelNotice  = slog.Level(2)
	LevelWarn    = slog.LevelWarn
	LevelError   = slog.LevelError
	LevelFatal   = slog.Level(12)
	LevelDefault = LevelInfo
)

var levelNames = map[slog.Leveler]string{
	LevelTrace:  "TRACE",
	LevelDebug:  "DEBUG",
	LevelInfo:   "INFO",
	LevelNotice: "NOTICE",
	LevelWarn:   "WARN",
	LevelError:  "ERROR",
	LevelFatal:  "FATAL",
}

func GetLevelFromString(l string) slog.Level {
	for k, v := range levelNames {
		if strings.ToUpper(l) == v {
			return k.Level()
		}
	}
	return LevelDefault
}

func GetLevelName(l slog.Level) string {
	for k, v := range levelNames {
		if k == l {
			return v
		}
	}
	return levelNames[LevelDefault]
}

func ReplaceLevelKey(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.LevelKey {
		a.Value = slog.StringValue(GetLevelName(a.Value.Any().(slog.Level)))
	}
	return a
}
