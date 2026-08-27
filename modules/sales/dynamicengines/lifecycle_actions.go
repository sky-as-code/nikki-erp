package dynamicengines

import (
	stdErr "errors"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
)

// Permission codes for the lifecycle operations. They match the action codes seeded in
// 1007002_sales_iam.sql — a code that drifts from its seed denies every request, and nothing in
// the resulting 403 points at the seed as the cause.
const (
	PermissionSuspend  = "suspend"
	PermissionActivate = "activate"
)

// Action names, namespaced by resource in the same style as the built-ins.
const (
	ActionSuspend   = "suspend"
	ActionActivate  = "activate"
	ActionArchive   = "archive"
	ActionUnarchive = "unarchive"
	ActionResolve   = "resolve"
)

// Param names the lifecycle actions read from the request.
const (
	paramId   = "id"
	paramCode = "code"
)

// defineSalesChannelActions adds the channel lifecycle.
//
// Archive rides on the built-in set_archived permission rather than taking one of its own: it is
// the same power the resource's own archive flag represents, and splitting them would let a role
// archive a channel through one route while being refused on the other.
func defineSalesChannelActions(engine drif.DynamicResourceEngine) error {
	return stdErr.Join(
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionSuspend,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/suspend",
			Permission:  PermissionSuspend,
			MainProcess: processChannelSuspend,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionActivate,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/activate",
			Permission:  PermissionActivate,
			MainProcess: processChannelActivate,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionArchive,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/archive",
			Permission:  drif.PermissionSetArchived,
			MainProcess: processChannelArchive,
		}),
		// Resolving a code to an id is a read of one row by a different key, so it reuses the read
		// permission. A separate code would let a role be granted "may resolve" while unable to
		// read what it resolved.
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionResolve,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    "resolve",
			Permission:  drif.PermissionRead,
			MainProcess: processChannelResolve,
		}),
	)
}

// defineSalesPointActions adds the sales point lifecycle.
//
// Unarchive is a separate action from activate but shares the set_archived permission with archive:
// they are the same power applied in either direction, and a role that may retire a kiosk should be
// able to undo its own mistake.
func defineSalesPointActions(engine drif.DynamicResourceEngine) error {
	return stdErr.Join(
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionSuspend,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/suspend",
			Permission:  PermissionSuspend,
			MainProcess: processPointSuspend,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionActivate,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/activate",
			Permission:  PermissionActivate,
			MainProcess: processPointActivate,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionArchive,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/archive",
			Permission:  drif.PermissionSetArchived,
			MainProcess: processPointArchive,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionUnarchive,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/unarchive",
			Permission:  drif.PermissionSetArchived,
			MainProcess: processPointUnarchive,
		}),
	)
}

func processChannelSuspend(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := channelServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.Suspend(ctx, readStringParam(input.Params, paramId))
	return toMutateActionResult(result, err)
}

func processChannelActivate(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := channelServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.Activate(ctx, readStringParam(input.Params, paramId))
	return toMutateActionResult(result, err)
}

func processChannelArchive(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := channelServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.Archive(ctx, readStringParam(input.Params, paramId))
	return toMutateActionResult(result, err)
}

func processChannelResolve(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := channelServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.ResolveByCode(ctx, readStringParam(input.Params, paramCode))
	return toMutateActionResult(result, err)
}

func processPointSuspend(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := pointServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.Suspend(ctx, readStringParam(input.Params, paramId))
	return toMutateActionResult(result, err)
}

func processPointActivate(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := pointServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.Activate(ctx, readStringParam(input.Params, paramId))
	return toMutateActionResult(result, err)
}

func processPointArchive(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := pointServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.Archive(ctx, readStringParam(input.Params, paramId))
	return toMutateActionResult(result, err)
}

func processPointUnarchive(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := pointServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.Unarchive(ctx, readStringParam(input.Params, paramId))
	return toMutateActionResult(result, err)
}

// channelServiceOf reaches the derived service the module installed during Init.
//
// The type assertion is what makes the lifecycle operations reachable: the engine hands the action
// its service as the base interface, and only the derived type carries Suspend, Activate and
// Archive. A failed assertion means Init did not install it, which is a wiring bug rather than a
// request problem — so it returns a Go error and answers 500.
func channelServiceOf(input drif.ProcessInput) (*services.SalesChannelDomainServiceImpl, error) {
	service, ok := input.ResourceService.(*services.SalesChannelDomainServiceImpl)
	if !ok {
		return nil, errors.New(
			"the sales channel engine is not running the derived channel service; " +
				"SalesModule.Init must install it with SetResourceService")
	}
	return service, nil
}

func pointServiceOf(input drif.ProcessInput) (*services.SalesPointDomainServiceImpl, error) {
	service, ok := input.ResourceService.(*services.SalesPointDomainServiceImpl)
	if !ok {
		return nil, errors.New(
			"the sales point engine is not running the derived sales point service; " +
				"SalesModule.Init must install it with SetResourceService")
	}
	return service, nil
}

// readStringParam reads one string from the flattened params the pipeline already bound.
//
// Never type-assert directly: a repository or JSON round-trip can hand back a different concrete
// type than expected, and a bare assertion panics the request. A malformed value degrades to the
// empty string, which the services reject by name.
func readStringParam(params dmodel.DynamicFields, field string) string {
	value, ok := params[field]
	if !ok || value == nil {
		return ""
	}
	if typed, ok := value.(string); ok {
		return typed
	}
	if typed, ok := value.(*string); ok && typed != nil {
		return *typed
	}
	return ""
}

// toMutateActionResult widens a mutation result into the engine's generic action result.
//
// The engine's own equivalent is package-private, so this is a local copy rather than an
// unnecessary duplicate. The ClientErrors must survive: a refused operation reports its reason
// through them and not through err, which is what makes the REST layer answer 400 rather than 500.
func toMutateActionResult(
	result *dyn.OpResult[dyn.MutateResultData], err error,
) (*drif.ActionResult, error) {
	if err != nil {
		return nil, err
	}
	out := &drif.ActionResult{
		ClientErrors: result.ClientErrors,
		HasData:      result.HasData,
	}
	if result.HasData {
		out.Data = result.Data
	}
	return out, nil
}
