package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"unicode"

	"github.com/labstack/echo/v5"
	"go.bryk.io/pkg/errors"
	"go.opentelemetry.io/otel/trace"

	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

const panicMessageMaxRunes = 256

// RecoverMiddleware turns a handler panic into an error the HTTP error handler answers as a
// generic 500, as Echo's Recover did. The panic value may hold request data, so the message that
// reaches the log and the span is stripped of control characters and truncated. The error is
// created inside the deferred recover, so its stack trace still shows the panicking frames.
func RecoverMiddleware(logger logging.LoggerService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(echoCtx *echo.Context) (err error) {
			defer func() {
				recovered := recover()
				if recovered == nil {
					return
				}
				if recovered == http.ErrAbortHandler {
					panic(recovered)
				}
				err = handlePanic(echoCtx.Request().Context(), logger, recovered)
			}()
			return next(echoCtx)
		}
	}
}

func handlePanic(ctx context.Context, logger logging.LoggerService, recovered any) error {
	panicErr := errors.Errorf("panic recovered: %s", sanitizePanicMessage(recovered))
	trace.SpanFromContext(ctx).RecordError(panicErr)
	writePanicLog(ctx, logger, panicErr)
	return panicErr
}

func writePanicLog(ctx context.Context, logger logging.LoggerService, panicErr error) {
	logger.Error(ctx, "handler panicked", panicErr)
}

func sanitizePanicMessage(recovered any) string {
	message := []rune(strings.Map(func(char rune) rune {
		if unicode.IsControl(char) {
			return ' '
		}
		return char
	}, fmt.Sprintf("%v", recovered)))
	if len(message) > panicMessageMaxRunes {
		return string(message[:panicMessageMaxRunes]) + "..."
	}
	return string(message)
}
