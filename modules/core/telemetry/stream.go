package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// AttrCancelled marks a span whose operation the client ended. It is not an error, so the
// status stays unset (BR §5, streaming).
const AttrCancelled = "coremart.cancelled"

func MarkCancelled(ctx context.Context) {
	trace.SpanFromContext(ctx).SetAttributes(attribute.Bool(AttrCancelled, true))
}
