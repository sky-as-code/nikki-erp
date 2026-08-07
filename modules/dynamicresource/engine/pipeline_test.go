package engine

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
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

// Both validation hooks are gated on ParamSchema being declared.
func TestExecuteActionSkipsHooksWithoutParamSchema(t *testing.T) {
	engine := newPipelineEngine(&stubRepository{})
	beforeCalled, afterCalled := false, false

	assert.NoError(t, engine.DefineAction(it.DynamicActionDefinition{
		ActionName: "unvalidated",
		BeforeValidation: func(_ corectx.Context, params dmodel.DynamicFields, _ *ft.ClientErrors) (dmodel.DynamicFields, error) {
			beforeCalled = true
			return params, nil
		},
		AfterValidationSuccess: func(_ corectx.Context, _ dmodel.DynamicFields) error {
			afterCalled = true
			return nil
		},
		MainProcess: noopProcess,
	}))

	_, err := engine.ExecuteAction(ownerContext(), "unvalidated", nil)

	assert.NoError(t, err)
	assert.False(t, beforeCalled, "BeforeValidation runs only with a ParamSchema")
	assert.False(t, afterCalled, "AfterValidationSuccess runs only with a ParamSchema")
}

// ValidateExtra is the one hook that runs whether or not a ParamSchema is declared.
func TestExecuteActionRunsValidateExtraWithoutParamSchema(t *testing.T) {
	engine := newPipelineEngine(&stubRepository{})
	called := false

	assert.NoError(t, engine.DefineAction(it.DynamicActionDefinition{
		ActionName: "extra_only",
		ValidateExtra: func(_ corectx.Context, _ dmodel.DynamicFields, foundModel *dmodel.DynamicFields, _ *ft.ClientErrors) error {
			called = true
			assert.Nil(t, foundModel, "foundModel is nil when KeysToFetch is absent")
			return nil
		},
		MainProcess: noopProcess,
	}))

	_, err := engine.ExecuteAction(ownerContext(), "extra_only", nil)

	assert.NoError(t, err)
	assert.True(t, called)
}

func TestExecuteActionFetchesKeysBeforeValidateExtra(t *testing.T) {
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
		ValidateExtra: func(_ corectx.Context, _ dmodel.DynamicFields, foundModel *dmodel.DynamicFields, _ *ft.ClientErrors) error {
			assert.NotNil(t, foundModel, "foundModel is fetched when KeysToFetch is provided")
			assert.Equal(t, "stored", (*foundModel)["name"])
			return nil
		},
		MainProcess: noopProcess,
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
		ValidateExtra: func(_ corectx.Context, _ dmodel.DynamicFields, _ *dmodel.DynamicFields, _ *ft.ClientErrors) error {
			called = true
			return nil
		},
		MainProcess: noopProcess,
	}))

	result, err := engine.ExecuteAction(ownerContext(), "needs_missing", nil)

	assert.NoError(t, err)
	assert.Positive(t, result.ClientErrors.Count())
	assert.False(t, called, "the flow stops before ValidateExtra when the record is missing")
}
