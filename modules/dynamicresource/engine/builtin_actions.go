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
			ActionType:  it.ActionTypeCreate,
			RestPath:    "",
			Permission:  it.PermissionCreate,
			MainProcess: processCreate,
		}),
		engine.DefineAction(it.DynamicActionDefinition{
			ActionName:  it.ActionUpdate,
			ActionType:  it.ActionTypeUpdatePatch,
			RestPath:    ":id",
			Permission:  it.PermissionUpdate,
			MainProcess: processUpdate,
		}),
		engine.DefineAction(it.DynamicActionDefinition{
			ActionName:  it.ActionDelete,
			ActionType:  it.ActionTypeDelete,
			RestPath:    ":id",
			Permission:  it.PermissionDelete,
			MainProcess: processDelete,
		}),
		// Archiving is a POST operation that is neither a create nor an update, hence Generic.
		engine.DefineAction(it.DynamicActionDefinition{
			ActionName:  it.ActionSetArchived,
			ActionType:  it.ActionTypeGeneric,
			RestPath:    ":id/archived",
			Permission:  it.PermissionSetArchived,
			MainProcess: processSetArchived,
		}),
		engine.DefineAction(it.DynamicActionDefinition{
			ActionName:  it.ActionGetById,
			ActionType:  it.ActionTypeRead,
			RestPath:    ":id",
			Permission:  it.PermissionRead,
			MainProcess: processGetById,
		}),
		// get_by_unique has no REST route: its unique keys vary per resource, so it is
		// reachable through ExecuteAction only.
		engine.DefineAction(it.DynamicActionDefinition{
			ActionName:  it.ActionGetByUnique,
			Permission:  it.PermissionRead,
			MainProcess: processGetByUnique,
		}),
		engine.DefineAction(it.DynamicActionDefinition{
			ActionName:  it.ActionSearch,
			ActionType:  it.ActionTypeRead,
			RestPath:    "",
			Permission:  it.PermissionRead,
			MainProcess: processSearch,
		}),
		// exists carries a query in its body, so it is a POST that creates nothing: Generic.
		engine.DefineAction(it.DynamicActionDefinition{
			ActionName:  it.ActionExists,
			ActionType:  it.ActionTypeGeneric,
			RestPath:    "exists",
			Permission:  it.PermissionRead,
			MainProcess: processExists,
		}),
		engine.DefineAction(it.DynamicActionDefinition{
			ActionName:  it.ActionGetSchema,
			ActionType:  it.ActionTypeRead,
			RestPath:    "meta/schema",
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
	if err != nil {
		return nil, err
	}
	// A search always has a payload, even when it matched nothing: the empty item list
	// still carries total/page/size. Unlike the single-record actions, HasData=false here
	// means "no rows", not "no result", so the data is attached either way.
	return &it.ActionResult{
		ClientErrors: result.ClientErrors,
		HasData:      result.HasData,
		Data:         result.Data,
	}, nil
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
