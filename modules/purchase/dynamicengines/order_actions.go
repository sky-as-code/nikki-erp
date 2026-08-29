package dynamicengines

import (
	stdErr "errors"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/services"
)

// The purchase order's lifecycle operations, exposed as engine actions so the engine does the
// permission check, param binding and response shaping. Each carries its own permission rather than
// reusing `update`, because confirming and approving are materially different powers from editing a
// description.

// Permission codes for the order operations. They match the action codes seeded in
// 0007002_purchase_iam.sql — a code that drifts from its seed denies every request.
const (
	PermissionConfirm     = "confirm"
	PermissionApprove     = "approve"
	PermissionCancel      = "cancel"
	PermissionSend        = "send"
	PermissionLock        = "lock"
	PermissionUnlock      = "unlock"
	PermissionAcknowledge = "acknowledge"
	PermissionMerge       = "merge"

	// PermissionReprice is its own power: repricing rewrites what the company will pay across a
	// whole order, but it commits nothing, so it is neither `update` nor `confirm`.
	PermissionReprice = "reprice"
)

// Action names, namespaced by resource in the same style as the built-ins.
const (
	ActionConfirm     = "confirm"
	ActionApprove     = "approve"
	ActionCancel      = "cancel"
	ActionSend        = "send"
	ActionLock        = "lock"
	ActionUnlock      = "unlock"
	ActionAcknowledge = "acknowledge"
	ActionDuplicate   = "duplicate"
	ActionReprice     = "reprice"

	ActionMerge               = "merge"
	ActionCreateAlternative   = "create_alternative"
	ActionCompareAlternatives = "compare_alternatives"
)

// Param names the order actions read from the request.
const (
	paramOrderId = "id"
	paramReason  = "reason"
	// paramAlternativeChoice carries the answer to the open-alternatives warning.
	paramAlternativeChoice = "alternative_choice"
	paramAlternativeVendor = "vendor_id"
	paramMergeOrderIds     = "order_ids"
)

func defineOrderActions(engine drif.DynamicResourceEngine) error {
	if err := defineOrderDeleteGuard(engine); err != nil {
		return err
	}
	return stdErr.Join(
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionConfirm,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/confirm",
			Permission:  PermissionConfirm,
			MainProcess: processOrderConfirm,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionApprove,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/approve",
			Permission:  PermissionApprove,
			MainProcess: processOrderApprove,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionCancel,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/cancel",
			Permission:  PermissionCancel,
			MainProcess: processOrderCancel,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionSend,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/send",
			Permission:  PermissionSend,
			MainProcess: processOrderSend,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionLock,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/lock",
			Permission:  PermissionLock,
			MainProcess: processOrderLock,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionUnlock,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/unlock",
			Permission:  PermissionUnlock,
			MainProcess: processOrderUnlock,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionAcknowledge,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/acknowledge",
			Permission:  PermissionAcknowledge,
			MainProcess: processOrderAcknowledge,
		}),
		// Merge is collection-level: it acts on several orders at once, so it takes their ids in
		// the body rather than one in the path.
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionMerge,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    "merge",
			Permission:  PermissionMerge,
			MainProcess: processOrderMerge,
		}),
		// Raising an alternative creates an order, so it carries the create permission.
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionCreateAlternative,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/create_alternative",
			Permission:  drif.PermissionCreate,
			MainProcess: processOrderCreateAlternative,
		}),
		// Comparing changes nothing, so it needs only read.
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionCompareAlternatives,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/compare_alternatives",
			Permission:  drif.PermissionRead,
			MainProcess: processOrderCompareAlternatives,
		}),
		// Reprice takes only the order id: which prices apply is the resolver's answer, and letting a
		// caller pass one in would make this an override with extra steps.
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionReprice,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/reprice",
			Permission:  PermissionReprice,
			MainProcess: processOrderReprice,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionDuplicate,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/duplicate",
			Permission:  drif.PermissionCreate,
			MainProcess: processOrderDuplicate,
		}),
	)
}

