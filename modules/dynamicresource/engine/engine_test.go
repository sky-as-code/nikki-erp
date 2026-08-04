package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/requestguard"
	it "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

func noopProcess(_ corectx.Context, _ it.ProcessInput) (*it.ActionResult, error) {
	return &it.ActionResult{HasData: true}, nil
}

// newTestEngine builds an engine over a schema that is never touched by these tests,
// which exercise only the in-memory action map.
func newTestEngine() *DynamicResourceEngineImpl {
	schema := dmodel.DefineModel("test_resource").Build()
	return NewDynamicResourceEngine(NewEngineParam{Schema: schema}).(*DynamicResourceEngineImpl)
}

func TestDefineAction(t *testing.T) {
	engine := newTestEngine()

	err := engine.DefineAction(it.DynamicActionDefinition{
		ActionName:  "send_invitation",
		Permission:  it.PermissionUpdate,
		MainProcess: noopProcess,
	})
	assert.NoError(t, err)

	definition, exists := engine.Action("send_invitation")
	assert.True(t, exists)
	assert.Equal(t, it.PermissionUpdate, definition.Permission)
	assert.Equal(t, []string{"send_invitation"}, engine.ActionNames())
}

func TestDefineActionRejectsDuplicate(t *testing.T) {
	engine := newTestEngine()
	definition := it.DynamicActionDefinition{ActionName: "dup", MainProcess: noopProcess}

	assert.NoError(t, engine.DefineAction(definition))

	err := engine.DefineAction(definition)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already defined")
}

func TestDefineActionRequiresMandatoryFields(t *testing.T) {
	engine := newTestEngine()

	err := engine.DefineAction(it.DynamicActionDefinition{MainProcess: noopProcess})
	assert.Error(t, err, "action name is mandatory")

	err = engine.DefineAction(it.DynamicActionDefinition{ActionName: "no_process"})
	assert.Error(t, err, "MainProcess is mandatory")
}

func TestModifyActionOverridesOnlyProvidedFields(t *testing.T) {
	engine := newTestEngine()
	keysToFetch := func(params dmodel.DynamicFields) dmodel.DynamicFields { return params }

	assert.NoError(t, engine.DefineAction(it.DynamicActionDefinition{
		ActionName:  "update_thing",
		Permission:  it.PermissionUpdate,
		KeysToFetch: keysToFetch,
		MainProcess: noopProcess,
	}))

	scope := requestguard.ResourceScopeDomain
	assert.NoError(t, engine.ModifyAction(it.DynamicActionDelta{
		ActionName:      "update_thing",
		PermissionScope: &scope,
	}))

	definition, _ := engine.Action("update_thing")
	assert.Equal(t, it.PermissionUpdate, definition.Permission, "untouched field is kept")
	assert.NotNil(t, definition.KeysToFetch, "untouched field is kept")
	assert.Equal(t, requestguard.ResourceScopeDomain, *definition.PermissionScope)
}

// The Permission field is a pointer in the delta precisely so that overriding it back to
// the empty string, which skips the permission check, stays expressible.
func TestModifyActionCanClearPermission(t *testing.T) {
	engine := newTestEngine()
	assert.NoError(t, engine.DefineAction(it.DynamicActionDefinition{
		ActionName:  "open_thing",
		Permission:  it.PermissionRead,
		MainProcess: noopProcess,
	}))

	open := ""
	assert.NoError(t, engine.ModifyAction(it.DynamicActionDelta{
		ActionName: "open_thing",
		Permission: &open,
	}))

	definition, _ := engine.Action("open_thing")
	assert.Equal(t, "", definition.Permission)
}

func TestModifyActionRejectsUnknownAction(t *testing.T) {
	engine := newTestEngine()

	err := engine.ModifyAction(it.DynamicActionDelta{ActionName: "ghost"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not defined")
}

func TestDefineBuiltinActions(t *testing.T) {
	engine := newTestEngine()

	assert.NoError(t, DefineBuiltinActions(engine))
	assert.Equal(t, []string{
		it.ActionCreate,
		it.ActionDelete,
		it.ActionExists,
		it.ActionGetById,
		it.ActionGetByUnique,
		it.ActionGetSchema,
		it.ActionSearch,
		it.ActionSetArchived,
		it.ActionUpdate,
	}, engine.ActionNames())

	// Every built-in must be replaceable, which is what makes the engine extensible.
	assert.NoError(t, engine.ModifyAction(it.DynamicActionDelta{
		ActionName:  it.ActionCreate,
		MainProcess: noopProcess,
	}))
}
