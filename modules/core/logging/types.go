package logging

import (
	"log/slog"
)

type Level string
type Attr map[string]any

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
	InnerLogger() any
	Clone() LoggerService
	Level() Level
	SetLevel(lvl Level)
	SetContext(key string, value string)
	GetContext(key string) string
	RemoveContext(key string)
	Debug(message string, data Attr)
	Debugf(format string, args ...any)
	Info(message string, data Attr)
	Infof(format string, args ...any)
	Warn(message string, data Attr)
	Warnf(format string, args ...any)
	Error(message string, err error)
	Errorf(format string, args ...any)
}
