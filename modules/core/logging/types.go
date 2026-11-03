package logging

import (
	"context"
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

// LoggerService takes ctx on every write so each record carries the span, idempotency id,
// tenant and org active at that point, not those of the place the logger was obtained.
// Code with no operation (boot, shutdown, CLI) passes context.Background(): the fields are null.
type LoggerService interface {
	InnerLogger() any
	Clone() LoggerService
	Level() Level
	SetLevel(lvl Level)
	Debug(ctx context.Context, message string, data Attr)
	Debugf(ctx context.Context, format string, args ...any)
	Info(ctx context.Context, message string, data Attr)
	Infof(ctx context.Context, format string, args ...any)
	Warn(ctx context.Context, message string, data Attr)
	Warnf(ctx context.Context, format string, args ...any)
	Error(ctx context.Context, message string, err error)
	Errorf(ctx context.Context, format string, args ...any)
}
