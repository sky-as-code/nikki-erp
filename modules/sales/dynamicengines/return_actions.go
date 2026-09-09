package dynamicengines

import (
	stdErr "errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
)

// Creating a return and processing it are separate actions so a front-desk role can raise a return
// that a supervisor releases: creating writes nothing irreversible, processing moves goods and money
// and cannot be undone.

const (
	// Records the request only; deliberately not the power to refund.
	PermissionCreateReturn = "create"

	// Its own permission rather than `update`, because it moves money out of the business.
	PermissionProcessReturn = "process_return"

	// `update`: cancelling before anything irreversible is an ordinary correction, and the state
	// machine already refuses it afterwards.
	PermissionCancelReturn = "update"

	ActionCreateReturn  = "create_return"
	ActionProcessReturn = "process_return"
	ActionCancelReturn  = "cancel_return"
)

func defineSalesReturnActions(engine drif.DynamicResourceEngine) error {
	return stdErr.Join(
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName: ActionCreateReturn,
			ActionType: drif.ActionTypeGeneric,

			// Underscores, never hyphens: the route must match the action code.
			RestPath:    "create_return",
			Permission:  PermissionCreateReturn,
			MainProcess: processCreateReturn,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionProcessReturn,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/process",
			Permission:  PermissionProcessReturn,
			MainProcess: processProcessReturn,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionCancelReturn,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/cancel",
			Permission:  PermissionCancelReturn,
			MainProcess: processCancelReturn,
		}),
	)
}

func processCreateReturn(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	policy := services.ResolveSalesPolicy(ctx, effectiveSettings)
	result, vErrs, err := services.CreateReturn(ctx, services.CreateReturnParams{
		SalesOrderId:         readStringParam(input.Params, "sales_order_id"),
		Reason:               readStringParam(input.Params, "reason"),
		InventoryDisposition: readStringParam(input.Params, "inventory_disposition"),
		Lines:                readReturnLines(input.Params),
	}, orderLock, policy)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
	}
	return &drif.ActionResult{Data: result}, nil
}

func processProcessReturn(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	result, vErrs, err := services.ProcessReturn(ctx, services.ProcessReturnParams{
		SalesReturnId: readStringParam(input.Params, paramId),
	}, orderLock, orderFulfillment, invoicingProvider, paymentOrders)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
	}
	return &drif.ActionResult{Data: result}, nil
}

func processCancelReturn(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	result, vErrs, err := services.CancelReturn(ctx, services.CancelReturnParams{
		SalesReturnId: readStringParam(input.Params, paramId),
		Reason:        readStringParam(input.Params, "reason"),
	}, orderLock)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
	}
	return &drif.ActionResult{Data: result}, nil
}

// readReturnLines reads only an order line and a quantity. The refund amount is computed from what
// the line historically carried, never accepted from the caller.
func readReturnLines(params map[string]any) []services.CreateReturnLine {
	raw, ok := params["lines"].([]any)
	if !ok {
		return nil
	}

	lines := make([]services.CreateReturnLine, 0, len(raw))
	for _, item := range raw {
		fields, ok := item.(map[string]any)
		if !ok {
			continue
		}
		lines = append(lines, services.CreateReturnLine{
			SalesOrderLineId: readStringParam(fields, "sales_order_line_id"),
			Quantity:         readDecimalParam(fields, "quantity"),
		})
	}
	return lines
}
