package dynamicengines

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/engine"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/services"
)

// The profile's actions are wrappers around the built-in CRUD rather than actions of their own, so
// what has to be pinned is different from the order's and the invoice's: that the built-ins are
// still there, that the wrapping did not change what they are, and that the field the credentials
// are asked for by is resolved to the column that holds them.

func newPaymentProfileTestEngine(t *testing.T) drif.DynamicResourceEngine {
	t.Helper()

	schema := dmodel.DefineModel(models.PaymentProfileSchemaName).Build()
	testEngine := engine.NewDynamicResourceEngine(engine.NewEngineParam{Schema: schema})
	require.NoError(t, engine.DefineBuiltinActions(testEngine))

	return testEngine
}

// Wrapping must not displace a built-in or change the route it answers on: a profile whose search
// stopped working, or whose create moved to another verb, is a larger regression than the
// encryption being absent.
func TestTheProfileBuiltinActionsKeepTheirShape(t *testing.T) {
	testEngine := newPaymentProfileTestEngine(t)
	before := map[string]drif.DynamicActionDefinition{}
	for _, name := range testEngine.ActionNames() {
		definition, _ := testEngine.Action(name)
		before[name] = definition
	}

	require.NoError(t, definePaymentProfileActions(testEngine))

	for name, original := range before {
		definition, exists := testEngine.Action(name)
		require.Truef(t, exists, "built-in action '%s' was displaced", name)
		assert.Equalf(t, original.ActionType, definition.ActionType, "%s changed HTTP method", name)
		assert.Equalf(t, original.RestPath, definition.RestPath, "%s changed route", name)
		assert.Equalf(t, original.Permission, definition.Permission, "%s changed permission", name)
	}
}

// The two write actions bind their own body, because the engine's binding drops "config" — it is
// not a schema field. Without the handler a create would be refused and an update would silently
// persist a profile with no credentials.
func TestTheWriteActionsInstallTheirOwnBinding(t *testing.T) {
	testEngine := newPaymentProfileTestEngine(t)
	require.NoError(t, definePaymentProfileActions(testEngine))

	for _, name := range []string{drif.ActionCreate, drif.ActionUpdate} {
		definition, exists := testEngine.Action(name)
		require.True(t, exists)
		assert.NotNilf(t, definition.RestHandler, "'%s' must bind its own body", name)
	}
}

// The read actions keep the engine's binding: they carry no body, and their field selection is
// rewritten inside the action instead.
func TestTheReadActionsKeepTheEngineBinding(t *testing.T) {
	testEngine := newPaymentProfileTestEngine(t)
	require.NoError(t, definePaymentProfileActions(testEngine))

	for _, name := range []string{drif.ActionGetById, drif.ActionSearch} {
		definition, exists := testEngine.Action(name)
		require.True(t, exists)
		assert.Nilf(t, definition.RestHandler, "'%s' needs no binding of its own", name)
	}
}

// A caller selects the credentials by the name they are returned under. Left alone, that name
// selects a field the schema does not declare and the credentials come back missing.
func TestASelectedConfigResolvesToTheStoredColumn(t *testing.T) {
	params := dmodel.DynamicFields{
		"fields": []string{"name", models.PaymentProfileFieldConfig},
	}

	swapConfigForEncryptedConfig(params)

	assert.Equal(t,
		[]string{"name", models.PaymentProfileFieldEncryptedConfig},
		params["fields"])
}

// Query strings arrive as []any once they have been through a JSON body, so the swap has to handle
// both shapes or the credentials come back missing on one of the two routes.
func TestASelectedConfigResolvesFromAnUntypedList(t *testing.T) {
	params := dmodel.DynamicFields{
		"fields": []any{"name", models.PaymentProfileFieldConfig},
	}

	swapConfigForEncryptedConfig(params)

	assert.Equal(t,
		[]any{"name", models.PaymentProfileFieldEncryptedConfig},
		params["fields"])
}

// A search echoes back the fields it resolved. Reporting the column name would have the frontend
// look for a key the payload does not contain, because the response carries "config".
func TestASearchEchoesBackTheFieldTheCallerAskedFor(t *testing.T) {
	original := paymentProfileService
	t.Cleanup(func() { paymentProfileService = original })
	// A zero-value service needs no key: the result below carries no credentials to decrypt.
	paymentProfileService = &services.PaymentProfileDomainService{}

	result := &drif.ActionResult{
		HasData: true,
		Data: dyn.PagedResultData[dmodel.DynamicFields]{
			Items:         []dmodel.DynamicFields{},
			DesiredFields: []string{"name", models.PaymentProfileFieldEncryptedConfig},
		},
	}

	require.NoError(t, decryptProfileResult(result))

	paged := result.Data.(dyn.PagedResultData[dmodel.DynamicFields])
	assert.Equal(t, []string{"name", models.PaymentProfileFieldConfig}, paged.DesiredFields)
}

// Without the service installed the actions must fail as a wiring error rather than panic on a nil
// dereference — the message is what tells whoever hits it that Init is missing a call.
func TestAMissingProfileServiceIsAWiringError(t *testing.T) {
	original := paymentProfileService
	t.Cleanup(func() { paymentProfileService = original })
	paymentProfileService = nil

	_, err := requirePaymentProfileService()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "SetPaymentProfileService")
}
