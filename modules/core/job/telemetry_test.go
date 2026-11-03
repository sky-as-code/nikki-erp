package job

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.bryk.io/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

func installTestProvider(t *testing.T) *tracetest.InMemoryExporter {
	t.Helper()
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	otel.SetTracerProvider(provider)
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })
	return exporter
}

type runObservation struct {
	idempotencyId string
	spanContext   trace.SpanContext
}

func observeRun(t *testing.T, jobFn JobHandleFn) (runObservation, error) {
	var seen runObservation
	err := Handle(logging.NewLogger(), "test.sweep", func(ctx context.Context, args *string) error {
		seen.idempotencyId, _ = corectx.IdempotencyIdOf(ctx)
		seen.spanContext = trace.SpanContextFromContext(ctx)
		return jobFn(ctx, args)
	}, time.Second, nil)
	return seen, err
}

// AC-20 / CTX-05: each run is its own root with its own id; the next periodic run is another.
func TestEachRunIsARootWithANewId(t *testing.T) {
	exporter := installTestProvider(t)
	noop := func(context.Context, *string) error { return nil }

	first, err := observeRun(t, noop)
	require.NoError(t, err)
	second, err := observeRun(t, noop)
	require.NoError(t, err)

	_, parseErr := corectx.ParseIdempotencyId(first.idempotencyId)
	assert.NoError(t, parseErr)
	assert.NotEqual(t, first.idempotencyId, second.idempotencyId)
	assert.NotEqual(t, first.spanContext.TraceID(), second.spanContext.TraceID())
	spans := exporter.GetSpans()
	require.Len(t, spans, 2)
	assert.Equal(t, RunSpanName, spans[0].Name)
	assert.False(t, spans[0].Parent.IsValid())
	assert.Contains(t, spans[0].Attributes, attribute.String(AttrJobName, "test.sweep"))
	assert.Equal(t, codes.Unset, spans[0].Status.Code)
}

func TestFailedOrPanickingRunMarksTheSpan(t *testing.T) {
	exporter := installTestProvider(t)

	_, err := observeRun(t, func(context.Context, *string) error { return errors.New("boom") })
	assert.Error(t, err)
	_, err = observeRun(t, func(context.Context, *string) error { panic("kaboom") })
	assert.Error(t, err, "a panic is contained and reported as the run's error")

	for _, span := range exporter.GetSpans() {
		assert.Equal(t, codes.Error, span.Status.Code)
	}
}

// ID-AC-09: the run's own deadline does not change its id.
func TestRunDeadlineKeepsTheId(t *testing.T) {
	installTestProvider(t)
	var derivedId string
	seen, err := observeRun(t, func(ctx context.Context, _ *string) error {
		child, cancel := context.WithTimeout(ctx, time.Millisecond)
		defer cancel()
		<-child.Done()
		derivedId, _ = corectx.IdempotencyIdOf(child)
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, seen.idempotencyId, derivedId)
}

// A retry of one execution keeps the id derived from its record (ID-AC-19).
func TestIdempotencyIdFromUlidIsAStableUuidV7(t *testing.T) {
	const executionId = "01M2JBE0000000001000000000"

	first := IdempotencyIdFromUlid(executionId)
	again := IdempotencyIdFromUlid(executionId)
	other := IdempotencyIdFromUlid("01M2JBE0000000001000000001")

	parsed, err := corectx.ParseIdempotencyId(first)
	require.NoError(t, err)
	assert.Equal(t, first, parsed)
	assert.Equal(t, first, again)
	assert.NotEqual(t, first, other)
	assert.Empty(t, IdempotencyIdFromUlid("not-a-ulid"))
}

func TestStartRunWithIdKeepsTheGivenId(t *testing.T) {
	installTestProvider(t)
	const id = "019a46c2-2a40-7b35-9e0f-8e2fd22b9b6c"
	ctx, span := StartRunWithId(context.Background(), "test.sweep", id)
	span.End()
	got, _ := corectx.IdempotencyIdOf(ctx)
	assert.Equal(t, id, got)
}
