package engine

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	it "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// ownerContext passes every permission check, so that tests that are not about
// permissions do not have to build entitlements.
func ownerContext() corectx.Context {
	ctx := corectx.NewRequestContext(context.Background())
	ctx.SetPermissions(corectx.ContextPermissions{IsOwner: true})
	return ctx
}

// deniedContext fails every permission check.
func deniedContext() corectx.Context {
	ctx := corectx.NewRequestContext(context.Background())
	ctx.SetPermissions(corectx.ContextPermissions{})
	return ctx
}

// stubRepository records the keys it was asked to fetch and returns a canned record.
type stubRepository struct {
	it.DynamicResourceRepository

	fetchedKeys dmodel.DynamicFields
	record      dmodel.DynamicFields
	found       bool
}

func (this *stubRepository) FindByKeys(
	_ corectx.Context, keys dmodel.DynamicFields,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	this.fetchedKeys = keys
	return &dyn.OpResult[dmodel.DynamicFields]{
		Data:    this.record,
		HasData: this.found,
	}, nil
}

// newPipelineEngine builds an engine whose repository is a stub, so that the pipeline
// can be exercised without a database.
func newPipelineEngine(repo it.DynamicResourceRepository) *DynamicResourceEngineImpl {
	schema := dmodel.DefineModel("test_resource").Build()
	engine := NewDynamicResourceEngine(NewEngineParam{
		Schema:     schema,
		Repository: repo,
	}).(*DynamicResourceEngineImpl)
	return engine
}

func TestExecuteActionRunsMainProcess(t *testing.T) {
	engine := newPipelineEngine(&stubRepository{})
	var seen dmodel.DynamicFields

	assert.NoError(t, engine.DefineAction(it.DynamicActionDefinition{
		ActionName: "do_it",
		MainProcess: func(_ corectx.Context, input it.ProcessInput) (*it.ActionResult, error) {
			seen = input.Params
			return &it.ActionResult{Data: "done", HasData: true}, nil
		},
	}))

	result, err := engine.ExecuteAction(ownerContext(), "do_it", dmodel.DynamicFields{"name": "abc"})

	assert.NoError(t, err)
	assert.Equal(t, "done", result.Data)
	assert.Equal(t, "abc", seen["name"])
}

// An action declaring KeysToFetch has its record read by the pipeline. Handing that record to
// MainProcess is what saves it re-reading a row the pipeline already read.
func TestExecuteActionPassesFoundModelToMainProcess(t *testing.T) {
	repo := &stubRepository{
		record: dmodel.DynamicFields{"id": "01ABC", "name": "stored"},
		found:  true,
	}
	engine := newPipelineEngine(repo)
	var seen *dmodel.DynamicFields

	assert.NoError(t, engine.DefineAction(it.DynamicActionDefinition{
		ActionName: "with_keys",
		KeysToFetch: func(params dmodel.DynamicFields) dmodel.DynamicFields {
			return dmodel.DynamicFields{"id": params["id"]}
		},
		MainProcess: func(_ corectx.Context, input it.ProcessInput) (*it.ActionResult, error) {
			seen = input.FoundModel
			return &it.ActionResult{HasData: true}, nil
		},
	}))

	_, err := engine.ExecuteAction(ownerContext(), "with_keys", dmodel.DynamicFields{"id": "01ABC"})

	assert.NoError(t, err)
	assert.NotNil(t, seen, "the pipeline already fetched this record")
	assert.Equal(t, "stored", (*seen)["name"])
}

// Without KeysToFetch there is nothing to hand over, and MainProcess must be able to tell that
// apart from a record that happens to be empty.
func TestExecuteActionLeavesFoundModelNilWithoutKeysToFetch(t *testing.T) {
	engine := newPipelineEngine(&stubRepository{})
	seen := &dmodel.DynamicFields{}

	assert.NoError(t, engine.DefineAction(it.DynamicActionDefinition{
		ActionName: "no_keys",
		MainProcess: func(_ corectx.Context, input it.ProcessInput) (*it.ActionResult, error) {
			seen = input.FoundModel
			return &it.ActionResult{HasData: true}, nil
		},
	}))

	_, err := engine.ExecuteAction(ownerContext(), "no_keys", nil)

	assert.NoError(t, err)
	assert.Nil(t, seen)
}

