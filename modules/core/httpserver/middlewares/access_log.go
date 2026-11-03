package middlewares

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"

	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

const accessLogMessage = "http request"

// AccessLogMiddleware writes one line per request: method, route template, status and duration,
// never the query string or headers. It hands a handler error to the HTTP error handler, as Echo's
// RequestLogger did, so the status it logs is the one sent. It reads the request context only
// after the inner middlewares ran, so the line carries the idempotency id they resolved.
func AccessLogMiddleware(logger logging.LoggerService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(echoCtx *echo.Context) error {
			started := time.Now()
			err := next(echoCtx)
			if err != nil {
				echoCtx.Echo().HTTPErrorHandler(echoCtx, err)
			}
			_, status := echo.ResolveResponseStatus(echoCtx.Response(), err)
			fields := logging.HttpFields{
				Method:     echoCtx.Request().Method,
				Route:      echoCtx.Path(),
				StatusCode: status,
				DurationMs: float64(time.Since(started).Microseconds()) / 1000,
			}
			writeAccessLog(echoCtx.Request().Context(), logger, fields)
			return nil
		}
	}
}

// writeAccessLog hands the HTTP fields over through ctx, the only way they reach the top level
// of a record: caller data stays nested and cannot pose as them.
func writeAccessLog(ctx context.Context, logger logging.LoggerService, fields logging.HttpFields) {
	write := logger.Info
	if fields.StatusCode >= http.StatusInternalServerError {
		write = logger.Warn
	}
	write(logging.WithHttpFields(ctx, fields), accessLogMessage, nil)
}
