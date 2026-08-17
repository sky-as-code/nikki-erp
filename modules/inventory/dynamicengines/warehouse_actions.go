package dynamicengines

import (
	stdErr "errors"

	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/services"
)

// The lifecycle operations of Warehouse and Inventory Location, and the putaway suggestion.
//
// Only these two resources have a suspend and resume, because only they have an operational state
// that is independent of archiving. Supply Relation, Storage Category and Putaway Rule are either
// part of the working set or archived, so archiving is their whole lifecycle and there is nothing
// here for them beyond the built-in actions.
//
// There is deliberately no activate or deactivate anywhere: suspension is reversible and temporary,
// archiving is how something leaves the working set, and a third pair of verbs meaning one of
// those would only invite them to drift apart.

const (
	ActionSuspend                = "suspend"
	ActionResume                 = "resume"
	ActionMoveLocation           = "move"
	ActionConfigureIncomingFlow  = "configure_incoming_flow"
	ActionConfigureOutgoingFlow  = "configure_outgoing_flow"
	ActionSuggestPutawayLocation = "suggest_location"
)

const (
	PermissionSuspend                = "suspend"
	PermissionResume                 = "resume"
	PermissionMoveLocation           = "move"
	PermissionConfigureIncomingFlow  = "configure_incoming_flow"
	PermissionConfigureOutgoingFlow  = "configure_outgoing_flow"
	PermissionSuggestPutawayLocation = "suggest_location"
)

const (
	paramWarehouseId      = "id"
	paramLocationId       = "id"
	paramFlow             = "flow"
	paramNewParentId      = "parent_location_id"
	paramArrivalLocation  = "arrival_location_id"
	paramPutawayWarehouse = "warehouse_id"
	paramProductId        = "product_id"
	paramProductCategory  = "product_category_id"
	paramPackageType      = "package_type_id"
)

// defineWarehouseActions adds suspend, resume and the two flow reconfigurations.
//
// The flow actions reach the application service rather than the domain one: each writes the
// warehouse and provisions its locations, and those have to happen together or not at all.
func defineWarehouseActions(engine drif.DynamicResourceEngine) error {
	return stdErr.Join(
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionSuspend,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/suspend",
			Permission:  PermissionSuspend,
			MainProcess: processSuspendWarehouse,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionResume,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/resume",
			Permission:  PermissionResume,
			MainProcess: processResumeWarehouse,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionConfigureIncomingFlow,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/configure_incoming_flow",
			Permission:  PermissionConfigureIncomingFlow,
			MainProcess: processConfigureIncomingFlow,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionConfigureOutgoingFlow,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/configure_outgoing_flow",
			Permission:  PermissionConfigureOutgoingFlow,
			MainProcess: processConfigureOutgoingFlow,
		}),
	)
}

// defineInventoryLocationActions adds suspend, resume and move.
//
// Suspend is allowed while the location still holds stock — locking a damaged rack that holds
// goods is the point — whereas archiving one that does is refused by the built-in set_archived.
// The asymmetry is deliberate and is enforced in the domain service.
func defineInventoryLocationActions(engine drif.DynamicResourceEngine) error {
	return stdErr.Join(
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionSuspend,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/suspend",
			Permission:  PermissionSuspend,
			MainProcess: processSuspendLocation,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionResume,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/resume",
			Permission:  PermissionResume,
			MainProcess: processResumeLocation,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionMoveLocation,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/move",
			Permission:  PermissionMoveLocation,
			MainProcess: processMoveLocation,
		}),
	)
}

// definePutawayRuleActions adds the suggestion lookup.
//
// It has no id in its path because it asks a question of the whole rule set rather than acting on
// one rule. It changes nothing: the answer is a destination and the rule that produced it.
func definePutawayRuleActions(engine drif.DynamicResourceEngine) error {
	return engine.DefineAction(drif.DynamicActionDefinition{
		ActionName:  ActionSuggestPutawayLocation,
		ActionType:  drif.ActionTypeGeneric,
		RestPath:    "suggest_location",
		Permission:  PermissionSuggestPutawayLocation,
		MainProcess: processSuggestPutawayLocation,
	})
}

// warehouseServiceOf reaches the derived warehouse service installed during Init.
//
// A failed assertion means Init did not install it, which is a wiring bug rather than a request
// problem, so it is a plain error and not a client one.
func warehouseServiceOf(input drif.ProcessInput) (*services.WarehouseDomainServiceImpl, error) {
	service, ok := input.ResourceService.(*services.WarehouseDomainServiceImpl)
	if !ok {
		return nil, errors.New(
			"the warehouse engine is not running the derived warehouse service; " +
				"InventoryModule.Init must install it with SetResourceService")
	}
	return service, nil
}

// locationServiceOf reaches the derived location service installed during Init.
func locationServiceOf(input drif.ProcessInput) (*services.InventoryLocationDomainServiceImpl, error) {
	service, ok := input.ResourceService.(*services.InventoryLocationDomainServiceImpl)
	if !ok {
		return nil, errors.New(
			"the inventory location engine is not running the derived location service; " +
				"InventoryModule.Init must install it with SetResourceService")
	}
	return service, nil
}

func processSuspendWarehouse(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := warehouseServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.Suspend(ctx, readStringField(input.Params, paramWarehouseId))
	return toMutateActionResult(result, err)
}

func processResumeWarehouse(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := warehouseServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.Resume(ctx, readStringField(input.Params, paramWarehouseId))
	return toMutateActionResult(result, err)
}

func processSuspendLocation(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := locationServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.Suspend(ctx, readStringField(input.Params, paramLocationId))
	return toMutateActionResult(result, err)
}

func processResumeLocation(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := locationServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.Resume(ctx, readStringField(input.Params, paramLocationId))
	return toMutateActionResult(result, err)
}

// processMoveLocation re-parents a location. An empty parent makes it a root, which is why the
// parameter is read without being required.
func processMoveLocation(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := locationServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.Move(ctx,
		readStringField(input.Params, paramLocationId),
		readStringField(input.Params, paramNewParentId))
	return toMutateActionResult(result, err)
}
