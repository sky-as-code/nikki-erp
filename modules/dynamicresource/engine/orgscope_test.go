package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	it "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// newOrgScopedTestEngine builds an engine whose schema declares an org column, which is what
// switches the org-scoping machinery on.
func newOrgScopedTestEngine() *DynamicResourceEngineImpl {
	schema := dmodel.DefineModel("test_org_resource").
		Field(basemodel.DefineFieldId(basemodel.FieldOrgId)).
		Build()
	return NewDynamicResourceEngine(NewEngineParam{Schema: schema}).(*DynamicResourceEngineImpl)
}

func TestOrgScopedDefaultsToTrueWhenUnset(t *testing.T) {
	definition := it.DynamicActionDefinition{ActionName: "read"}

	// The default has to be the safe direction: an action that never mentions org scoping is
	// scoped, so forgetting the field cannot silently expose every org's rows.
	assert.True(t, definition.OrgScoped())
}

func TestOrgScopedHonoursExplicitValues(t *testing.T) {
	assert.True(t, it.DynamicActionDefinition{IsOrgScoped: util.ToPtr(true)}.OrgScoped())
	assert.False(t, it.DynamicActionDefinition{IsOrgScoped: util.ToPtr(false)}.OrgScoped())
}

func TestSchemaWithoutOrgIdIsNotScoped(t *testing.T) {
	// newTestEngine's schema declares no org column, so a resource that cannot be org-filtered
	// is left alone rather than being made unreachable.
	assert.False(t, newTestEngine().schemaHasOrgId())
	assert.True(t, newOrgScopedTestEngine().schemaHasOrgId())
}

func TestModifyActionWithdrawsOrgScoping(t *testing.T) {
	engine := newOrgScopedTestEngine()
	require.NoError(t, engine.DefineAction(it.DynamicActionDefinition{
		ActionName:  "read",
		MainProcess: noopProcess,
	}))

	require.NoError(t, engine.ModifyAction(it.DynamicActionDelta{
		ActionName:  "read",
		IsOrgScoped: util.ToPtr(false),
	}))

	definition, _ := engine.Action("read")
	assert.False(t, definition.OrgScoped())
}

func TestModifyActionLeavesOrgScopingAloneWhenDeltaIsNil(t *testing.T) {
	engine := newOrgScopedTestEngine()
	require.NoError(t, engine.DefineAction(it.DynamicActionDefinition{
		ActionName:  "read",
		IsOrgScoped: util.ToPtr(false),
		MainProcess: noopProcess,
	}))

	// A delta that says nothing about org scoping must not resurrect the default.
	require.NoError(t, engine.ModifyAction(it.DynamicActionDelta{
		ActionName: "read",
		Permission: util.ToPtr(it.PermissionRead),
	}))

	definition, _ := engine.Action("read")
	assert.False(t, definition.OrgScoped())
}

func TestWithdrawOrgScopingCoversEveryAction(t *testing.T) {
	engine := newOrgScopedTestEngine()
	require.NoError(t, DefineBuiltinActions(engine))
	require.NoError(t, engine.DefineAction(it.DynamicActionDefinition{
		ActionName:  "custom",
		MainProcess: noopProcess,
	}))

	require.NoError(t, WithdrawOrgScoping(engine))

	for _, name := range engine.ActionNames() {
		definition, _ := engine.Action(name)
		assert.Falsef(t, definition.OrgScoped(), "action %q should be withdrawn", name)
	}
}

func TestConstrainSearchGraphAddsOrgConditionWhenNoGraphSent(t *testing.T) {
	params := dmodel.DynamicFields{}

	require.NoError(t, constrainSearchGraph(params, "org_1"))

	graph, ok := params[queryParamGraph].(*dmodel.SearchGraph)
	require.True(t, ok, "graph should be replaced by a built SearchGraph")
	assert.Len(t, graph.GetAnd(), 1)
}

func TestConstrainSearchGraphWrapsTheCallersGraph(t *testing.T) {
	// The caller's graph arrives already decoded, the way searchParams leaves it.
	params := dmodel.DynamicFields{
		queryParamGraph: map[string]any{
			"if": []any{"name", "=", "widget"},
		},
	}

	require.NoError(t, constrainSearchGraph(params, "org_1"))

	graph, ok := params[queryParamGraph].(*dmodel.SearchGraph)
	require.True(t, ok)
	// Both the caller's condition and the org condition survive, ANDed together: the caller's
	// filter narrows the result, it never widens it past their org.
	assert.Len(t, graph.GetAnd(), 2)
}

func TestApplyOrgConstraintOnlyRewritesSearch(t *testing.T) {
	params := dmodel.DynamicFields{basemodel.FieldId: "rec_1"}

	require.NoError(t, applyOrgConstraint(it.ActionGetById, params, "org_1"))

	// A single-row action identifies its row by key, so it gets no graph.
	_, hasGraph := params[queryParamGraph]
	assert.False(t, hasGraph)
}
