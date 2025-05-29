package logging

import (
	"log/slog"
)

type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

var levelNameMap = map[string]Level{
	"debug": LevelDebug,
	"info":  LevelInfo,
	"warn":  LevelWarn,
	"error": LevelError,
}

var slogLevelMap = map[Level]slog.Level{
	LevelDebug: slog.LevelDebug,
	LevelInfo:  slog.LevelInfo,
	LevelWarn:  slog.LevelWarn,
	LevelError: slog.LevelError,
}

type LoggerService interface {
	Level() Level
	SetLevel(lvl Level)
	InnerLogger() any
	Debug(message string, data any)
	Debugf(format string, args ...any)
	Info(message string, data any)
	Infof(format string, args ...any)
	Warn(message string, data any)
	Warnf(format string, args ...any)
	Error(message string, err error)
	Errorf(format string, args ...any)
}
