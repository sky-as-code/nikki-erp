package middlewares

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/sky-as-code/nikki-erp/modules/core/telemetry"
)

const (
	HeaderTraceId = "X-Trace-ID"

	tracingScope = "github.com/sky-as-code/nikki-erp/modules/core/httpserver"
)

var knownHttpMethods = map[string]struct{}{
	http.MethodGet: {}, http.MethodHead: {}, http.MethodPost: {}, http.MethodPut: {}, http.MethodPatch: {},
	http.MethodDelete: {}, http.MethodConnect: {}, http.MethodOptions: {}, http.MethodTrace: {},
}

// TracingMiddleware must be the first middleware. It continues a valid inbound traceparent or
// starts a root SERVER span, and answers X-Trace-ID before anything else can write the response.
// A handler error is rendered here, inside the span, so the span records the status sent.
func TracingMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(echoCtx *echo.Context) error {
		request := echoCtx.Request()
		ctx := telemetry.Propagator().Extract(request.Context(), extractionCarrier(request.Header))
		route := echoCtx.Path()
		ctx, span := telemetry.Tracer(tracingScope).Start(ctx, serverSpanName(request.Method, route),
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(serverRequestAttributes(echoCtx, route)...),
		)
		defer span.End()

		echoCtx.Response().Header().Set(HeaderTraceId, span.SpanContext().TraceID().String())
		echoCtx.SetRequest(request.WithContext(ctx))

		err := next(echoCtx)
		if err != nil {
			echoCtx.Echo().HTTPErrorHandler(echoCtx, err)
		}
		_, status := echo.ResolveResponseStatus(echoCtx.Response(), err)
		finishServerSpan(span, status)
		return nil
	}
}

// extractionCarrier hides a repeated traceparent, and the tracestate that belongs to it, from the
// propagator: it reads only the first value, and trusting either of two parents would be a guess.
func extractionCarrier(header http.Header) propagation.HeaderCarrier {
	if len(header.Values("traceparent")) <= 1 {
		return propagation.HeaderCarrier(header)
	}
	unambiguous := header.Clone()
	unambiguous.Del("traceparent")
	unambiguous.Del("tracestate")
	return propagation.HeaderCarrier(unambiguous)
}

func serverSpanName(method string, route string) string {
	if _, known := knownHttpMethods[method]; !known {
		method = "HTTP"
	}
	if route == "" {
		return method
	}
	return method + " " + route
}

// serverRequestAttributes leaves out the query string and headers: they may carry tokens or
// personal data.
func serverRequestAttributes(echoCtx *echo.Context, route string) []attribute.KeyValue {
	request := echoCtx.Request()
	attrs := []attribute.KeyValue{
		semconv.URLPath(request.URL.Path),
		semconv.URLScheme(echoCtx.Scheme()),
		semconv.ServerAddress(request.Host),
		semconv.NetworkProtocolVersion(strconv.Itoa(request.ProtoMajor) + "." + strconv.Itoa(request.ProtoMinor)),
		semconv.UserAgentOriginal(request.UserAgent()),
	}
	if _, known := knownHttpMethods[request.Method]; known {
		attrs = append(attrs, semconv.HTTPRequestMethodKey.String(request.Method))
	} else {
		attrs = append(attrs, semconv.HTTPRequestMethodKey.String("_OTHER"),
			semconv.HTTPRequestMethodOriginal(request.Method))
	}
	if route != "" {
		attrs = append(attrs, semconv.HTTPRoute(route))
	}
	return attrs
}

// finishServerSpan follows the HTTP semantic conventions: a SERVER span is an error only for 5xx;
// a 4xx is the client's fault and leaves the status unset.
func finishServerSpan(span trace.Span, status int) {
	span.SetAttributes(semconv.HTTPResponseStatusCode(status))
	if status >= http.StatusInternalServerError {
		span.SetAttributes(semconv.ErrorTypeKey.String(strconv.Itoa(status)))
		span.SetStatus(codes.Error, "")
	}
}
