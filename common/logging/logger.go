package logging

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
)

func NewLogger(levels ...Level) *loggerImpl {
	level := LevelInfo // default level
	if len(levels) > 0 {
		level = levels[0]
	}

	levelVar := new(slog.LevelVar)
	levelVar.Set(slogLevelMap[level])
	slogger := NewSloggerWithCorrectCallDepth(levelVar, 1)
	return &loggerImpl{
		level:    level,
		levelVar: levelVar,
		slogger:  slogger,
	}
}

type loggerImpl struct {
	level    Level
	levelVar *slog.LevelVar
	slogger  *slog.Logger
}

func (this *loggerImpl) Level() Level {
	return this.level
}

func (this *loggerImpl) SetLevel(lvl Level) {
	this.levelVar.Set(slogLevelMap[lvl])
}

func (this *loggerImpl) InnerLogger() any {
	return this.slogger
}

func (this *loggerImpl) Debug(message string, data any) {
	this.writeLogData(LevelDebug, message, data)
}

func (this *loggerImpl) Debugf(format string, args ...any) {
	this.writeLogFormat(LevelDebug, format, args...)
}

func (this *loggerImpl) Info(message string, data any) {
	this.writeLogData(LevelInfo, message, data)
}

func (this *loggerImpl) Infof(format string, args ...any) {
	this.writeLogFormat(LevelInfo, format, args...)
}

func (this *loggerImpl) Warn(message string, data any) {
	this.writeLogData(LevelWarn, message, data)
}

func (this *loggerImpl) Warnf(format string, args ...any) {
	this.writeLogFormat(LevelWarn, format, args...)
}

func (this *loggerImpl) Error(message string, data any) {
	this.writeLogData(LevelError, message, data)
}

func (this *loggerImpl) Errorf(format string, args ...any) {
	this.writeLogFormat(LevelError, format, args...)
}

func (this *loggerImpl) IfError(err error, message string, data any) {
	if err != nil {
		this.Error(message, data)
	}
}

func (this *loggerImpl) IfErrorf(err error, format string, args ...any) {
	if err != nil {
		this.Errorf(format, args...)
	}
}

func (this *loggerImpl) writeLogData(level Level, message string, data any) {
	this.slogger.Log(context.Background(), slogLevelMap[level], message, slog.Any("data", data))
}

func (this *loggerImpl) writeLogFormat(level Level, format string, args ...any) {
	this.slogger.Log(context.Background(), slogLevelMap[level], format, args...)
}

func NewSloggerWithCorrectCallDepth(level slog.Leveler, callDepth int) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: false, // we’ll manually inject source
	})

	return slog.New(&HandlerWithCorrectCallDepth{
		inner:     handler,
		callDepth: callDepth,
	})
}

type HandlerWithCorrectCallDepth struct {
	inner     slog.Handler
	callDepth int
}

func (this *HandlerWithCorrectCallDepth) Enabled(ctx context.Context, level slog.Level) bool {
	return this.inner.Enabled(ctx, level)
}

func (this *HandlerWithCorrectCallDepth) Handle(ctx context.Context, r slog.Record) error {
	// Copy record and inject the real source location
	pcs := make([]uintptr, 1)
	runtime.Callers(this.callDepth, pcs)
	frames := runtime.CallersFrames(pcs)
	frame, _ := frames.Next()

	r = slog.NewRecord(r.Time, r.Level, r.Message, 0)
	r.AddAttrs(slog.String("source", fmt.Sprintf("%s:%d", frame.File, frame.Line)))

	return this.inner.Handle(ctx, r)
}

func (this *HandlerWithCorrectCallDepth) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &HandlerWithCorrectCallDepth{
		inner:     this.inner.WithAttrs(attrs),
		callDepth: this.callDepth,
	}
}

func (this *HandlerWithCorrectCallDepth) WithGroup(name string) slog.Handler {
	return &HandlerWithCorrectCallDepth{
		inner:     this.inner.WithGroup(name),
		callDepth: this.callDepth,
	}
}

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}
