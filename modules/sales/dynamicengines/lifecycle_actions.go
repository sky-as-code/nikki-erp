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

// Permission codes for the lifecycle operations. They must match the action codes seeded in
// 1007002_sales_iam.sql; a code that drifts denies every request with no hint in the 403.
const (
	PermissionSuspend  = "suspend"
	PermissionActivate = "activate"
)

const (
	ActionSuspend   = "suspend"
	ActionActivate  = "activate"
	ActionArchive   = "archive"
	ActionUnarchive = "unarchive"
	ActionResolve   = "resolve"
)

const (
	paramId   = "id"
	paramCode = "code"
)

// Archive rides on the built-in set_archived permission: splitting them would let a role archive a
// channel through one route while being refused on the other.
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
		// Resolve is a read by a different key, so it reuses the read permission.
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionResolve,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    "resolve",
			Permission:  drif.PermissionRead,
			MainProcess: processChannelResolve,
		}),
	)
}

// Unarchive is a separate action from activate but shares archive's set_archived permission: it is
// the same power in reverse, so whoever may retire a kiosk can undo it.
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

// channelServiceOf asserts to the derived type because the engine hands the action its service as
// the base interface, and only the derived type carries Suspend, Activate and Archive. A failed
// assertion is a wiring bug, so it answers a Go error (500).
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

// readStringParam avoids a bare type assertion: a repository or JSON round-trip can hand back a
// different concrete type, and a bare assertion panics the request. A malformed value degrades to
// the empty string, which the services reject by name.
// readBoolParam reads a flag, treating anything absent or not a bool as false.
//
// A checkbox that did not arrive is an unticked one, and defaulting to false is the conservative
// reading for every flag that grants rather than restricts.
func readBoolParam(params dmodel.DynamicFields, field string) bool {
	value, ok := params[field]
	if !ok || value == nil {
		return false
	}
	if typed, ok := value.(bool); ok {
		return typed
	}
	if typed, ok := value.(*bool); ok && typed != nil {
		return *typed
	}
	return false
}

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

// toMutateActionResult is a local copy because the engine's equivalent is package-private. The
// ClientErrors must survive: a refused operation reports its reason through them and not through
// err, which is what makes the REST layer answer 400 rather than 500.
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