func processOrderConfirm(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := orderServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.Confirm(
		ctx, readOrderId(input), readStringParam(input.Params, paramAlternativeChoice))
	return toMutateActionResult(result, err)
}

func processOrderApprove(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := orderServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.Approve(ctx, readOrderId(input))
	return toMutateActionResult(result, err)
}

func processOrderCancel(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := orderServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.Cancel(ctx, readOrderId(input), readStringParam(input.Params, paramReason))
	return toMutateActionResult(result, err)
}

func processOrderSend(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := orderServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.Send(ctx, readOrderId(input))
	return toMutateActionResult(result, err)
}

func processOrderLock(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := orderServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.Lock(ctx, readOrderId(input))
	return toMutateActionResult(result, err)
}

// The reason is validated in the service rather than through a ParamSchema: an action whose params
// mix resource fields with action-specific input is not described by any single schema.
func processOrderUnlock(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := orderServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.Unlock(ctx, readOrderId(input), readStringParam(input.Params, paramReason))
	return toMutateActionResult(result, err)
}

func processOrderAcknowledge(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := orderServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.Acknowledge(ctx, readOrderId(input))
	return toMutateActionResult(result, err)
}

// Duplicate returns the new order rather than an affected count so the caller has its id.
func processOrderDuplicate(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := orderServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.Duplicate(ctx, readOrderId(input))
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

// Reprice returns which lines moved rather than an affected count.
func processOrderReprice(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := orderServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.Reprice(ctx, readOrderId(input))
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

func processOrderMerge(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := orderServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.MergeOrders(ctx, readStringList(input.Params, paramMergeOrderIds))
	return toMutateActionResult(result, err)
}

func processOrderCreateAlternative(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := orderServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.CreateAlternative(
		ctx, readOrderId(input), readStringParam(input.Params, paramAlternativeVendor))
	if err != nil {
		return nil, err
	}
	out := &drif.ActionResult{ClientErrors: result.ClientErrors, HasData: result.HasData}
	if result.HasData {
		out.Data = result.Data
	}
	return out, nil
}

func processOrderCompareAlternatives(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	comparison, err := services.CompareAlternatives(ctx, readOrderId(input))
	if err != nil {
		return nil, err
	}
	return &drif.ActionResult{Data: comparison, HasData: true}, nil
}

// readStringList reads a list of ids from the request body. A malformed list reads as empty rather
// than partially, so the merge refuses instead of silently merging a subset.
func readStringList(params dmodel.DynamicFields, field string) []string {
	raw, ok := params[field]
	if !ok || raw == nil {
		return nil
	}
	switch typed := raw.(type) {
	case []string:
		return typed
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if text, ok := item.(string); ok {
				out = append(out, text)
			}
		}
		return out
	default:
		return nil
	}
}

// orderServiceOf reaches the derived service installed during Init. The engine hands the action its
// service as the base interface, and only the derived type carries Confirm, Approve and Cancel, so
// a failed assertion is a wiring bug rather than a request problem.
func orderServiceOf(input drif.ProcessInput) (*services.PurchaseOrderDomainServiceImpl, error) {
	service, ok := input.ResourceService.(*services.PurchaseOrderDomainServiceImpl)
	if !ok {
		return nil, errors.New(
			"the purchase order engine is not running the derived order service; " +
				"PurchaseModule.Init must install it with SetResourceService")
	}
	return service, nil
}

func readOrderId(input drif.ProcessInput) string {
	return readStringParam(input.Params, paramOrderId)
}

func readStringParam(params dmodel.DynamicFields, field string) string {
	value, ok := params[field]
	if !ok || value == nil {
		return ""
	}
	if typed, ok := value.(string); ok {
		return typed
	}
	return ""
}

// toMutateActionResult widens a mutation result into the engine's generic action result. It
// duplicates the engine's package-private toActionResult; ClientErrors must survive, because a
// refused operation reports its reason through them and not through err.
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
