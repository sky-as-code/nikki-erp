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

// DefineAction validates a definition when it is called, which is during Init — so a malformed one
// takes the application down at start-up rather than failing a request. These move that failure
// here, the same way the order actions' tests do.

func newInvoiceTestEngine(t *testing.T) drif.DynamicResourceEngine {
	t.Helper()

	schema := dmodel.DefineModel("paymentinvoice_invoice").Build()
	testEngine := engine.NewDynamicResourceEngine(engine.NewEngineParam{Schema: schema})
	require.NoError(t, engine.DefineBuiltinActions(testEngine))
	return testEngine
}

func TestTheIssueActionIsAccepted(t *testing.T) {
	testEngine := newInvoiceTestEngine(t)

	require.NoError(t, defineInvoiceActions(testEngine))

	_, exists := testEngine.Action(constants.ActionIssue)
	assert.True(t, exists)
}

// Issuing carries its own permission rather than reusing "update". Correcting a draft's note and
// closing it into an accounting document are different authorities, and one permission covering
// both would make the smaller grant imply the larger.
//
// The code is also what the IAM seed grants, so changing it here without the seed denies every
// request with nothing in the response pointing at why.
func TestIssueAssertsItsOwnPermission(t *testing.T) {
	testEngine := newInvoiceTestEngine(t)
	require.NoError(t, defineInvoiceActions(testEngine))

	definition, exists := testEngine.Action(constants.ActionIssue)
	require.True(t, exists)

	assert.Equal(t, constants.ActionIssue, definition.Permission)
	assert.NotEqual(t, drif.PermissionUpdate, definition.Permission)
}

// Issuing is neither a create of the invoice (it already exists) nor an update of it, so it is
// Generic — which is also what maps it to POST. A Read would register it as GET and drop the body.
func TestIssueIsPostedNotRead(t *testing.T) {
	testEngine := newInvoiceTestEngine(t)
	require.NoError(t, defineInvoiceActions(testEngine))

	definition, exists := testEngine.Action(constants.ActionIssue)
	require.True(t, exists)

	assert.Equal(t, drif.ActionTypeGeneric, definition.ActionType)
	assert.Equal(t, "POST", definition.ActionType.HttpMethod())
}

// The action is scoped to one invoice, so its path carries the id segment. A hyphen would be
// rejected by DefineAction outright, which is why the action name is snake_case.
func TestTheIssueRestPathCarriesTheInvoiceId(t *testing.T) {
	testEngine := newInvoiceTestEngine(t)
	require.NoError(t, defineInvoiceActions(testEngine))

	definition, exists := testEngine.Action(constants.ActionIssue)
	require.True(t, exists)

	assert.Equal(t, ":id/issue", definition.RestPath)
	assert.Regexp(t, drif.RestPathRegex, definition.RestPath)
}

// The invoice engine keeps its built-in CRUD alongside the action: a resource whose search stopped
// working would be a far larger regression than the action being absent.
func TestTheInvoiceBuiltinActionsSurvive(t *testing.T) {
	testEngine := newInvoiceTestEngine(t)
	require.NoError(t, defineInvoiceActions(testEngine))

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

// Without the service installed the action must fail as a wiring error rather than panic on a nil
// dereference — the message is what tells whoever hits it that Init is missing a call.
func TestAMissingInvoiceServiceIsAWiringError(t *testing.T) {
	original := invoiceService
	t.Cleanup(func() { invoiceService = original })
	invoiceService = nil

	_, err := requireInvoiceService()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "SetInvoiceService")
}
