package services

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/sky-as-code/nikki-erp/modules/core/job"
	itext "github.com/sky-as-code/nikki-erp/modules/jobscheduler/interfaces/external"
)

const tickJobName = "jobscheduler.tick"

// startTickSpan gives a tick its own root and idempotency id: a tick is scheduler housekeeping,
// and the trace must not span the scheduler's lifetime (BR §6.2).
func startTickSpan(ctx context.Context, reason WakeReason) (context.Context, trace.Span) {
	ctx, span := job.StartRun(ctx, tickJobName)
	span.SetAttributes(attribute.String("coremart.job.wake_reason", string(reason)))
	return ctx, span
}

// startAttemptSpan opens the attempt's root span. Its idempotency id is derived from the
// execution id, the retry state that already exists: every attempt of one execution shares it,
// while the next occurrence is a new execution row and so a new id (CTX-05).
func startAttemptSpan(ctx context.Context, item claimedWork, jobKey string) (context.Context, trace.Span) {
	ctx, span := job.StartRunWithId(ctx, jobKey, job.IdempotencyIdFromUlid(string(item.Execution.Id)))
	span.SetAttributes(attribute.Int("coremart.job.attempt", int(item.Attempt.Number)))
	return ctx, span
}

func markAttemptOutcome(ctx context.Context, outcome itext.ActionOutcome) {
	if !outcome.Succeeded {
		trace.SpanFromContext(ctx).SetStatus(codes.Error, outcome.ErrorCode)
	}
}
