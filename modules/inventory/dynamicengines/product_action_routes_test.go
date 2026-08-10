package dynamicengines

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/engine"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// The Products capabilities are engine actions rather than hand-written routes, so their HTTP
// surface is decided by the action definition. A wrong ActionType or RestPath is invisible until
// a request 404s, and RestPath rejects hyphens outright — hence these.

func TestProductTemplateActionsAreDefined(t *testing.T) {
	engineImpl := newTestEngine(t, models.ProductTemplateSchemaName)

	require.NoError(t, defineProductTemplateActions(engineImpl))

	generate, ok := engineImpl.Action("generate_variants")
	require.True(t, ok, "generate_variants must be defined")
	assert.Equal(t, drif.ActionTypeGeneric, generate.ActionType, "a generic action is POSTed")
	assert.Equal(t, ":id/generate_variants", generate.RestPath)
	assert.Equal(t, drif.PermissionUpdate, generate.Permission,
		"regenerating a live product's variants is an update, not a read")
	assert.NotNil(t, generate.KeysToFetch, "the template is fetched for MainProcess")

	resolve, ok := engineImpl.Action("resolve_selection")
	require.True(t, ok, "resolve_selection must be defined")
	assert.Equal(t, drif.ActionTypeGeneric, resolve.ActionType)
	assert.Equal(t, "resolve_selection", resolve.RestPath)
	assert.Equal(t, drif.PermissionRead, resolve.Permission,
		"resolving a selection reads the catalog; materializing is guarded by the flag")
}

func TestProductVariantActionsAreDefined(t *testing.T) {
	engineImpl := newTestEngine(t, models.ProductVariantSchemaName)

	require.NoError(t, defineProductVariantActions(engineImpl))

	effective, ok := engineImpl.Action("get_effective")
	require.True(t, ok, "get_effective must be defined")
	assert.Equal(t, drif.ActionTypeRead, effective.ActionType, "a read action is GETed")
	assert.Equal(t, ":id/effective", effective.RestPath)
	assert.Equal(t, drif.PermissionRead, effective.Permission)
	assert.NotNil(t, effective.KeysToFetch, "the variant is fetched for MainProcess")
}

// RestPath is validated against ^:?[a-zA-Z0-9_]+(/:?[a-zA-Z0-9_]+)*$ — a hyphen is rejected at
// definition time, which is why these paths use underscores.
func TestActionRestPathsUseUnderscores(t *testing.T) {
	for _, path := range []string{":id/generate_variants", "resolve_selection", ":id/effective"} {
		assert.NotContainsf(t, path, "-", "RestPath %q must not contain a hyphen", path)
	}
}

// newTestEngine builds an engine carrying the built-in CRUD actions, which is what the registry
// hands a module: the Products specs modify set_archived and delete, so those must already exist.
func newTestEngine(t *testing.T, schemaName string) drif.DynamicResourceEngine {
	t.Helper()
	schema := dmodel.DefineModel(schemaName).Build()
	built := engine.NewDynamicResourceEngine(engine.NewEngineParam{Schema: schema})
	require.NoError(t, engine.DefineBuiltinActions(built))
	return built
}
