package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"runtime"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/env"
)

// skippedCallStackDepth reaches the caller of a public write method:
// getFileName <- write* <- Debug/Info/... <- caller.
const skippedCallStackDepth = 3

func NewLogger(levels ...Level) LoggerService {
	return NewLoggerTo(os.Stdout, levels...)
}

// NewLoggerTo writes to out instead of stdout; tests read the JSON lines back from it.
func NewLoggerTo(out io.Writer, levels ...Level) LoggerService {
	level := LevelInfo
	if len(levels) > 0 {
		level = levels[0]
	}
	levelVar := new(slog.LevelVar)
	levelVar.Set(slogLevelMap[level])
	return &loggerImpl{
		level:    level,
		levelVar: levelVar,
		slogger:  slog.New(NewTelemetryHandler(out, levelVar)),
	}
}

type loggerImpl struct {
	level    Level
	levelVar *slog.LevelVar
	slogger  *slog.Logger
}

func (this *loggerImpl) InnerLogger() any {
	return this.slogger
}

func (this *loggerImpl) Clone() LoggerService {
	return &loggerImpl{
		level:    this.level,
		levelVar: this.levelVar,
		slogger:  this.slogger,
	}
}

func (this *loggerImpl) Level() Level {
	return this.level
}

func (this *loggerImpl) SetLevel(lvl Level) {
	this.level = lvl
	this.levelVar.Set(slogLevelMap[lvl])
	this.Infof(context.Background(), "Log level is set to %s", lvl)
}

func (this *loggerImpl) Debug(ctx context.Context, message string, data Attr) {
	this.writeLogData(ctx, LevelDebug, message, data)
}

func (this *loggerImpl) Debugf(ctx context.Context, format string, args ...any) {
	this.writeLogFormat(ctx, LevelDebug, format, args...)
}

func (this *loggerImpl) Info(ctx context.Context, message string, data Attr) {
	this.writeLogData(ctx, LevelInfo, message, data)
}

func (this *loggerImpl) Infof(ctx context.Context, format string, args ...any) {
	this.writeLogFormat(ctx, LevelInfo, format, args...)
}

func (this *loggerImpl) Warn(ctx context.Context, message string, data Attr) {
	this.writeLogData(ctx, LevelWarn, message, data)
}

func (this *loggerImpl) Warnf(ctx context.Context, format string, args ...any) {
	this.writeLogFormat(ctx, LevelWarn, format, args...)
}

func (this *loggerImpl) Error(ctx context.Context, message string, err error) {
	this.writeErrorData(ctx, message, err)
}

func (this *loggerImpl) Errorf(ctx context.Context, format string, args ...any) {
	this.writeLogFormat(ctx, LevelError, format, args...)
}

func (this *loggerImpl) writeLogData(ctx context.Context, level Level, message string, data Attr) {
	if !this.slogger.Enabled(ctx, slogLevelMap[level]) {
		return
	}
	fileNameLine := getFileName(skippedCallStackDepth)
	attrs := []slog.Attr{slog.String(attrKeySource, fileNameLine)}
	if data != nil {
		attrs = append(attrs, slog.Any(attrKeyData, data))
	}
	this.slogger.LogAttrs(ctx, slogLevelMap[level], message, attrs...)
}

func (this *loggerImpl) writeLogFormat(ctx context.Context, level Level, format string, args ...any) {
	if !this.slogger.Enabled(ctx, slogLevelMap[level]) {
		return
	}
	fileNameLine := getFileName(skippedCallStackDepth)
	this.slogger.LogAttrs(ctx, slogLevelMap[level], fmt.Sprintf(format, args...), slog.String(attrKeySource, fileNameLine))
}

func (this *loggerImpl) writeErrorData(ctx context.Context, message string, err error) {
	if !this.slogger.Enabled(ctx, slog.LevelError) {
		return
	}
	isLocal := env.IsLocal()
	fileNameLine := getFileName(skippedCallStackDepth)
	attrs := []slog.Attr{slog.String(attrKeySource, fileNameLine)}
	if err != nil {
		attrs = append(attrs,
			slog.String(attrKeyErrorType, errorTypeOf(err)),
			slog.Any(attrKeyData, NewErrReport(err, isLocal)),
		)
	}
	this.slogger.LogAttrs(ctx, slog.LevelError, message, attrs...)

	if isLocal && err != nil {
		printStackTrace(err)
	}
}

// errorTypeOf groups errors by the Go type of the innermost cause, a bounded set, rather than by
// the message, which would make every distinct failure its own group.
func errorTypeOf(err error) string {
	cause := err
	for {
		next := errors.Unwrap(cause)
		if next == nil {
			break
		}
		cause = next
	}
	return fmt.Sprintf("%T", cause)
}

func getFileName(depth int) string {
	pc, file, line, _ := runtime.Caller(depth)
	fn := runtime.FuncForPC(pc)
	if fn != nil {
		return fmt.Sprintf("%s:%d %s", file, line, path.Base(fn.Name()))
	}
	return fmt.Sprintf("%s:%d", file, line)
}

func printStackTrace(err error) {
	rec := NewErrReport(err, false)
	printRed("Stacktrace:", "")
	for i, frame := range rec.Stacktrace {
		printRed(frame, fmt.Sprintf("[%d] ", i))
	}
}

// printRed writes the local-env stack trace for a human at a terminal, not a log record.
func printRed(message string, prefix string) {
	fmt.Printf("\033[31m%s%v\033[0m\n", prefix, message)
}
