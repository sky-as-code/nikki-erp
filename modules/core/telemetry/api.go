package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

const (
	ScopeName = "github.com/sky-as-code/nikki-erp"

	// MetadataIdempotencyId is the bus/envelope metadata key; traceparent, tracestate and
	// baggage use their W3C header names.
	MetadataIdempotencyId = "idempotency_id"
)

func Tracer(scope string) trace.Tracer {
	return otel.Tracer(scope)
}

func Propagator() propagation.TextMapPropagator {
	return otel.GetTextMapPropagator()
}

// StartDetached starts a new root INTERNAL span for work that outlives the request, linked to
// the span ctx carries. The returned context keeps ctx's values (idempotency id included) but
// not its cancellation, so the request ending does not abort the work.
func StartDetached(ctx context.Context, name string) (context.Context, trace.Span) {
	options := []trace.SpanStartOption{trace.WithNewRoot(), trace.WithSpanKind(trace.SpanKindInternal)}
	if source := trace.SpanContextFromContext(ctx); source.IsValid() {
		options = append(options, trace.WithLinks(trace.Link{SpanContext: source}))
	}
	return Tracer(ScopeName).Start(context.WithoutCancel(ctx), name, options...)
}

// InjectToMetadata writes the active span context, allowlisted baggage and the idempotency id
// into an envelope's metadata.
func InjectToMetadata(ctx context.Context, metadata map[string]string) {
	if metadata == nil {
		return
	}
	Propagator().Inject(ctx, propagation.MapCarrier(metadata))
	if id, ok := corectx.IdempotencyIdOf(ctx); ok {
		metadata[MetadataIdempotencyId] = id
	}
}

// ExtractFromMetadata is InjectToMetadata's inverse. An invalid idempotency id is dropped, not
// replaced: the consumer decides whether its unit of work needs a new one.
func ExtractFromMetadata(ctx context.Context, metadata map[string]string) context.Context {
	if metadata == nil {
		return ctx
	}
	ctx = Propagator().Extract(ctx, propagation.MapCarrier(metadata))
	if id, err := corectx.ParseIdempotencyId(metadata[MetadataIdempotencyId]); err == nil {
		ctx = corectx.WithIdempotencyId(ctx, id)
	}
	return ctx
}
