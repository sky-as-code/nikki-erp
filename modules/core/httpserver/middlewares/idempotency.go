package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	"go.bryk.io/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/core/telemetry"
)

const (
	HeaderIdempotencyId = telemetry.HeaderIdempotencyId

	attrIdempotencyId = "coremart.idempotency_id"
)

var errIdempotencyIdAmbiguous = errors.New("idempotency id: more than one value")

// IdempotencyMiddleware takes a valid X-Idempotency-ID over or mints one when the header is
// absent, and answers it in the response. A header that is present but empty, malformed or
// repeated is a 400: replacing it would hide a caller that loses its ID between retries.
//
// The 400 is answered through rejectVia (the CORS middleware in the server chain): CORS sits
// further in, so without it a cross-origin browser could not read the rejection. Only this path
// runs it, so an accepted request still gets its CORS headers exactly once, from the chain.
func IdempotencyMiddleware(logger logging.LoggerService, rejectVia ...echo.MiddlewareFunc) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(echoCtx *echo.Context) error {
			request := echoCtx.Request()
			id, rejection, err := resolveIdempotencyId(request.Header)
			if err != nil {
				return err
			}
			if rejection != nil {
				logIdempotencyRejection(request.Context(), logger, rejection)
				return wrapHandler(respondInvalidIdempotencyId, rejectVia)(echoCtx)
			}
			ctx := corectx.WithIdempotencyId(request.Context(), id)
			trace.SpanFromContext(ctx).SetAttributes(attribute.String(attrIdempotencyId, id))
			echoCtx.Response().Header().Set(HeaderIdempotencyId, id)
			echoCtx.SetRequest(request.WithContext(ctx))
			return next(echoCtx)
		}
	}
}

// resolveIdempotencyId reads every value before parsing: Header.Get would silently pick the first.
// rejection is the client's fault; err is a generator failure, answered as a 500 (FMT-05).
func resolveIdempotencyId(header http.Header) (id string, rejection error, err error) {
	values := header.Values(HeaderIdempotencyId)
	if len(values) == 0 {
		id, err = corectx.NewIdempotencyId()
		return id, nil, err
	}
	if len(values) > 1 || strings.Contains(values[0], ",") {
		return "", errIdempotencyIdAmbiguous, nil
	}
	id, rejection = corectx.ParseIdempotencyId(values[0])
	return id, rejection, nil
}

func respondInvalidIdempotencyId(echoCtx *echo.Context) error {
	return echoCtx.JSON(http.StatusBadRequest, ft.ClientErrors{*newInvalidIdempotencyIdError()})
}

func wrapHandler(handler echo.HandlerFunc, middlewares []echo.MiddlewareFunc) echo.HandlerFunc {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

func newInvalidIdempotencyIdError() *ft.ClientErrorItem {
	return ft.NewValidationError(HeaderIdempotencyId, ft.ErrorKey("err_invalid_idempotency_id"),
		"The X-Idempotency-ID header must hold exactly one UUID version 7")
}

// logIdempotencyRejection records only the kind of fault, never the raw header value.
func logIdempotencyRejection(ctx context.Context, logger logging.LoggerService, rejection error) {
	logger.Info(ctx, "rejected X-Idempotency-ID header", logging.Attr{"reason": rejection.Error()})
}
