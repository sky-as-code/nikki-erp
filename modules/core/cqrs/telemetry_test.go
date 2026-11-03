package cqrs

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

const (
	firstIdempotencyId  = "019a46c2-2a40-7b35-9e0f-8e2fd22b9b6c"
	secondIdempotencyId = "019a46c2-2a40-7b35-9e0f-8e2fd22b9b6d"
)

type probeRequest struct {
	Delay time.Duration `json:"delay"`
}

func (this probeRequest) CqrsRequestType() RequestType {
	return RequestType{Module: "test", Submodule: "telemetry", Action: "probe"}
}

// probeResult is what the handler saw on its context.
type probeResult struct {
	IdempotencyId string `json:"idempotency_id"`
	TraceId       string `json:"trace_id"`
	SpanId        string `json:"span_id"`
	CallerMeta    string `json:"caller_meta"`
}

func installTestProvider(t *testing.T) *tracetest.InMemoryExporter {
	t.Helper()
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })
	return exporter
}

func newProbeBus(t *testing.T) *WatermillCqrsBus {
	pubSub := gochannel.NewGoChannel(gochannel.Config{}, watermill.NopLogger{})
	bus := &WatermillCqrsBus{
		logger:     logging.NewLogger(),
		publisher:  pubSub,
		subscriber: pubSub,
		marshaler:  cqrs.JSONMarshaler{GenerateName: cqrs.NamedStruct(cqrs.StructName)},
		maxTimeout: 5 * time.Second,
	}
	handler := NewHandler(func(ctx context.Context, packet *RequestPacket[probeRequest]) (*Reply[probeResult], error) {
		time.Sleep(packet.Request().Delay)
		id, _ := corectx.IdempotencyIdOf(ctx)
		spanCtx := trace.SpanContextFromContext(ctx)
		return &Reply[probeResult]{Result: probeResult{
			IdempotencyId: id,
			TraceId:       spanCtx.TraceID().String(),
			SpanId:        spanCtx.SpanID().String(),
			CallerMeta:    MetadataFrom(ctx)[MetaIdempotencyKey],
		}}, nil
	})
	subCtx, cancel := context.WithCancel(context.Background())
	t.Cleanup(func() {
		cancel()
		_ = pubSub.Close()
	})
	require.NoError(t, bus.SubscribeRequests(subCtx, handler))
	return bus
}

func senderContext(id string) (context.Context, trace.Span) {
	ctx := corectx.WithIdempotencyId(context.Background(), id)
	return otel.Tracer("test").Start(ctx, "sender")
}

// The bus is an in-process request/reply: the handler span is a child of the sender (A13), and
// the id travels unchanged, through the sender's deadline too (ID-AC-09).
func TestHandlerContinuesTheSenderTraceAndId(t *testing.T) {
	exporter := installTestProvider(t)
	bus := newProbeBus(t)
	ctx, sender := senderContext(firstIdempotencyId)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var result probeResult
	require.NoError(t, bus.Request(ctx, probeRequest{}, &result))
	sender.End()

	assert.Equal(t, firstIdempotencyId, result.IdempotencyId)
	assert.Equal(t, sender.SpanContext().TraceID().String(), result.TraceId)
	handlerSpan := findSpan(t, exporter, "cqrs:test_telemetry.probe")
	assert.Equal(t, trace.SpanKindInternal, handlerSpan.SpanKind)
	assert.Equal(t, sender.SpanContext().SpanID(), handlerSpan.Parent.SpanID())
	assert.Equal(t, handlerSpan.SpanContext.SpanID().String(), result.SpanId)
}

// AC-10: concurrent requests carrying different ids and traces never see each other's.
func TestConcurrentRequestsDoNotLeakContext(t *testing.T) {
	installTestProvider(t)
	bus := newProbeBus(t)
	ids := []string{firstIdempotencyId, secondIdempotencyId}
	results := make([]probeResult, len(ids))
	senders := make([]trace.Span, len(ids))

	var wg sync.WaitGroup
	for i, id := range ids {
		ctx, sender := senderContext(id)
		senders[i] = sender
		wg.Add(1)
		go func(index int, reqCtx context.Context) {
			defer wg.Done()
			assert.NoError(t, bus.Request(reqCtx, probeRequest{Delay: 20 * time.Millisecond}, &results[index]))
		}(i, ctx)
	}
	wg.Wait()

	for i, id := range ids {
		senders[i].End()
		assert.Equal(t, id, results[i].IdempotencyId)
		assert.Equal(t, senders[i].SpanContext().TraceID().String(), results[i].TraceId)
	}
}

// A caller cannot forge the trace or the id through WithMetadata, and the handler does not see
// the bus-owned keys as caller metadata.
func TestTelemetryKeysAreReserved(t *testing.T) {
	installTestProvider(t)
	bus := newProbeBus(t)
	ctx, sender := senderContext(firstIdempotencyId)
	defer sender.End()
	ctx = WithMetadata(ctx, map[string]string{
		MetaTraceParent:    "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01",
		MetaIdempotencyId:  secondIdempotencyId,
		MetaIdempotencyKey: "kept",
	})
	assert.NotContains(t, MetadataFrom(ctx), MetaTraceParent)
	assert.NotContains(t, MetadataFrom(ctx), MetaIdempotencyId)

	var result probeResult
	require.NoError(t, bus.Request(ctx, probeRequest{}, &result))

	assert.Equal(t, firstIdempotencyId, result.IdempotencyId)
	assert.Equal(t, sender.SpanContext().TraceID().String(), result.TraceId)
	assert.Equal(t, "kept", result.CallerMeta)
}

// A sender without an id gets none minted by the bus: a lost id is a propagation bug to surface.
func TestBusNeverMintsAnId(t *testing.T) {
	installTestProvider(t)
	bus := newProbeBus(t)

	var result probeResult
	require.NoError(t, bus.Request(context.Background(), probeRequest{}, &result))

	assert.Empty(t, result.IdempotencyId)
}

// findSpan waits for the span: the handler ends its span after it has sent the reply.
func findSpan(t *testing.T, exporter *tracetest.InMemoryExporter, name string) tracetest.SpanStub {
	t.Helper()
	var found tracetest.SpanStub
	require.Eventually(t, func() bool {
		for _, span := range exporter.GetSpans() {
			if span.Name == name {
				found = span
				return true
			}
		}
		return false
	}, 2*time.Second, 5*time.Millisecond, "no span named %q", name)
	return found
}
