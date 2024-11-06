package pslog

import (
	"fmt"
	"log/slog"
)

// const (
// 	LevelTrace slog.Level = -8
// 	LevelDebug slog.Level = slog.LevelDebug
// 	LevelInfo  slog.Level = slog.LevelInfo
// 	LevelWarn  slog.Level = slog.LevelWarn
// 	LevelError slog.Level = slog.LevelError
// 	LevelFatal slog.Level = 12
// )

const (
	LevelTrace Level = -8
	LevelDebug Level = -4
	LevelInfo  Level = 0
	LevelWarn  Level = 4
	LevelError Level = 8
	LevelFatal Level = 12
)

type Level slog.Level

func (l Level) Level() slog.Level {
	return slog.Level(l)
}

func (l Level) String() string {
	return levelNames[l]
}

var levelNames = map[slog.Leveler]string{
	LevelTrace: "TRACE",
	LevelDebug: "DEBUG",
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
	LevelFatal: "FATAL",
}

func ReplaceAttrs(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.LevelKey {
		level := a.Value.Any().(slog.Level)
		levelLabel, exists := levelNames[Level(level)]
		if !exists {
			levelLabel = level.String()
		}
		a.Value = slog.StringValue(levelLabel)
	}
	return a
}

func ParseLevelString(level string) (Level, error) {
	switch level {
	case "fatal":
		return LevelFatal, nil
	case "error":
		return LevelError, nil
	case "warn":
		return LevelWarn, nil
	case "info":
		return LevelInfo, nil
	case "debug":
		return LevelDebug, nil
	case "trace":
		return LevelTrace, nil
	default:
		return LevelError, fmt.Errorf("unknown level: %s", level)
	}
}

func ParseLevel(level slog.Leveler) Level {
	return Level(level.Level())
}
