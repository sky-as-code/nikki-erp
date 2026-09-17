package event

import (
	"context"
	"testing"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

// A subscriber consumes on a goroutine registered at boot, so the scope a publisher was executing
// under reaches it only if the bus carries it. Losing it is not a visible failure: it surfaces far
// away, as a repository panicking on a query with no tenant, and only in a multi-tenant build.

// recordingLogger implements only what the code under test calls. The embedded interface is nil, so
// any other method would panic rather than pass silently - which is the point.
type recordingLogger struct {
	logging.LoggerService
	errors []error
}

func (this *recordingLogger) Error(_ string, err error) {
	this.errors = append(this.errors, err)
}

func newTestMessage() *message.Message {
	return message.NewMessage(watermill.NewUUID(), nil)
}

func TestDomainConstraintsRoundTripThroughAMessage(t *testing.T) {
	constraints := dmodel.DynamicFields{
		"tenant_id": model.Id("01JQZ0X0000000000000000001"),
		"org_id":    model.Id("01JQZ0X0000000000000000002"),
	}
	ctx := corectx.NewRequestContext(context.Background())
	ctx.SetDomainConstraints(constraints)

	msg := newTestMessage()
	setDomainConstraints(ctx, msg)

	bus := &EventBusImpl{}
	assert.Equal(t, constraints, bus.domainConstraintsOf(msg, "test/topic"),
		"the subscriber must see exactly what the publisher was scoped by")
}

// A publisher hands the bus a plain context.Context, not the corectx.Context the constraints were
// set on - and a detached background context in the async case. Both still carry the values.
func TestDomainConstraintsSurviveADetachedContext(t *testing.T) {
	ctx := corectx.NewRequestContext(context.Background())
	ctx.SetDomainConstraints(dmodel.DynamicFields{"tenant_id": model.Id("01JQZ0X0000000000000000001")})

	detached := context.WithoutCancel(ctx.InnerContext())

	msg := newTestMessage()
	setDomainConstraints(detached, msg)

	bus := &EventBusImpl{}
	assert.Equal(t,
		dmodel.DynamicFields{"tenant_id": model.Id("01JQZ0X0000000000000000001")},
		bus.domainConstraintsOf(msg, "test/topic"))
}

// A single-tenant build scopes nothing, and must not end up with an empty constraints map that
// reads as "scoped to nothing" downstream.
func TestAnUnscopedPublishStampsNothing(t *testing.T) {
	msg := newTestMessage()
	setDomainConstraints(context.Background(), msg)

	assert.Empty(t, msg.Metadata.Get(MetaDomainConstraints))

	bus := &EventBusImpl{}
	assert.Nil(t, bus.domainConstraintsOf(msg, "test/topic"))
}

// Constraints that cannot be read are dropped, not fatal: the event itself decoded, and a handler
// running unscoped fails loudly on its first query, whereas discarding the message here would look
// like an event that was never published.
func TestUnreadableConstraintsAreLoggedAndDropped(t *testing.T) {
	msg := newTestMessage()
	msg.Metadata.Set(MetaDomainConstraints, "{not json")

	logger := &recordingLogger{}
	bus := &EventBusImpl{logger: logger}

	assert.Nil(t, bus.domainConstraintsOf(msg, "test/topic"))
	assert.Len(t, logger.errors, 1)
}

type testEvent struct {
	Name string   `json:"name"`
	Ids  []string `json:"ids"`
}

// The prototype is a type, not a buffer. Handing out one reused value is what forced every
// subscriber to copy the struct - and its slices - before using it.
func TestNewPayloadOfAllocatesAFreshValuePerMessage(t *testing.T) {
	prototype := &testEvent{}

	first, ok := newPayloadOf(prototype).(*testEvent)
	require.True(t, ok)
	second, ok := newPayloadOf(prototype).(*testEvent)
	require.True(t, ok)

	assert.NotSame(t, first, second)
	assert.NotSame(t, prototype, first)

	first.Name = "first"
	first.Ids = []string{"a"}

	assert.Empty(t, second.Name, "one message's payload must not be visible in the next")
	assert.Empty(t, second.Ids)
	assert.Empty(t, prototype.Name, "the prototype is never decoded into")
}

func TestScopeContextAppliesThePublishersScope(t *testing.T) {
	constraints := dmodel.DynamicFields{"tenant_id": model.Id("01JQZ0X0000000000000000001")}
	ctx := corectx.NewRequestContext(context.Background())

	EventDelivery{Constraints: constraints}.ScopeContext(ctx)

	assert.Equal(t, constraints, ctx.GetDomainConstraints())
}

func TestScopeContextLeavesAnUnscopedDeliveryAlone(t *testing.T) {
	ctx := corectx.NewRequestContext(context.Background())

	EventDelivery{}.ScopeContext(ctx)

	assert.Nil(t, ctx.GetDomainConstraints(),
		"an empty map would read as scoped-to-nothing rather than unscoped")
}
