package logging

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"
	"runtime"

	"github.com/sky-as-code/nikki-erp/common/env"
	"go.bryk.io/pkg/errors"
)

const skippedCallStackDepth = 3

func NewLogger(levels ...Level) *loggerImpl {
	level := LevelInfo // default level
	if len(levels) > 0 {
		level = levels[0]
	}

	levelVar := new(slog.LevelVar)
	levelVar.Set(slogLevelMap[level])
	slogger := NewSloggerWithCorrectCallDepth(levelVar, 6)
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

func (this *loggerImpl) Debug(message string, data Attr) {
	this.writeLogData(LevelDebug, message, data)
}

func (this *loggerImpl) Debugf(format string, args ...any) {
	this.writeLogFormat(LevelDebug, format, args...)
}

func (this *loggerImpl) Info(message string, data Attr) {
	this.writeLogData(LevelInfo, message, data)
}

func (this *loggerImpl) Infof(format string, args ...any) {
	this.writeLogFormat(LevelInfo, format, args...)
}

func (this *loggerImpl) Warn(message string, data Attr) {
	this.writeLogData(LevelWarn, message, data)
}

func (this *loggerImpl) Warnf(format string, args ...any) {
	this.writeLogFormat(LevelWarn, format, args...)
}

func (this *loggerImpl) Error(message string, err error) {
	this.writeErrorData(message, err)
}

func (this *loggerImpl) Errorf(format string, args ...any) {
	this.writeLogFormat(LevelError, format, args...)
}

func (this *loggerImpl) writeLogData(level Level, message string, data Attr) {
	fileNameLine := getFileName(skippedCallStackDepth)
	this.slogger.Log(context.Background(), slogLevelMap[level], message, slog.Any("data", data), slog.String("source", fileNameLine))
}

func (this *loggerImpl) writeLogFormat(level Level, format string, args ...any) {
	fileNameLine := getFileName(skippedCallStackDepth)
	this.slogger.Log(context.Background(), slogLevelMap[level], fmt.Sprintf(format, args...), slog.String("source", fileNameLine))
}

func (this *loggerImpl) writeErrorData(message string, err error) {
	var typedErr *errors.Error
	var data any
	isLocal := env.IsLocal()
	codec := NewJsonCodec(isLocal)

	if errors.As(err, &typedErr) {
		js, _ := errors.Report(typedErr, codec)
		data = string(js)
	} else {
		data = err
	}

	// Copy writeLogData implementation here so the stackDepth is correct
	fileNameLine := getFileName(skippedCallStackDepth)
	this.slogger.Log(context.Background(), slog.LevelError, message, slog.Any("data", data), slog.String("source", fileNameLine))

	if isLocal {
		printStackTrace(err)
	}
}

func getFileName(depth int) string {
	pc, file, line, _ := runtime.Caller(depth)
	fn := runtime.FuncForPC(pc)
	if fn != nil {
		return fmt.Sprintf("%s:%d %s", file, line, path.Base(fn.Name()))
	}
	return fmt.Sprintf("%s:%d", file, line)
}

func NewSloggerWithCorrectCallDepth(level slog.Leveler, callDepth int) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: false, // we’ll manually inject source
	})

	return slog.New(handler)
}

func printStackTrace(err error) {
	rec := NewErrReport(err, false)
	printRed("Stacktrace:", "")
	for i, frame := range rec.Stacktrace {
		// Print in red color
		printRed(frame, fmt.Sprintf("[%d] ", i))
	}
}

func printRed(message string, prefix string) {
	fmt.Printf("\033[31m%s%v\033[0m\n", prefix, message)
}
