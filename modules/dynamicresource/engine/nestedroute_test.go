package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/util"
	it "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

func nestedAction(name string, restPath string) it.DynamicActionDefinition {
	return it.DynamicActionDefinition{
		ActionName:         name,
		ActionType:         it.ActionTypeRead,
		RestPath:           restPath,
		PrimarySchema:      util.ToPtr("vdmc_kiosks"),
		PrimaryRestIdParam: util.ToPtr("kiosk_id"),
		MainProcess:        noopProcess,
	}
}

func TestFullRestPathIsFlatByDefault(t *testing.T) {
	engine := newTestEngine()
	restApi := NewDynamicRestApi(engine).(*DynamicRestApiImpl)

	path := restApi.fullRestPath(it.DynamicActionDefinition{RestPath: ":id"})

	assert.Equal(t, "/test_resource/:id", path)
}

func TestFullRestPathNestsUnderThePrimaryResource(t *testing.T) {
	engine := newTestEngine()
	restApi := NewDynamicRestApi(engine).(*DynamicRestApiImpl)

	path := restApi.fullRestPath(nestedAction("get_by_id", ":id"))

	// The shape the spec asks for: {primary-schema}/{primary-id}/{current-schema}/{current-id}
	assert.Equal(t, "/vdmc_kiosks/:kiosk_id/test_resource/:id", path)
}

func TestFullRestPathNestsTheBasePath(t *testing.T) {
	engine := newTestEngine()
	restApi := NewDynamicRestApi(engine).(*DynamicRestApiImpl)

	path := restApi.fullRestPath(nestedAction("search", ""))

	assert.Equal(t, "/vdmc_kiosks/:kiosk_id/test_resource", path)
}

func TestEngineAndRestApiAgreeOnTheRouteShape(t *testing.T) {
	engine := newTestEngine()
	restApi := NewDynamicRestApi(engine).(*DynamicRestApiImpl)

	// assertRouteFree compares actions by the engine's notion of a full path, while echo
	// registers the REST api's. If the two ever drift, duplicate routes stop being detected
	// and echo panics at boot instead.
	for _, definition := range []it.DynamicActionDefinition{
		{RestPath: ""},
		{RestPath: ":id"},
		{RestPath: "meta/schema"},
		nestedAction("get_by_id", ":id"),
		nestedAction("search", ""),
	} {
		assert.Equal(t, engine.fullRestPath(definition), restApi.fullRestPath(definition))
	}
}

func TestDefineActionRejectsPrimarySchemaWithoutIdParam(t *testing.T) {
	engine := newTestEngine()

	err := engine.DefineAction(it.DynamicActionDefinition{
		ActionName:    "get_by_id",
		ActionType:    it.ActionTypeRead,
		RestPath:      ":id",
		PrimarySchema: util.ToPtr("vdmc_kiosks"),
		MainProcess:   noopProcess,
	})

	assert.ErrorContains(t, err, "PrimaryRestIdParam")
}

func TestDefineActionRejectsIdParamWithoutPrimarySchema(t *testing.T) {
	engine := newTestEngine()

	err := engine.DefineAction(it.DynamicActionDefinition{
		ActionName:         "get_by_id",
		ActionType:         it.ActionTypeRead,
		RestPath:           ":id",
		PrimaryRestIdParam: util.ToPtr("kiosk_id"),
		MainProcess:        noopProcess,
	})

	assert.ErrorContains(t, err, "without a PrimarySchema")
}

func TestDefineActionRejectsHyphenatedPrimarySchema(t *testing.T) {
	engine := newTestEngine()

	err := engine.DefineAction(it.DynamicActionDefinition{
		ActionName:         "get_by_id",
		ActionType:         it.ActionTypeRead,
		RestPath:           ":id",
		PrimarySchema:      util.ToPtr("vdmc-kiosks"),
		PrimaryRestIdParam: util.ToPtr("kiosk_id"),
		MainProcess:        noopProcess,
	})

	assert.ErrorContains(t, err, "PrimarySchema")
}

func TestDefineActionRejectsColonPrefixedIdParam(t *testing.T) {
	engine := newTestEngine()

	// The engine writes the ":" itself, so declaring one would produce "/::kiosk_id".
	err := engine.DefineAction(it.DynamicActionDefinition{
		ActionName:         "get_by_id",
		ActionType:         it.ActionTypeRead,
		RestPath:           ":id",
		PrimarySchema:      util.ToPtr("vdmc_kiosks"),
		PrimaryRestIdParam: util.ToPtr(":kiosk_id"),
		MainProcess:        noopProcess,
	})

	assert.ErrorContains(t, err, "without the ':'")
}

func TestFlatAndNestedActionsMayShareARestPath(t *testing.T) {
	engine := newTestEngine()

	require.NoError(t, engine.DefineAction(it.DynamicActionDefinition{
		ActionName:  "get_by_id",
		ActionType:  it.ActionTypeRead,
		RestPath:    ":id",
		MainProcess: noopProcess,
	}))

	// Same RestPath and same method, but a different full path: this is a legitimate pair,
	// and comparing RestPath alone would reject it.
	assert.NoError(t, engine.DefineAction(nestedAction("get_by_id_in_kiosk", ":id")))
}

func TestDuplicateNestedRouteIsStillRejected(t *testing.T) {
	engine := newTestEngine()

	require.NoError(t, engine.DefineAction(nestedAction("get_by_id", ":id")))

	err := engine.DefineAction(nestedAction("get_by_id_again", ":id"))

	assert.ErrorContains(t, err, "already taken")
}

func TestRoutableActionsOrderNestedBasePathAfterFlatBasePath(t *testing.T) {
	engine := newTestEngine()
	// Both actions declare the same RestPath (""), so RestPath alone cannot tell them apart and
	// the old ordering fell through to the action-name tiebreaker - which put the nested route,
	// carrying a path param, ahead of the parameterless flat one.
	require.NoError(t, engine.DefineAction(nestedAction("a_nested_search", "")))
	require.NoError(t, engine.DefineAction(it.DynamicActionDefinition{
		ActionName:  "z_flat_search",
		ActionType:  it.ActionTypeRead,
		RestPath:    "",
		MainProcess: noopProcess,
	}))

	restApi := NewDynamicRestApi(engine).(*DynamicRestApiImpl)
	ordered := restApi.routableActions()

	// "/test_resource" has no path param; "/vdmc_kiosks/:kiosk_id/test_resource" has one.
	// Echo matches in registration order, so the parameterless route must be registered first
	// or the nested pattern swallows it.
	require.Len(t, ordered, 2)
	assert.Equal(t, "z_flat_search", ordered[0].ActionName)
	assert.Equal(t, "a_nested_search", ordered[1].ActionName)
}
