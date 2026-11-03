package cqrs

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/sky-as-code/nikki-erp/modules/core/telemetry"
)

// The W3C keys and the idempotency id are reserved like the delivery keys: a caller forging
// them through WithMetadata would graft its request onto somebody else's trace.
const (
	MetaTraceParent   = "traceparent"
	MetaTraceState    = "tracestate"
	MetaBaggage       = "baggage"
	MetaIdempotencyId = telemetry.MetadataIdempotencyId
)

// injectTelemetry stamps the sender's span context, allowlisted baggage and idempotency id.
func injectTelemetry(ctx context.Context, msg *message.Message) {
	telemetry.InjectToMetadata(ctx, msg.Metadata)
}

// startHandlerSpan continues the sender's trace. The bus is an in-process request/reply, so the
// handler is a child of the sender rather than a linked root; the idempotency id is the
// sender's and is never minted here, since a missing one is a propagation bug to surface.
func startHandlerSpan(ctx context.Context, msg *message.Message, topic string) (context.Context, trace.Span) {
	ctx = telemetry.ExtractFromMetadata(ctx, msg.Metadata)
	return telemetry.Tracer(telemetry.ScopeName).Start(ctx, topic, trace.WithSpanKind(trace.SpanKindInternal))
}

func endHandlerSpan(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "cqrs handler failed")
	}
	span.End()
}

// startHandling builds the handler context: the bus timeout, the sender's span and idempotency
// id, then the request context as before. The caller ends the span and calls the cancel func.
func (this *WatermillCqrsBus) startHandling(
	msg *message.Message, topic string,
) (context.Context, trace.Span, context.CancelFunc) {
	timedCtx, cancel := context.WithTimeout(context.Background(), this.maxTimeout)
	spanCtx, span := startHandlerSpan(timedCtx, msg, topic)
	return this.createIncomingContext(spanCtx, msg), span, cancel
}
