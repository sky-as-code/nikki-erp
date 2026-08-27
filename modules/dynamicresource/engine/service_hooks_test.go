package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	it "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// newHookService builds a service on a stub repository whose action lookup reads a map the test
// fills in, which is the shape the registry produces with the engine's own Action method.
func newHookService(
	repo it.DynamicResourceRepository, definitions map[string]it.DynamicActionDefinition,
) *DynamicResourceServiceImpl {
	schema := dmodel.DefineModel("test_hooked_resource").Build()
	return NewDynamicResourceService(NewServiceParam{
		Schema:     schema,
		Repository: repo,
		ActionLookup: func(name string) (it.DynamicActionDefinition, bool) {
			definition, exists := definitions[name]
			return definition, exists
		},
	}).(*DynamicResourceServiceImpl)
}

// The order is the one every crud helper uses, so that a hook set behaves the same whichever
// operation carries it: sanitize, then the extra rules, then the last word before the write.
func TestServiceRunsHooksInCrudOrder(t *testing.T) {
	order := []string{}
	service := newHookService(&stubRepository{}, map[string]it.DynamicActionDefinition{
		it.ActionSearch: {
			ActionName: it.ActionSearch,
			BeforeValidation: func(_ corectx.Context, params dmodel.DynamicFields, _ *ft.ClientErrors) (dmodel.DynamicFields, error) {
				order = append(order, "before")
				params["sanitized"] = true
				return params, nil
			},
			AfterValidationSuccess: func(_ corectx.Context, _ dmodel.DynamicFields) error {
				order = append(order, "after")
				return nil
			},
			ValidateExtra: func(_ corectx.Context, params dmodel.DynamicFields, _ *dmodel.DynamicFields, _ *ft.ClientErrors) error {
				order = append(order, "extra")
				assert.Equal(t, true, params["sanitized"], "ValidateExtra sees what BeforeValidation produced")
				return nil
			},
		},
	})

	params, cErrs, err := service.runHooks(ownerContext(), it.ActionSearch, dmodel.DynamicFields{})

	assert.NoError(t, err)
	assert.Zero(t, cErrs.Count())
	assert.Equal(t, []string{"before", "extra", "after"}, order)
	assert.Equal(t, true, params["sanitized"])
}

// A guard that reports a violation stops the flow, so the operation never reaches its helper.
func TestServiceHooksStopOnClientError(t *testing.T) {
	service := newHookService(&stubRepository{}, map[string]it.DynamicActionDefinition{
		it.ActionSetArchived: {
			ActionName: it.ActionSetArchived,
			ValidateExtra: func(_ corectx.Context, _ dmodel.DynamicFields, _ *dmodel.DynamicFields, vErrs *ft.ClientErrors) error {
				vErrs.Append(*ft.NewAnonymousBusinessViolation("test.refused", "refused"))
				return nil
			},
		},
	})

	_, cErrs, err := service.runHooks(ownerContext(), it.ActionSetArchived, dmodel.DynamicFields{})

	assert.NoError(t, err)
	assert.Positive(t, cErrs.Count())
}

// An action name the lookup never heard of runs nothing rather than failing.
func TestServiceHooksTolerateMissingDefinition(t *testing.T) {
	service := newHookService(&stubRepository{}, map[string]it.DynamicActionDefinition{})

	params, cErrs, err := service.runHooks(ownerContext(), it.ActionExists, dmodel.DynamicFields{"id": "x"})

	assert.NoError(t, err)
	assert.Zero(t, cErrs.Count())
	assert.Equal(t, dmodel.DynamicFields{"id": "x"}, params)
}

// A service built without a lookup is legal, and runs no hook rather than panicking.
func TestServiceWithoutActionLookupRunsNoHook(t *testing.T) {
	schema := dmodel.DefineModel("test_lookupless_resource").Build()
	service := NewDynamicResourceService(NewServiceParam{Schema: schema}).(*DynamicResourceServiceImpl)

	params, cErrs, err := service.runHooks(ownerContext(), it.ActionCreate, dmodel.DynamicFields{"id": "x"})

	assert.NoError(t, err)
	assert.Zero(t, cErrs.Count())
	assert.Equal(t, dmodel.DynamicFields{"id": "x"}, params)
}

// The point of resolving the hooks through the engine rather than copying them at construction:
// a guard attached during a module's Init still reaches the service built before it.
func TestServiceSeesActionModifiedAfterConstruction(t *testing.T) {
	repo := &stubRepository{}
	schema := dmodel.DefineModel("test_late_modify").Build()

	// The registry's own order: engine first, then the service reading off it.
	engine := NewDynamicResourceEngine(NewEngineParam{Schema: schema, Repository: repo})
	service := NewDynamicResourceService(NewServiceParam{
		Schema: schema, Repository: repo, ActionLookup: engine.Action,
	}).(*DynamicResourceServiceImpl)
	engine.SetResourceService(service)

	assert.NoError(t, engine.DefineAction(it.DynamicActionDefinition{
		ActionName:  it.ActionSearch,
		MainProcess: noopProcess,
	}))
	called := false
	assert.NoError(t, engine.ModifyAction(it.DynamicActionDelta{
		ActionName: it.ActionSearch,
		ValidateExtra: func(_ corectx.Context, _ dmodel.DynamicFields, _ *dmodel.DynamicFields, _ *ft.ClientErrors) error {
			called = true
			return nil
		},
	}))

	_, _, err := service.runHooks(ownerContext(), it.ActionSearch, dmodel.DynamicFields{})

	assert.NoError(t, err)
	assert.True(t, called, "the service reads the definition the engine holds now, not a copy")
}

