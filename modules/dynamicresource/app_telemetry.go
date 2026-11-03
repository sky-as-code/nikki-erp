package dynamicresource

import (
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/telemetry"
)

// traceAction opens the INTERNAL span of one built-in application action, named
// {schema}.{action} (BE-05). Defer the returned func with the method's named error: a Go error
// is a 5xx-class failure and marks the span, while client errors travel in the result and leave
// the status unset.
func (this *DefaultApplicationServiceImpl) traceAction(ctx corectx.Context, action CrudAction) func(*error) {
	end := StartActionSpan(ctx, this.ResourceCode(), string(action))
	return func(outErr *error) {
		if recovered := recover(); recovered != nil {
			end(ft.RecoverPanicFailedTo(recovered, "run "+string(action)))
			panic(recovered)
		}
		end(*outErr)
	}
}

// StartActionSpan opens an application action's span as the current span of ctx, in place, so
// the domain service and repository below record under it, and returns the func that ends it.
// Ending makes the previous span current again while keeping every value the action set on ctx
// meanwhile (acting org, request action), which a derived clone would have hidden from the
// caller.
func StartActionSpan(ctx corectx.Context, schemaName string, action string) func(err error) {
	if ctx == nil {
		return func(error) {}
	}
	inner := ctx.InnerContext()
	parent := trace.SpanFromContext(inner)
	spanCtx, span := telemetry.Tracer(telemetry.ScopeName).Start(inner, schemaName+"."+action,
		trace.WithSpanKind(trace.SpanKindInternal))
	isCurrent := corectx.ReplaceInner(ctx, spanCtx)
	return func(err error) {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "")
		}
		span.End()
		if isCurrent {
			corectx.ReplaceInner(ctx, trace.ContextWithSpan(ctx.InnerContext(), parent))
		}
	}
}
