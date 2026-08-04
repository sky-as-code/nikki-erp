package engine

import (
	stdErr "errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	it "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// DefineBuiltinActions registers the CRUD actions every resource gets for free.
// The registry calls it right after constructing an engine.
//
// They declare no ParamSchema: the crud helpers the service delegates to already validate
// the params against the resource schema, inject the service fields, check the unique
// constraints and enforce the etag on update. Declaring a schema here would validate twice.
// A module that needs a pipeline-level schema on a built-in can still add one with
// ModifyAction; the second validation is idempotent on already-sanitized params.
func DefineBuiltinActions(engine it.DynamicResourceEngine) error {
	return stdErr.Join(
		engine.DefineAction(it.DynamicActionDefinition{
			ActionName:  it.ActionCreate,
			Permission:  it.PermissionCreate,
			MainProcess: processCreate,
		}),
		engine.DefineAction(it.DynamicActionDefinition{
			ActionName:  it.ActionUpdate,
			Permission:  it.PermissionUpdate,
			MainProcess: processUpdate,
		}),
		engine.DefineAction(it.DynamicActionDefinition{
			ActionName:  it.ActionDelete,
			Permission:  it.PermissionDelete,
			MainProcess: processDelete,
		}),
		engine.DefineAction(it.DynamicActionDefinition{
			ActionName:  it.ActionSetArchived,
			Permission:  it.PermissionSetArchived,
			MainProcess: processSetArchived,
		}),
		engine.DefineAction(it.DynamicActionDefinition{
			ActionName:  it.ActionGetById,
			Permission:  it.PermissionRead,
			MainProcess: processGetById,
		}),
		engine.DefineAction(it.DynamicActionDefinition{
			ActionName:  it.ActionGetByUnique,
			Permission:  it.PermissionRead,
			MainProcess: processGetByUnique,
		}),
		engine.DefineAction(it.DynamicActionDefinition{
			ActionName:  it.ActionSearch,
			Permission:  it.PermissionRead,
			MainProcess: processSearch,
		}),
		engine.DefineAction(it.DynamicActionDefinition{
			ActionName:  it.ActionExists,
			Permission:  it.PermissionRead,
			MainProcess: processExists,
		}),
		engine.DefineAction(it.DynamicActionDefinition{
			ActionName:  it.ActionGetSchema,
			Permission:  it.PermissionRead,
			MainProcess: processGetSchema,
		}),
	)
}

func processCreate(ctx corectx.Context, input it.ProcessInput) (*it.ActionResult, error) {
	result, err := input.ResourceService.Create(ctx, input.Params)
	return toActionResult(result, err)
}

func processUpdate(ctx corectx.Context, input it.ProcessInput) (*it.ActionResult, error) {
	result, err := input.ResourceService.Update(ctx, input.Params)
	return toActionResult(result, err)
}

func processDelete(ctx corectx.Context, input it.ProcessInput) (*it.ActionResult, error) {
	result, err := input.ResourceService.Delete(ctx, input.Params)
	return toActionResult(result, err)
}

func processSetArchived(ctx corectx.Context, input it.ProcessInput) (*it.ActionResult, error) {
	result, err := input.ResourceService.SetArchived(ctx, input.Params)
	return toActionResult(result, err)
}

func processGetById(ctx corectx.Context, input it.ProcessInput) (*it.ActionResult, error) {
	result, err := input.ResourceService.GetById(ctx, input.Params)
	return toActionResult(result, err)
}

func processGetByUnique(ctx corectx.Context, input it.ProcessInput) (*it.ActionResult, error) {
	result, err := input.ResourceService.GetOne(ctx, input.Params)
	return toActionResult(result, err)
}

func processSearch(ctx corectx.Context, input it.ProcessInput) (*it.ActionResult, error) {
	result, err := input.ResourceService.Search(ctx, input.Params)
	return toActionResult(result, err)
}

func processExists(ctx corectx.Context, input it.ProcessInput) (*it.ActionResult, error) {
	result, err := input.ResourceService.Exists(ctx, input.Params)
	return toActionResult(result, err)
}

// processGetSchema serves the resource schema in the simplified shape the clients cache.
func processGetSchema(_ corectx.Context, input it.ProcessInput) (*it.ActionResult, error) {
	return &it.ActionResult{
		Data:    input.ResourceService.Schema().ToSimplized(),
		HasData: true,
	}, nil
}
