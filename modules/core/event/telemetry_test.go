package event

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
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
	testTopic         = "test.events.telemetry"
	testIdempotencyId = "019a46c2-2a40-7b35-9e0f-8e2fd22b9b6c"
)

type probeEvent struct {
	Name string `json:"name"`
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

func newProbeBus(t *testing.T) (*EventBusImpl, chan EventDelivery) {
	pubSub := gochannel.NewGoChannel(gochannel.Config{}, watermill.NopLogger{})
	bus := &EventBusImpl{
		logger:     logging.NewLogger(),
		publisher:  pubSub,
		subscriber: pubSub,
		maxTimeout: time.Second,
		marshaler:  cqrs.JSONMarshaler{GenerateName: cqrs.NamedStruct(cqrs.StructName)},
	}
	subCtx, cancel := context.WithCancel(context.Background())
	t.Cleanup(func() {
		cancel()
		_ = pubSub.Close()
	})
	deliveries, err := bus.SubscribeEvent(subCtx, *NewEventRequest("", testTopic, "", nil), &probeEvent{})
	require.NoError(t, err)
	return bus, deliveries
}

func publishProbe(t *testing.T, bus *EventBusImpl, ctx context.Context) {
	body, err := json.Marshal(probeEvent{Name: "probe"})
	require.NoError(t, err)
	request := NewEventRequest("", testTopic, "", &message.Message{Payload: body})
	require.NoError(t, bus.PublishRequest(ctx, *request))
}

func receive(t *testing.T, deliveries chan EventDelivery) EventDelivery {
	t.Helper()
	select {
	case delivery := <-deliveries:
		return delivery
	case <-time.After(2 * time.Second):
		require.FailNow(t, "no delivery")
		return EventDelivery{}
	}
}

// BR §6.2 / AC-20: the consumer is a new root linked to the publisher, and keeps its id.
func TestConsumerIsALinkedRootKeepingTheId(t *testing.T) {
	exporter := installTestProvider(t)
	bus, deliveries := newProbeBus(t)
	pubCtx := corectx.WithIdempotencyId(context.Background(), testIdempotencyId)
	pubCtx, publisher := otel.Tracer("test").Start(pubCtx, "publisher")

	publishProbe(t, bus, pubCtx)
	publisher.End()
	handleCtx, consumer := receive(t, deliveries).StartHandling(context.Background())
	consumer.End()

	id, ok := corectx.IdempotencyIdOf(handleCtx)
	assert.True(t, ok)
	assert.Equal(t, testIdempotencyId, id)
	stub := exporter.GetSpans()[len(exporter.GetSpans())-1]
	assert.Equal(t, testTopic+" process", stub.Name)
	assert.Equal(t, trace.SpanKindConsumer, stub.SpanKind)
	assert.False(t, stub.Parent.IsValid())
	assert.NotEqual(t, publisher.SpanContext().TraceID(), stub.SpanContext.TraceID())
	require.Len(t, stub.Links, 1)
	assert.Equal(t, publisher.SpanContext().SpanID(), stub.Links[0].SpanContext.SpanID())
}

// An old message with no metadata is handled normally under a fresh trace and a fresh id,
// minted once at the boundary.
func TestMessageWithoutMetadataGetsAFreshTraceAndId(t *testing.T) {
	exporter := installTestProvider(t)
	bus, deliveries := newProbeBus(t)
	body, err := json.Marshal(probeEvent{Name: "legacy"})
	require.NoError(t, err)
	require.NoError(t, bus.publisher.Publish(testTopic,
		message.NewMessage(watermill.NewUUID(), body), message.NewMessage(watermill.NewUUID(), body)))

	first, second := receive(t, deliveries), receive(t, deliveries)
	firstCtx, firstSpan := first.StartHandling(context.Background())
	firstSpan.End()

	assert.Equal(t, "legacy", first.Payload.(*probeEvent).Name)
	id, ok := corectx.IdempotencyIdOf(firstCtx)
	require.True(t, ok)
	_, parseErr := corectx.ParseIdempotencyId(id)
	assert.NoError(t, parseErr)
	assert.Equal(t, id, first.Trace.IdempotencyId, "minted once, at the boundary")
	assert.NotEqual(t, first.Trace.IdempotencyId, second.Trace.IdempotencyId)
	stub := exporter.GetSpans()[0]
	assert.False(t, stub.Parent.IsValid())
	assert.Empty(t, stub.Links)
}

func TestRawChannelTraceMintsAnId(t *testing.T) {
	installTestProvider(t)
	ctx, span := NewDeliveryTrace(testTopic).StartHandling(context.Background())
	span.End()

	id, ok := corectx.IdempotencyIdOf(ctx)
	assert.True(t, ok)
	_, err := corectx.ParseIdempotencyId(id)
	assert.NoError(t, err)
	assert.True(t, trace.SpanContextFromContext(ctx).IsValid())
}
