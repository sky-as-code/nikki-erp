package engine

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	ds "github.com/sky-as-code/nikki-erp/common/datastructure"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	it "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// memberContext is a caller who belongs to org_mine and to nothing else.
func memberContext() corectx.Context {
	ctx := corectx.NewRequestContext(context.Background())
	ctx.SetPermissions(corectx.ContextPermissions{
		IsOwner:    true,
		UserOrgIds: ds.NewSetFrom(model.Id("org_mine")),
	})
	return ctx
}

// orgScopedEngineWithAction wires an engine over an org-bearing schema plus one action that
// records the params it was handed, so a test can assert on what reached MainProcess.
func orgScopedEngineWithAction(
	t *testing.T, actionName string, seen *dmodel.DynamicFields,
) *DynamicResourceEngineImpl {
	t.Helper()

	engine := newOrgScopedTestEngine()
	require.NoError(t, engine.DefineAction(it.DynamicActionDefinition{
		ActionName: actionName,
		MainProcess: func(_ corectx.Context, input it.ProcessInput) (*it.ActionResult, error) {
			*seen = input.Params
			return &it.ActionResult{HasData: true}, nil
		},
	}))
	return engine
}

func TestExecuteActionRejectsMissingOrgId(t *testing.T) {
	var seen dmodel.DynamicFields
	engine := orgScopedEngineWithAction(t, "custom", &seen)

	result, err := engine.ExecuteAction(memberContext(), "custom", dmodel.DynamicFields{})

	require.NoError(t, err, "a missing org is the caller's mistake, not a server fault")
	assert.Positive(t, result.ClientErrors.Count())
	assert.Nil(t, seen, "the action must not run without an org")
}

func TestExecuteActionRejectsAnOrgTheCallerDoesNotBelongTo(t *testing.T) {
	var seen dmodel.DynamicFields
	engine := orgScopedEngineWithAction(t, "custom", &seen)

	result, err := engine.ExecuteAction(memberContext(), "custom", dmodel.DynamicFields{
		basemodel.FieldOrgId: "org_someone_else",
	})

	require.NoError(t, err)
	assert.Positive(t, result.ClientErrors.Count())
	assert.Nil(t, seen, "naming another org must not reach the action")
}

func TestExecuteActionAcceptsTheCallersOwnOrg(t *testing.T) {
	var seen dmodel.DynamicFields
	engine := orgScopedEngineWithAction(t, "custom", &seen)

	result, err := engine.ExecuteAction(memberContext(), "custom", dmodel.DynamicFields{
		basemodel.FieldOrgId: "org_mine",
	})

	require.NoError(t, err)
	assert.Zero(t, result.ClientErrors.Count())
	assert.Equal(t, "org_mine", seen[basemodel.FieldOrgId])
}

func TestExecuteActionSkipsScopingForAWithdrawnAction(t *testing.T) {
	var seen dmodel.DynamicFields
	engine := newOrgScopedTestEngine()
	require.NoError(t, engine.DefineAction(it.DynamicActionDefinition{
		ActionName:  "global",
		IsOrgScoped: util.ToPtr(false),
		MainProcess: func(_ corectx.Context, input it.ProcessInput) (*it.ActionResult, error) {
			seen = input.Params
			return &it.ActionResult{HasData: true}, nil
		},
	}))

	result, err := engine.ExecuteAction(memberContext(), "global", dmodel.DynamicFields{})

	require.NoError(t, err)
	assert.Zero(t, result.ClientErrors.Count())
	assert.NotNil(t, seen, "a withdrawn action runs without an org")
}

func TestExecuteActionSkipsScopingWhenSchemaHasNoOrgColumn(t *testing.T) {
	var seen dmodel.DynamicFields
	// newTestEngine's schema declares no org_id, so the resource cannot be org-filtered and
	// must keep working without one - this is what spares settings and reference tables.
	engine := newTestEngine()
	require.NoError(t, engine.DefineAction(it.DynamicActionDefinition{
		ActionName: "custom",
		MainProcess: func(_ corectx.Context, input it.ProcessInput) (*it.ActionResult, error) {
			seen = input.Params
			return &it.ActionResult{HasData: true}, nil
		},
	}))

	result, err := engine.ExecuteAction(memberContext(), "custom", dmodel.DynamicFields{})

	require.NoError(t, err)
	assert.Zero(t, result.ClientErrors.Count())
	assert.NotNil(t, seen)
}

func TestSearchGetsTheOrgConditionAppliedToItsGraph(t *testing.T) {
	var seen dmodel.DynamicFields
	engine := orgScopedEngineWithAction(t, it.ActionSearch, &seen)

	_, err := engine.ExecuteAction(memberContext(), it.ActionSearch, dmodel.DynamicFields{
		basemodel.FieldOrgId: "org_mine",
	})

	require.NoError(t, err)
	graph, ok := seen[queryParamGraph].(*dmodel.SearchGraph)
	require.True(t, ok, "search must reach the service carrying an org-constrained graph")
	assert.Len(t, graph.GetAnd(), 1)
}

// A caller with no permission at all must be denied, whether or not they sent an org. The
// org step runs first (it produces the org the permission check needs), so this guards
// against the org error becoming a way to probe a resource you cannot access.
func TestOrgScopeDoesNotBypassThePermissionCheck(t *testing.T) {
	var seen dmodel.DynamicFields
	engine := newOrgScopedTestEngine()
	require.NoError(t, engine.DefineAction(it.DynamicActionDefinition{
		ActionName: "custom",
		Permission: it.PermissionRead,
		MainProcess: func(_ corectx.Context, input it.ProcessInput) (*it.ActionResult, error) {
			seen = input.Params
			return &it.ActionResult{HasData: true}, nil
		},
	}))

	result, err := engine.ExecuteAction(deniedContext(), "custom", dmodel.DynamicFields{
		basemodel.FieldOrgId: "org_mine",
	})

	require.NoError(t, err)
	assert.Positive(t, result.ClientErrors.Count(), "a denied caller is still denied")
	assert.Nil(t, seen)
}
