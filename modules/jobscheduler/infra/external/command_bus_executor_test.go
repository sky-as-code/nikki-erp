package external

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

// fakeBus records what was sent and answers whatever the test set up.
type fakeBus struct {
	registered map[string]bool
	replyErr   error
	sentType   string
	sentMeta   map[string]string
	sendCount  int
}

func newFakeBus(registered ...string) *fakeBus {
	bus := &fakeBus{registered: map[string]bool{}}
	for _, name := range registered {
		bus.registered[name] = true
	}
	return bus
}

func (this *fakeBus) SubscribeRequests(context.Context, ...cqrs.RequestHandler) error { return nil }

func (this *fakeBus) RequestNoReply(context.Context, cqrs.Request) error { return nil }

func (this *fakeBus) Request(ctx context.Context, request cqrs.Request, _ any) error {
	this.sendCount++
	this.sentType = request.CqrsRequestType().String()
	this.sentMeta = cqrs.MetadataFrom(ctx)
	return this.replyErr
}

func (this *fakeBus) IsRequestTypeRegistered(requestType string) bool {
	return this.registered[requestType]
}

// A typo would otherwise become a job that runs on schedule forever, publishes into the void, and
// reports success every time. There is no later point at which the mistake surfaces on its own.
func TestValidateRejectsAnUnregisteredCommand(t *testing.T) {
	executor := NewCommandBusExecutor(newFakeBus("inventory_maintenance.rebuild"))

	errs := executor.Validate(nil, map[string]any{
		"command_name": "inventory_maintenance.rebiuld", // transposed
	})

	require.NotNil(t, errs)
	assert.True(t, errs.Has("action_config.command_name"))
}

func TestValidateAcceptsARegisteredCommand(t *testing.T) {
	executor := NewCommandBusExecutor(newFakeBus("inventory_maintenance.rebuild"))

	errs := executor.Validate(nil, map[string]any{
		"command_name": "inventory_maintenance.rebuild",
	})

	assert.Nil(t, errs)
}

func TestValidateRejectsMalformedCommandNames(t *testing.T) {
	bus := newFakeBus()
	executor := NewCommandBusExecutor(bus)

	for _, name := range []string{
		"",
		"nodot",
		"inventory.rebuild",   // missing the submodule segment
		"Inventory_Maint.Run", // module and submodule are lower case
		".rebuild",
		"inventory_maintenance.",
	} {
		errs := executor.Validate(nil, map[string]any{"command_name": name})
		require.NotNil(t, errs, "name %q should be rejected", name)
	}
}

func TestExecuteRoutesToTheNamedCommand(t *testing.T) {
	bus := newFakeBus("inventory_maintenance.rebuild")
	executor := NewCommandBusExecutor(bus)

	outcome := executor.Execute(context.Background(), testInput(map[string]any{
		"command_name": "inventory_maintenance.rebuild",
	}))

	require.True(t, outcome.Succeeded)
	assert.Equal(t, "inventory_maintenance.rebuild", bus.sentType,
		"the name must round-trip through parsing back to the same routing key")
}

// The reason for using Request over RequestNoReply: a handler that fails must produce a failed
// attempt. With no reply, every command attempt would be recorded as a success and the retry
// engine would be dead weight for command jobs.
func TestHandlerFailureBecomesAFailedRetryableAttempt(t *testing.T) {
	bus := newFakeBus("inventory_maintenance.rebuild")
	bus.replyErr = errors.New("the handler exploded")
	executor := NewCommandBusExecutor(bus)

	outcome := executor.Execute(context.Background(), testInput(map[string]any{
		"command_name": "inventory_maintenance.rebuild",
	}))

	assert.False(t, outcome.Succeeded, "a failing handler must not be recorded as success")
	assert.True(t, outcome.Retryable)
	assert.Equal(t, ErrorCodeCommand, outcome.ErrorCode)
}

// The same metadata the REST action sends as headers, so a handler can be made idempotent by the
// same means whichever transport invoked it.
func TestExecuteAttachesSchedulerMetadata(t *testing.T) {
	bus := newFakeBus("inventory_maintenance.rebuild")
	executor := NewCommandBusExecutor(bus)

	executor.Execute(context.Background(), testInput(map[string]any{
		"command_name": "inventory_maintenance.rebuild",
	}))

	require.NotNil(t, bus.sentMeta)
	assert.Equal(t, "inventory:rebuild:2026-08-20T10:00:00Z", bus.sentMeta[cqrs.MetaIdempotencyKey])
	assert.Equal(t, "01M2JBJ0000000001000000000", bus.sentMeta[cqrs.MetaSchedulerJobId])
	assert.Equal(t, "01M2JBE0000000001000000000", bus.sentMeta[cqrs.MetaSchedulerExecutionId])
	assert.Equal(t, "2", bus.sentMeta[cqrs.MetaSchedulerAttempt])
}

// A shutdown must not spend the job's retry budget: a rolling deploy would otherwise exhaust the
// attempts of every job that happened to be running.
func TestCancellationDuringShutdownIsRecognized(t *testing.T) {
	bus := newFakeBus("inventory_maintenance.rebuild")
	bus.replyErr = context.Canceled
	executor := NewCommandBusExecutor(bus)

	outcome := executor.Execute(context.Background(), testInput(map[string]any{
		"command_name": "inventory_maintenance.rebuild",
	}))

	assert.True(t, IsShutdownOutcome(outcome))
}

func TestExecuteRefusesAnUnusableNameWithoutSending(t *testing.T) {
	bus := newFakeBus()
	executor := NewCommandBusExecutor(bus)

	outcome := executor.Execute(context.Background(), testInput(map[string]any{
		"command_name": "not a command",
	}))

	assert.False(t, outcome.Succeeded)
	assert.False(t, outcome.Retryable, "a permanently bad name must not be retried")
	assert.Equal(t, 0, bus.sendCount, "nothing should reach the bus")
}

func TestParseRequestTypeRoundTrips(t *testing.T) {
	requestType, ok := parseRequestType("inventory_maintenance.rebuildSnapshot")

	require.True(t, ok)
	assert.Equal(t, "inventory", requestType.Module)
	assert.Equal(t, "maintenance", requestType.Submodule)
	assert.Equal(t, "rebuildSnapshot", requestType.Action)
	assert.Equal(t, "inventory_maintenance.rebuildSnapshot", requestType.String())
}

// The executor bounds each attempt itself; the bus has its own, longer, request timeout.
func TestExecuteAppliesThePerAttemptTimeout(t *testing.T) {
	bus := newFakeBus("inventory_maintenance.rebuild")
	executor := NewCommandBusExecutor(bus)

	in := testInput(map[string]any{"command_name": "inventory_maintenance.rebuild"})
	in.Timeout = 50 * time.Millisecond

	outcome := executor.Execute(context.Background(), in)

	assert.True(t, outcome.Succeeded, "a prompt handler still succeeds within the bound")
}
