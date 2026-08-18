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

// The purchase order's lifecycle operations, exposed as engine actions.
//
// They are engine actions rather than hand-written REST handlers, per docs/wiki/07 §6.7: the engine
// already does the permission check, the param binding and the response shaping, and a handler
// would have to restate all three. Each is a POST, because none of them is a CRUD verb — confirming
// is not an update to an order, it is an event that happens to one.
//
// Each carries its OWN permission rather than reusing `update`, because they are materially
// different powers. Confirming commits the business to a purchase and cannot be undone by an edit;
// approving is the control that spending policy rests on. A role that may correct a typo in a
// description should not thereby be able to do either. The IAM seed declares the same set.

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

	ActionMerge               = "merge"
	ActionCreateAlternative   = "create_alternative"
	ActionCompareAlternatives = "compare_alternatives"
)

// Param names the order actions read from the request.
const (
	paramOrderId = "id"
	paramReason  = "reason"
	// paramAlternativeChoice carries the answer to the open-alternatives warning of §31.
	paramAlternativeChoice = "alternative_choice"
	paramAlternativeVendor = "vendor_id"
	paramMergeOrderIds     = "order_ids"
)

// defineOrderActions adds the lifecycle operations alongside the delete guard.
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
		// Raising an alternative creates an order, so it carries create rather than a permission of
		// its own — the power it grants is the power to create an order.
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionCreateAlternative,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/create_alternative",
			Permission:  drif.PermissionCreate,
			MainProcess: processOrderCreateAlternative,
		}),
		// Comparing takes nothing and changes nothing: it answers a question about orders the
		// caller can already read.
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionCompareAlternatives,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/compare_alternatives",
			Permission:  drif.PermissionRead,
			MainProcess: processOrderCompareAlternatives,
		}),
		// Duplicating is a create, and carries the create permission rather than one of its own:
		// it produces a new draft order from data the caller can already read, which is exactly
		// what a role allowed to create orders may do by hand.
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

// The reason is validated in the service rather than through a ParamSchema, following the note
// PAY-010 left: an action whose params mix resource fields with action-specific input is not
// described by any single schema, and the check belongs where the rule is.
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

// Duplicate returns the new order rather than an affected count, because the caller's next move is
// to open it — and without the id in the response they would have to search for it.
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

// readStringList reads a list of ids from the request body.
//
// A malformed list is read as empty rather than guessed at, and the merge then refuses for needing
// two orders — which names the real problem better than a partial list silently merged.
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

// orderServiceOf reaches the derived service the module installed during Init.
//
// The type assertion is what makes the lifecycle operations reachable: the engine hands the action
// its service as the base interface, and only the derived type carries Confirm, Approve and Cancel.
// A failed assertion means Init did not install it, which is a wiring bug rather than a request
// problem.
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

// toMutateActionResult widens a mutation result into the engine's generic action result.
//
// The engine's own toActionResult is package-private to modules/dynamicresource/engine, so this is
// a local equivalent rather than an unnecessary duplicate: the ClientErrors must survive, because
// a refused operation reports its reason through them and not through err.
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
