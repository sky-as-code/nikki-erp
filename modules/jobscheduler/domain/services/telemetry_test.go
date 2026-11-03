package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/job"
	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/infra/repository"
)

func attemptOf(executionId string, number int32) claimedWork {
	return claimedWork{
		Execution: repository.ClaimedExecution{Id: model.Id(executionId)},
		Attempt:   attemptRecord{Number: number},
	}
}

func attemptIdOf(t *testing.T, item claimedWork) string {
	ctx, span := startAttemptSpan(context.Background(), item, "inventory.rebuild")
	span.End()
	id, ok := corectx.IdempotencyIdOf(ctx)
	require.True(t, ok)
	return id
}

// ID-AC-19 / CTX-05: the retries of one execution share its id, the next occurrence (a new
// execution row) gets another, and every attempt is a root span of its own.
func TestAttemptsOfOneExecutionShareTheId(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	otel.SetTracerProvider(provider)
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })

	firstTry := attemptIdOf(t, attemptOf("01M2JBE0000000001000000000", 1))
	retry := attemptIdOf(t, attemptOf("01M2JBE0000000001000000000", 2))
	nextRun := attemptIdOf(t, attemptOf("01M2JBE0000000002000000000", 1))

	assert.Equal(t, firstTry, retry)
	assert.NotEqual(t, firstTry, nextRun)
	spans := exporter.GetSpans()
	require.Len(t, spans, 3)
	for _, span := range spans {
		assert.Equal(t, job.RunSpanName, span.Name)
		assert.False(t, span.Parent.IsValid())
	}
	assert.NotEqual(t, spans[0].SpanContext.TraceID(), spans[1].SpanContext.TraceID())
}
