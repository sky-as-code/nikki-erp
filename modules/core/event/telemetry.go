package event

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/telemetry"
)

// DeliveryTrace is what a message carried of its publisher's operation. The consumer runs after
// the publisher has moved on, so it gets a root span linked to Source rather than a child of it,
// while the idempotency id is kept as is (BR §6.2, CTX-06).
type DeliveryTrace struct {
	Topic string

	// Source is the publisher's span context; invalid for a message published without one.
	Source trace.SpanContext

	// Baggage is the allowlisted baggage the publisher sent.
	Baggage baggage.Baggage

	// IdempotencyId is the publisher's, or one minted once at this consumer boundary for a message
	// that carried none (an older publisher). Empty only when minting failed.
	IdempotencyId string
}

func injectTelemetry(ctx context.Context, msg *message.Message) {
	telemetry.InjectToMetadata(ctx, msg.Metadata)
}

func deliveryTraceOf(msg *message.Message, topic string) DeliveryTrace {
	extracted := telemetry.ExtractFromMetadata(context.Background(), msg.Metadata)
	id, ok := corectx.IdempotencyIdOf(extracted)
	if !ok {
		id, _ = corectx.NewIdempotencyId()
	}
	return DeliveryTrace{
		Topic:         topic,
		Source:        trace.SpanContextFromContext(extracted),
		Baggage:       baggage.FromContext(extracted),
		IdempotencyId: id,
	}
}

// NewDeliveryTrace is the trace of a message that carried no metadata at all, such as one read
// off a raw pub/sub channel: a fresh root span and an idempotency id minted at this boundary.
func NewDeliveryTrace(topic string) DeliveryTrace {
	id, _ := corectx.NewIdempotencyId()
	return DeliveryTrace{Topic: topic, IdempotencyId: id}
}

// StartHandling opens the consumer's root span and puts the delivery's idempotency id and
// baggage on ctx. Call it on the context the handler's request context is built from, and end
// the span once the handler returns.
func (this EventDelivery) StartHandling(ctx context.Context) (context.Context, trace.Span) {
	return this.Trace.StartHandling(ctx)
}

func (this DeliveryTrace) StartHandling(ctx context.Context) (context.Context, trace.Span) {
	if this.IdempotencyId != "" {
		ctx = corectx.WithIdempotencyId(ctx, this.IdempotencyId)
	}
	if this.Baggage.Len() > 0 {
		ctx = baggage.ContextWithBaggage(ctx, this.Baggage)
	}
	options := []trace.SpanStartOption{
		trace.WithNewRoot(),
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(attribute.String("messaging.destination.name", this.Topic)),
	}
	if this.Source.IsValid() {
		options = append(options, trace.WithLinks(trace.Link{SpanContext: this.Source}))
	}
	return telemetry.Tracer(telemetry.ScopeName).Start(ctx, this.Topic+" process", options...)
}
