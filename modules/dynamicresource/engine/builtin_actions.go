package engine

import (
	stdErr "errors"

	"github.com/sky-as-code/nikki-erp/common/util"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	it "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// DefineBuiltinActions registers the CRUD actions every resource gets for free.
// The registry calls it right after constructing an engine.
//
// An empty `only` defines all of them; a non-empty list defines just those, so that an
// engine can be created without the actions its resource has no business exposing. An
// action left out has no REST route and is refused by the resource service.
//
// They declare no ParamSchema: the crud helpers the service delegates to already validate
// the params against the resource schema, inject the service fields, check the unique
// constraints and enforce the etag on update. Declaring a schema here would validate twice.
// A module that needs a pipeline-level schema on a built-in can still add one with
// ModifyAction; the second validation is idempotent on already-sanitized params.
func DefineBuiltinActions(engine it.DynamicResourceEngine, only ...it.CrudAction) error {
	allowed := make(map[it.CrudAction]bool, len(only))
	for _, name := range only {
		allowed[name] = true
	}
	isWanted := func(name it.CrudAction) bool {
		return len(allowed) == 0 || allowed[name]
	}

	builtins := []struct {
		crudAction it.CrudAction
		definition it.DynamicActionDefinition
	}{
		{it.CrudActionCreate, it.DynamicActionDefinition{
			ActionName:  it.ActionCreate,
			ActionType:  it.ActionTypeCreate,
			RestPath:    "",
			Permission:  it.PermissionCreate,
			MainProcess: processCreate,
		}},
		{it.CrudActionUpdate, it.DynamicActionDefinition{
			ActionName:  it.ActionUpdate,
			ActionType:  it.ActionTypeUpdatePatch,
			RestPath:    ":id",
			Permission:  it.PermissionUpdate,
			MainProcess: processUpdate,
		}},
		{it.CrudActionDelete, it.DynamicActionDefinition{
			ActionName:  it.ActionDelete,
			ActionType:  it.ActionTypeDelete,
			RestPath:    ":id",
			Permission:  it.PermissionDelete,
			MainProcess: processDelete,
		}},
		// Archiving is a POST operation that is neither a create nor an update, hence Generic.
		{it.CrudActionSetArchived, it.DynamicActionDefinition{
			ActionName:  it.ActionSetArchived,
			ActionType:  it.ActionTypeGeneric,
			RestPath:    ":id/archived",
			Permission:  it.PermissionSetArchived,
			MainProcess: processSetArchived,
		}},
		{it.CrudActionGetById, it.DynamicActionDefinition{
			ActionName:  it.ActionGetById,
			ActionType:  it.ActionTypeRead,
			RestPath:    ":id",
			Permission:  it.PermissionRead,
			MainProcess: processGetById,
		}},
		// get_by_unique has no REST route: its unique keys vary per resource, so it is
		// reachable through ExecuteAction only.
		{it.CrudActionGetByUnique, it.DynamicActionDefinition{
			ActionName:  it.ActionGetByUnique,
			Permission:  it.PermissionRead,
			MainProcess: processGetByUnique,
		}},
		{it.CrudActionSearch, it.DynamicActionDefinition{
			ActionName:  it.ActionSearch,
			ActionType:  it.ActionTypeRead,
			RestPath:    "",
			Permission:  it.PermissionRead,
			MainProcess: processSearch,
		}},
		// exists carries a query in its body, so it is a POST that creates nothing: Generic.
		{it.CrudActionExists, it.DynamicActionDefinition{
			ActionName:  it.ActionExists,
			ActionType:  it.ActionTypeGeneric,
			RestPath:    "exists",
			Permission:  it.PermissionRead,
			MainProcess: processExists,
		}},
		// get_schema describes the resource's shape, which is the same in every org, and
		// touches no row - so it is one of the few actions that legitimately opts out.
		{it.CrudActionGetSchema, it.DynamicActionDefinition{
			ActionName:  it.ActionGetSchema,
			ActionType:  it.ActionTypeRead,
			RestPath:    "meta/schema",
			Permission:  it.PermissionRead,
			IsOrgScoped: util.ToPtr(false),
			MainProcess: processGetSchema,
		}},
	}

	errs := make([]error, 0, len(builtins)+1)
	for _, builtin := range builtins {
		if !isWanted(builtin.crudAction) {
			continue
		}
		errs = append(errs, engine.DefineAction(builtin.definition))
	}
	// Not a CRUD built-in: it needs the engine itself, to reach the computed-field function
	// registry, and is always defined. See computed_rest.go.
	errs = append(errs, defineComputeFieldAction(engine))
	return stdErr.Join(errs...)
}

// WithdrawOrgScoping opts every action currently defined on the engine out of org scoping.
//
// It exists for a resource whose org_id is *optional*, where NULL means "global" or
// "domain-scoped" rather than "belongs to no org" - iam_role and iam_entitlement are the two in
// the tree. Requiring ?org_id= on those would make every domain-scoped row unreachable, which
// is a silent under-grant rather than a visible error.
//
// Call it after the module has defined its own actions, so that those are covered too.
// It is deliberately all-or-nothing: a resource that needs scoping on some actions and not
// others should say so per action, where the reason can be written down next to the exception.
func WithdrawOrgScoping(engine it.DynamicResourceEngine) error {
	errs := make([]error, 0)
	for _, name := range engine.ActionNames() {
		errs = append(errs, engine.ModifyAction(it.DynamicActionDelta{
			ActionName:  name,
			IsOrgScoped: util.ToPtr(false),
		}))
	}
	return stdErr.Join(errs...)
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
