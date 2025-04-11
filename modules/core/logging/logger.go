package logging

import (
	"context"
	"fmt"
	"log"
	"path"
	"runtime"

	. "github.com/sky-as-code/nikki-erp/modules/shared"
	"github.com/sky-as-code/nikki-erp/utility/json"
)

const (
	FORMAT_NONE = ""
	FORMAT_JSON = "json"
)

func NewLogger(level Lvl) *loggerImpl {
	return &loggerImpl{level: level}
}

type loggerImpl struct {
	level     Lvl
	requestId string
}

func (this *loggerImpl) Level() Lvl {
	return this.level
}

func (this *loggerImpl) SetLevel(level Lvl) {
	this.level = level
}

func (this *loggerImpl) Debug(i ...any) {
	if this.requestId != "" {
		i = append([]any{fmt.Sprintf("requestId: %s ", this.requestId)}, i...)
	}
	this.writeLog(LVL_DEBUG, FORMAT_NONE, i...)
}

func (this *loggerImpl) Debugf(format string, args ...any) {
	if this.requestId != "" {
		format = fmt.Sprintf("requestId: %s %s", this.requestId, format)
	}
	this.writeLog(LVL_DEBUG, format, args...)
}

func (this *loggerImpl) Debugj(j JSON) {
	this.writeLog(LVL_DEBUG, FORMAT_JSON, j)
}

func (this *loggerImpl) Info(i ...any) {
	if this.requestId != "" {
		i = append([]any{fmt.Sprintf("requestId: %s ", this.requestId)}, i...)
	}
	this.writeLog(LVL_INFO, FORMAT_NONE, i...)
}

func (this *loggerImpl) Infof(format string, args ...any) {
	if this.requestId != "" {
		format = fmt.Sprintf("requestId: %s %s", this.requestId, format)
	}
	this.writeLog(LVL_INFO, format, args...)
}

func (this *loggerImpl) Infoj(j JSON) {
	this.writeLog(LVL_INFO, FORMAT_JSON, j)
}

func (this *loggerImpl) Warn(i ...any) {
	if this.requestId != "" {
		i = append([]any{fmt.Sprintf("requestId: %s ", this.requestId)}, i...)
	}
	this.writeLog(LVL_WARN, FORMAT_NONE, i...)
}

func (this *loggerImpl) Warnf(format string, args ...any) {
	if this.requestId != "" {
		format = fmt.Sprintf("requestId: %s %s", this.requestId, format)
	}
	this.writeLog(LVL_WARN, format, args...)
}

func (this *loggerImpl) Warnj(j JSON) {
	this.writeLog(LVL_WARN, FORMAT_JSON, j)
}

func (this *loggerImpl) Error(i ...any) {
	if this.requestId != "" {
		i = append([]any{fmt.Sprintf("requestId: %s ", this.requestId)}, i...)
	}
	this.writeLog(LVL_ERROR, FORMAT_NONE, i...)
}

func (this *loggerImpl) Errorf(format string, args ...any) {
	if this.requestId != "" {
		format = fmt.Sprintf("requestId: %s %s", this.requestId, format)
	}
	this.writeLog(LVL_ERROR, format, args...)
}

func (this *loggerImpl) Errorj(j JSON) {
	this.writeLog(LVL_ERROR, FORMAT_JSON, j)
}

func (this *loggerImpl) IfError(err error, i ...any) {
	if err != nil {
		this.writeLog(LVL_ERROR, FORMAT_NONE, i...)
	}
}

func (this *loggerImpl) IfErrorf(err error, format string, args ...any) {
	if err != nil {
		this.writeLog(LVL_ERROR, format, args...)
	}
}

func (this *loggerImpl) Fatal(i ...any) {
	this.writeLog(LVL_FATAL, FORMAT_NONE, i...)
}

func (this *loggerImpl) Fatalj(j JSON) {
	this.writeLog(LVL_FATAL, FORMAT_JSON, j)
}

func (this *loggerImpl) Fatalf(format string, args ...any) {
	this.writeLog(LVL_FATAL, format, args...)
}

func (this *loggerImpl) WithRequestId(requestId string) {
	this.requestId = requestId
}

func (this *loggerImpl) writeLog(level Lvl, format string, args ...any) {
	if level > this.Level() {
		return
	}

	message := ""

	switch format {
	case FORMAT_NONE:
		message = fmt.Sprint(args...)
	case FORMAT_JSON:
		b, err := json.Marshal(args[0])
		if err != nil {
			panic(err)
		}
		message = string(b)
	default:
		message = fmt.Sprintf(format, args...)
	}

	// If some day you see all the logs with "logger.go(<line number>)",
	// try increasing this value until you see the correct file name :)
	skippedCallStackDepth := 3
	log.Println(LvlName[level], getFileName(skippedCallStackDepth), message)
}

func getFileName(depth int) string {
	_, file, line, _ := runtime.Caller(depth)
	return fmt.Sprintf("%s(%d)", path.Base(file), line)
}

var loggerKey = "logger"

// IntoContext return a new context with the logger injected
func IntoContext(ctx context.Context, logger LoggerService) {
	ctx = context.WithValue(ctx, loggerKey, logger)
}

// FromContext return the logger from a context if any,
// if no logger in the context, it returns a default Logger
func FromContext(ctx context.Context) LoggerService {
	if l, ok := ctx.Value(loggerKey).(LoggerService); ok {
		return l
	}

	return Logger()
}

func WithRequestId(ctx context.Context, requestId string) LoggerService {
	logger := FromContext(ctx)
	logger.WithRequestId(requestId)
	return logger
}

func Copy(dst context.Context, src context.Context) {
	IntoContext(dst, FromContext(src))
}
