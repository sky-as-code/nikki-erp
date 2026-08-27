package external

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/models"
)

func newTestDispatcher() *ActionDispatcher {
	bus := newFakeBus("inventory_maintenance.rebuild")
	commandBus := NewCommandBusExecutor(bus)
	restApi := newTestExecutor()
	return NewActionDispatcher(ActionExecutors{
		commandBus.ActionType(): commandBus,
		restApi.ActionType():    restApi,
	})
}

// An action type nothing can run must be refused at registration. Accepting it would produce a
// job that is registered and scheduled and fails at every single occurrence, reporting the
// mistake at 3am rather than at the moment somebody was looking at the response.
func TestUnknownActionTypeIsRejected(t *testing.T) {
	errs := newTestDispatcher().ValidateActionConfig("smoke_signal", map[string]any{})

	require.NotNil(t, errs)
	assert.True(t, errs.Has(models.JobFieldActionType))
}

func TestValidationIsRoutedToTheMatchingExecutor(t *testing.T) {
	dispatcher := newTestDispatcher()

	// A command name the fake bus does not know: only the command-bus executor could produce
	// this error, so its presence proves the routing.
	errs := dispatcher.ValidateActionConfig(models.ActionTypeCommandBus, map[string]any{
		"command_name": "inventory_maintenance.rebiuld",
	})

	require.NotNil(t, errs)
	assert.True(t, errs.Has("action_config.command_name"))
}

func TestValidConfigPassesThrough(t *testing.T) {
	dispatcher := newTestDispatcher()

	errs := dispatcher.ValidateActionConfig(models.ActionTypeCommandBus, map[string]any{
		"command_name": "inventory_maintenance.rebuild",
	})

	assert.Nil(t, errs)
}

// A nil config is forwarded rather than short-circuited, so the executor's own message about the
// missing key is what the caller sees. Refusing here would replace a specific error with a
// generic one, or with none.
func TestNilConfigReachesTheExecutor(t *testing.T) {
	errs := newTestDispatcher().ValidateActionConfig(models.ActionTypeRestApi, nil)

	require.NotNil(t, errs, "the REST executor requires a url and a method")
	assert.False(t, errs.Has(models.JobFieldActionType),
		"a missing key is the executor's error, not an unknown-action-type error")
}

func TestExecutorForAnswersNilForAnUnknownType(t *testing.T) {
	dispatcher := newTestDispatcher()

	assert.Nil(t, dispatcher.ExecutorFor("smoke_signal"))
	assert.NotNil(t, dispatcher.ExecutorFor(models.ActionTypeRestApi))
	assert.NotNil(t, dispatcher.ExecutorFor(models.ActionTypeCommandBus))
}

// The dispatcher and the executors must agree on the type strings. A mismatch would make every
// job of that type unregistrable while the executor sat there working.
func TestDispatcherKeysMatchTheExecutorActionTypes(t *testing.T) {
	dispatcher := newTestDispatcher()

	for actionType, executor := range dispatcher.executors {
		assert.Equal(t, actionType, executor.ActionType(),
			"the map key must be the executor's own action type, not a second copy of it")
	}
}