// Delete is the one operation whose guard needs a record the crud helper cannot supply, so the
// service resolves it from KeysToFetch and closes over it.
func TestServicePrepareResolvesFoundModelForDelete(t *testing.T) {
	repo := &stubRepository{
		record: dmodel.DynamicFields{"id": "rec-1", "status": "cancelled"},
		found:  true,
	}
	service := newHookService(repo, map[string]it.DynamicActionDefinition{
		it.ActionDelete: {
			ActionName: it.ActionDelete,
			KeysToFetch: func(params dmodel.DynamicFields) dmodel.DynamicFields {
				return dmodel.DynamicFields{"id": params["id"]}
			},
		},
	})

	definition := service.action(it.ActionDelete)
	params, foundModel, cErrs, err := service.prepare(
		ownerContext(), definition, dmodel.DynamicFields{"id": "rec-1"})

	assert.NoError(t, err)
	assert.Zero(t, cErrs.Count())
	assert.Equal(t, dmodel.DynamicFields{"id": "rec-1"}, repo.fetchedKeys)
	assert.NotNil(t, foundModel)
	assert.Equal(t, "cancelled", (*foundModel)["status"])

	// The adapter is what carries that record into the crud helper's key-only hook.
	seen := false
	hook := deleteValidateExtraFn(
		func(_ corectx.Context, gotParams dmodel.DynamicFields, gotFound *dmodel.DynamicFields, _ *ft.ClientErrors) error {
			seen = true
			assert.Equal(t, params, gotParams)
			assert.Equal(t, "cancelled", (*gotFound)["status"])
			return nil
		},
		params, foundModel,
	)
	cErrsOut := ft.ClientErrors{}
	assert.NoError(t, hook(ownerContext(), dmodel.DynamicFields{"id": "rec-1"}, &cErrsOut))
	assert.True(t, seen)
}

// crud.Create only adopts the field data of an entity it did not hand out, so the adapter has
// to return a new one when the hook produced a different map.
func TestBeforeValidationAdapterReturnsNewEntity(t *testing.T) {
	original := it.NewDynamicEntityFrom(dmodel.DynamicFields{"name": "raw"})
	adapted := beforeValidationFn(
		func(_ corectx.Context, _ dmodel.DynamicFields, _ *ft.ClientErrors) (dmodel.DynamicFields, error) {
			return dmodel.DynamicFields{"name": "sanitized"}, nil
		},
	)

	cErrs := ft.ClientErrors{}
	result, err := adapted(ownerContext(), original, &cErrs)

	assert.NoError(t, err)
	assert.NotSame(t, original, result)
	assert.Equal(t, "sanitized", result.GetFieldData()["name"])
}

// On a reported violation the hook's output is discarded, and the entity handed in comes back
// unchanged — never nil, which crud.Update would dereference.
func TestBeforeValidationAdapterKeepsEntityOnViolation(t *testing.T) {
	original := it.NewDynamicEntityFrom(dmodel.DynamicFields{"name": "raw"})
	adapted := beforeValidationFn(
		func(_ corectx.Context, _ dmodel.DynamicFields, vErrs *ft.ClientErrors) (dmodel.DynamicFields, error) {
			vErrs.Append(*ft.NewAnonymousBusinessViolation("test.bad", "bad"))
			return dmodel.DynamicFields{"name": "ignored"}, nil
		},
	)

	cErrs := ft.ClientErrors{}
	result, err := adapted(ownerContext(), original, &cErrs)

	assert.NoError(t, err)
	assert.Same(t, original, result)
	assert.Equal(t, "raw", result.GetFieldData()["name"])
}

func TestValidateExtraAdapters(t *testing.T) {
	t.Run("create passes no record", func(t *testing.T) {
		adapted := createValidateExtraFn(
			func(_ corectx.Context, params dmodel.DynamicFields, foundModel *dmodel.DynamicFields, _ *ft.ClientErrors) error {
				assert.Equal(t, "new", params["name"])
				assert.Nil(t, foundModel, "a create validates against nothing that already exists")
				return nil
			},
		)
		cErrs := ft.ClientErrors{}
		assert.NoError(t, adapted(ownerContext(),
			it.NewDynamicEntityFrom(dmodel.DynamicFields{"name": "new"}), &cErrs))
	})

	t.Run("update passes the stored record", func(t *testing.T) {
		adapted := updateValidateExtraFn(
			func(_ corectx.Context, params dmodel.DynamicFields, foundModel *dmodel.DynamicFields, _ *ft.ClientErrors) error {
				assert.Equal(t, "edited", params["name"])
				assert.Equal(t, "stored", (*foundModel)["name"])
				return nil
			},
		)
		cErrs := ft.ClientErrors{}
		assert.NoError(t, adapted(ownerContext(),
			it.NewDynamicEntityFrom(dmodel.DynamicFields{"name": "edited"}),
			it.NewDynamicEntityFrom(dmodel.DynamicFields{"name": "stored"}),
			&cErrs))
	})

	t.Run("a nil hook adapts to a nil slot", func(t *testing.T) {
		assert.Nil(t, createValidateExtraFn(nil))
		assert.Nil(t, updateValidateExtraFn(nil))
		assert.Nil(t, beforeValidationFn(nil))
		assert.Nil(t, afterValidationFn(nil))
		assert.Nil(t, deleteValidateExtraFn(nil, nil, nil))
		assert.Nil(t, deleteAfterValidationFn(nil, nil))
	})
}
