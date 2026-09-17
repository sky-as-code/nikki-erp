package message

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

	itMessage "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/message"
)

// This broker's messages are raw bytes, so unlike the internal event bus there is nowhere to put
// metadata: the publishing scope has to ride on the envelope. A consumer subscribes on a goroutine
// with no request behind it, and in a multi-tenant deployment every table it reads is tenant-scoped,
// so an event published unscoped fails on the consumer's first query.

type capturingPublisher struct {
	topic string
	body  []byte
}

func (this *capturingPublisher) Publish(_ context.Context, topic string, body []byte) error {
	this.topic = topic
	this.body = body
	return nil
}

func anEvent() itMessage.IntegrationEvent {
	return itMessage.IntegrationEvent{
		EventId:       "01M2MS9HKENQR772ESNJM6VP3C",
		EventType:     "SalesPaymentCaptured",
		AggregateId:   "01M2MS9HMX4ZC0QF4Z4W4B8BVW",
		SchemaVersion: "1",
		OccurredAt:    1789551494,
		OrgId:         "01JQZ0X0000000000000000002",
		Payload:       map[string]any{"sales_order_id": "01M2MS9HKENQR772ESNJM6VP3C"},
	}
}

func publishWith(t *testing.T, ctx context.Context) map[string]any {
	t.Helper()

	broker := &capturingPublisher{}
	require.NoError(t, NewPublisher(broker).Publish(ctx, anEvent()))

	decoded := map[string]any{}
	require.NoError(t, json.Unmarshal(broker.body, &decoded))
	return decoded
}

func TestPublishCarriesThePublishingScope(t *testing.T) {
	ctx := corectx.NewRequestContext(context.Background())
	ctx.SetDomainConstraints(dmodel.DynamicFields{
		"tenant_id": model.Id("01JQZ0X0000000000000000001"),
	})

	decoded := publishWith(t, ctx)

	assert.Equal(t,
		map[string]any{"tenant_id": "01JQZ0X0000000000000000001"},
		decoded["scope"],
		"a consumer has no other way to learn whose data the event is about")
}

// The sweep detaches its context so the publish outlives whatever cancels the job, which is what
// context.Background() used to be for - and why the scope was being dropped.
func TestPublishCarriesTheScopeThroughADetachedContext(t *testing.T) {
	sweepCtx := corectx.NewRequestContext(context.Background())
	sweepCtx.SetDomainConstraints(dmodel.DynamicFields{
		"tenant_id": model.Id("01JQZ0X0000000000000000001"),
	})

	detached := corectx.CloneWithInner(sweepCtx, context.WithoutCancel(sweepCtx.InnerContext()))

	decoded := publishWith(t, detached)

	assert.Equal(t,
		map[string]any{"tenant_id": "01JQZ0X0000000000000000001"},
		decoded["scope"])
}

// A single-tenant build scopes nothing, and the key is omitted rather than sent as an empty object
// that a consumer would have to tell apart from a real scope.
func TestPublishOmitsAnAbsentScope(t *testing.T) {
	decoded := publishWith(t, context.Background())

	_, present := decoded["scope"]
	assert.False(t, present)
}

func TestPublishGoesToTheEventTypesOwnTopic(t *testing.T) {
	broker := &capturingPublisher{}
	require.NoError(t, NewPublisher(broker).Publish(context.Background(), anEvent()))

	assert.Equal(t, TopicOf("SalesPaymentCaptured"), broker.topic)
}
