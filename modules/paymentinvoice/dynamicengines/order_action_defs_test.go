package dynamicengines

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/engine"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/constants"
)

// DefineAction validates a definition when it is called, which is during Init — so a malformed
// one takes the whole application down at start-up rather than failing a request. These tests
// move that failure to here.
//
// The rules it enforces are easy to trip over: a RestPath may not contain a hyphen, and an action
// that declares a RestPath must also declare an ActionType.

func newOrderTestEngine(t *testing.T) drif.DynamicResourceEngine {
	t.Helper()

	schema := dmodel.DefineModel("paymentinvoice_order").Build()
	testEngine := engine.NewDynamicResourceEngine(engine.NewEngineParam{Schema: schema})
	require.NoError(t, engine.DefineBuiltinActions(testEngine))
	return testEngine
}

func TestOrderActionsAreAccepted(t *testing.T) {
	testEngine := newOrderTestEngine(t)

	require.NoError(t, defineOrderActions(testEngine))

	for _, name := range []string{
		constants.ActionCreatePayment,
		constants.ActionRefund,
		constants.ActionRemovePosOrders,
	} {
		_, exists := testEngine.Action(name)
		assert.True(t, exists, "action '%s' was not defined", name)
	}
}

// Each action carries its own permission rather than reusing "update". Being allowed to correct
// an order's description is not the same authority as being allowed to hand money back, and one
// permission covering both would make the smaller grant imply the larger.
//
// The permission codes are also what the IAM seed grants, so a code changed here without the seed
// denies every request with nothing in the response pointing at why.
func TestEachMoneyActionHasItsOwnPermission(t *testing.T) {
	testEngine := newOrderTestEngine(t)
	require.NoError(t, defineOrderActions(testEngine))

	permissions := map[string]string{}
	for _, name := range []string{
		constants.ActionCreatePayment,
		constants.ActionRefund,
		constants.ActionRemovePosOrders,
	} {
		definition, exists := testEngine.Action(name)
		require.True(t, exists)

		assert.Equal(t, name, definition.Permission,
			"action '%s' must assert its own permission", name)
		assert.Empty(t, permissions[definition.Permission],
			"'%s' and '%s' share a permission", name, permissions[definition.Permission])
		permissions[definition.Permission] = name
	}
}

// Taking a payment is neither a create of the order nor an update of it, so all three are Generic
// — which is also what maps them to POST and lets them carry a request body. A Read here would
// register them as GET and silently drop every body.
func TestTheMoneyActionsArePostedNotRead(t *testing.T) {
	testEngine := newOrderTestEngine(t)
	require.NoError(t, defineOrderActions(testEngine))

	for _, name := range []string{
		constants.ActionCreatePayment,
		constants.ActionRefund,
		constants.ActionRemovePosOrders,
	} {
		definition, exists := testEngine.Action(name)
		require.True(t, exists)

		assert.Equal(t, drif.ActionTypeGeneric, definition.ActionType, name)
		assert.Equal(t, "POST", definition.ActionType.HttpMethod(), name)
	}
}

// A hyphen in a RestPath is rejected by DefineAction outright, which is why the action names are
// snake_case. This pins the paths against someone "tidying" them into kebab-case.
func TestRestPathsAreSnakeCase(t *testing.T) {
	testEngine := newOrderTestEngine(t)
	require.NoError(t, defineOrderActions(testEngine))

	for name, expected := range map[string]string{
		constants.ActionCreatePayment:   "create_payment",
		constants.ActionRefund:          "refund",
		constants.ActionRemovePosOrders: "remove_pos_orders/:pos_id",
	} {
		definition, exists := testEngine.Action(name)
		require.True(t, exists)

		assert.Equal(t, expected, definition.RestPath)
		assert.Regexp(t, drif.RestPathRegex, definition.RestPath)
	}
}

// The order engine keeps the built-in CRUD alongside these three, so defining them must not
// displace anything: a resource whose search stopped working would be a far larger regression
// than the actions being absent.
func TestTheBuiltinActionsSurvive(t *testing.T) {
	testEngine := newOrderTestEngine(t)
	require.NoError(t, defineOrderActions(testEngine))

	for _, name := range []string{
		drif.ActionSearch,
		drif.ActionGetById,
		drif.ActionUpdate,
		drif.ActionGetSchema,
	} {
		_, exists := testEngine.Action(name)
		assert.True(t, exists, "built-in action '%s' was displaced", name)
	}
}

// Without the service installed, the actions must fail as a wiring error rather than panic on a
// nil dereference — the message is what tells whoever hits it that Init is missing a call.
func TestAMissingServiceIsAWiringError(t *testing.T) {
	original := orderService
	t.Cleanup(func() { orderService = original })
	orderService = nil

	_, err := requireOrderService()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "SetOrderService")
}