func TestExecuteActionRejectsUnknownAction(t *testing.T) {
	engine := newPipelineEngine(&stubRepository{})

	_, err := engine.ExecuteAction(ownerContext(), "ghost", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not defined")
}

// A permission denial must surface as client errors so the REST layer answers 400,
// never as a Go error, which would become a 500.
func TestExecuteActionDeniesWithoutPermission(t *testing.T) {
	engine := newPipelineEngine(&stubRepository{})
	called := false

	assert.NoError(t, engine.DefineAction(it.DynamicActionDefinition{
		ActionName: "guarded",
		Permission: it.PermissionUpdate,
		MainProcess: func(_ corectx.Context, _ it.ProcessInput) (*it.ActionResult, error) {
			called = true
			return &it.ActionResult{HasData: true}, nil
		},
	}))

	result, err := engine.ExecuteAction(deniedContext(), "guarded", nil)

	assert.NoError(t, err)
	assert.Positive(t, result.ClientErrors.Count())
	assert.False(t, called, "main process must not run when permission is denied")
}

func TestExecuteActionSkipsPermissionCheckWhenEmpty(t *testing.T) {
	engine := newPipelineEngine(&stubRepository{})
	assert.NoError(t, engine.DefineAction(it.DynamicActionDefinition{
		ActionName:  "open",
		Permission:  "",
		MainProcess: noopProcess,
	}))

	result, err := engine.ExecuteAction(deniedContext(), "open", nil)

	assert.NoError(t, err)
	assert.Zero(t, result.ClientErrors.Count())
}

// The pipeline no longer runs the validator hooks at all: they are executed by the crud
// helper the resource service delegates to. The two tests that asserted pipeline-level
// BeforeValidation / AfterValidationSuccess / ValidateExtra behaviour were deleted with
// that step; the hooks are covered in service_test.go instead.

// KeysToFetch survives the hook removal because it still fills ProcessInput.FoundModel.
func TestExecuteActionFetchesKeysForMainProcess(t *testing.T) {
	repo := &stubRepository{
		record: dmodel.DynamicFields{"id": "rec-1", "name": "stored"},
		found:  true,
	}
	engine := newPipelineEngine(repo)

	assert.NoError(t, engine.DefineAction(it.DynamicActionDefinition{
		ActionName: "needs_record",
		KeysToFetch: func(params dmodel.DynamicFields) dmodel.DynamicFields {
			return dmodel.DynamicFields{"id": params["id"]}
		},
		MainProcess: func(_ corectx.Context, input it.ProcessInput) (*it.ActionResult, error) {
			assert.NotNil(t, input.FoundModel, "FoundModel is fetched when KeysToFetch is provided")
			assert.Equal(t, "stored", (*input.FoundModel)["name"])
			return &it.ActionResult{}, nil
		},
	}))

	_, err := engine.ExecuteAction(ownerContext(), "needs_record", dmodel.DynamicFields{"id": "rec-1"})

	assert.NoError(t, err)
	assert.Equal(t, dmodel.DynamicFields{"id": "rec-1"}, repo.fetchedKeys)
}

// A record the action wants to load but cannot find is a client error, not a Go error.
func TestExecuteActionReportsMissingRecordAsClientError(t *testing.T) {
	engine := newPipelineEngine(&stubRepository{found: false})
	called := false

	assert.NoError(t, engine.DefineAction(it.DynamicActionDefinition{
		ActionName: "needs_missing",
		KeysToFetch: func(params dmodel.DynamicFields) dmodel.DynamicFields {
			return dmodel.DynamicFields{"id": "nope"}
		},
		MainProcess: func(_ corectx.Context, _ it.ProcessInput) (*it.ActionResult, error) {
			called = true
			return &it.ActionResult{}, nil
		},
	}))

	result, err := engine.ExecuteAction(ownerContext(), "needs_missing", nil)

	assert.NoError(t, err)
	assert.Positive(t, result.ClientErrors.Count())
	assert.False(t, called, "the flow stops before MainProcess when the record is missing")
}
