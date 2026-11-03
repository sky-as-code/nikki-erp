package job

import (
	"context"

	"github.com/google/uuid"
	"go.bryk.io/pkg/ulid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/telemetry"
)

// RunSpanName names every job run's root span; the job is an attribute, never part of the name.
const RunSpanName = "job.run"

const AttrJobName = "coremart.job.name"

// StartRun opens a run's root span with a new idempotency id. A run has no request behind it,
// and the next periodic run is an operation of its own (BR §6.2, CTX-05).
func StartRun(ctx context.Context, jobName string) (context.Context, trace.Span) {
	return StartRunWithId(ctx, jobName, "")
}

// StartRunWithId is StartRun for a run whose idempotency id comes from retry state that already
// exists, so every attempt of one logical execution shares it. An empty id mints a new one.
func StartRunWithId(ctx context.Context, jobName string, idempotencyId string) (context.Context, trace.Span) {
	if idempotencyId == "" {
		idempotencyId, _ = corectx.NewIdempotencyId()
	}
	if idempotencyId != "" {
		ctx = corectx.WithIdempotencyId(ctx, idempotencyId)
	}
	return telemetry.Tracer(telemetry.ScopeName).Start(ctx, RunSpanName,
		trace.WithNewRoot(),
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String(AttrJobName, jobName)),
	)
}

func EndRun(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "job run failed")
	}
	span.End()
}

// IdempotencyIdFromUlid maps a ULID to a UUIDv7 by setting the version and variant bits. Both
// put a millisecond timestamp in the first 48 bits, so the result is a valid, time-ordered
// UUIDv7 and the same ULID always yields the same id. Empty when raw is not a ULID.
func IdempotencyIdFromUlid(raw string) string {
	parsed, err := ulid.Parse(raw)
	if err != nil {
		return ""
	}
	var mapped uuid.UUID
	copy(mapped[:], parsed.Bytes())
	mapped[6] = (mapped[6] & 0x0f) | 0x70
	mapped[8] = (mapped[8] & 0x3f) | 0x80
	return mapped.String()
}
