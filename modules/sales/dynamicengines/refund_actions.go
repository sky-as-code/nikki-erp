package dynamicengines

import (
	stdErr "errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
)

// Asking for money back against an order, and seeing what came of it.
//
// Client-agnostic on purpose: a customer app, a support console and a kiosk bridge all raise refunds
// the same way, and all of them go through the ONE return operation that a goods return uses. A
// second refund path would eventually refund differently from the first.

const (
	paramRefundLines      = "lines"
	paramRefundReason     = "reason"
	reasonRefundMalformed = "sales_return.lines_malformed"
)

// defineSalesOrderRefundActions hangs both routes off the ORDER, because that is what a customer
// asking for their money back names — they know what they bought, not which delivery it became.
func defineSalesOrderRefundActions(engine drif.DynamicResourceEngine) error {
	return stdErr.Join(
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName: ActionOrderRefunds,
			ActionType: drif.ActionTypeGeneric,
			RestPath:   ":id/refunds",

			// Raising a refund moves money, so it answers `update` on the order rather than `read`.
			Permission:  drif.PermissionUpdate,
			MainProcess: processCreateOrderRefund,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionOrderRefunds + "_view",
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/refunds_view",
			Permission:  drif.PermissionRead,
			MainProcess: processViewOrderRefunds,
		}),
	)
}

// processCreateOrderRefund raises a refund and reports it as PENDING.
//
// Creation is not success, and the response says so in as many words: the money has not moved yet,
// and a client that read a 200 as "refunded" would tell a customer they had been paid back when the
// legs may still be in flight.
func processCreateOrderRefund(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	lines, vErrs := readRefundLines(input.Params)
	if vErrs != nil {
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
	}

	result, vErrs, err := services.CreateReturn(ctx, services.CreateReturnParams{
		SalesOrderId: readStringParam(input.Params, paramId),
		Reason:       readStringParam(input.Params, paramRefundReason),

		// A customer asking is the only reason reachable from here. A fulfillment-failure refund is
		// raised by the failure policy itself and never by a request, which is what stops a caller
		// dressing an ordinary refund up as an automatic one to skip the goods-return step.
		RefundReason: models.SalesRefundReasonCustomerRequested,
		Lines:        lines,
	}, orderLock, services.ResolveSalesPolicy(ctx, effectiveSettings))
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
	}

	return &drif.ActionResult{
		HasData: true,
		Data: map[string]any{
			"refund_id":     result.SalesReturnId,
			"refund_status": result.RefundStatus,
			"refund_total":  result.RefundTotal,
		},
	}, nil
}

// processViewOrderRefunds lists what has been asked for and what has actually been paid.
func processViewOrderRefunds(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	views, err := services.ViewOrderRefunds(ctx, readStringParam(input.Params, paramId))
	if err != nil {
		return nil, err
	}

	refunds := make([]map[string]any, 0, len(views))
	for _, view := range views {
		items := make([]map[string]any, 0, len(view.Items))
		for _, item := range view.Items {
			items = append(items, map[string]any{
				"sales_order_line_id": item.SalesOrderLineId,
				"fulfillment_id":      item.FulfillmentId,
				"fulfillment_item_id": item.FulfillmentItemId,
				"requested_qty":       item.RequestedQty,

				// What actually went back, which is the number a customer service agent needs: the
				// requested figure says only what was asked for.
				"refunded_qty": item.RefundedQty,
			})
		}
		refunds = append(refunds, map[string]any{
			"refund_id":     view.RefundId,
			"refund_status": view.RefundStatus,
			"refund_reason": view.RefundReason,
			"return_type":   view.ReturnType,
			"refund_total":  view.RefundTotal,
			"items":         items,
		})
	}
	return &drif.ActionResult{
		HasData: true,
		Data:    map[string]any{"refunds": refunds},
	}, nil
}

// readRefundLines parses what to refund. A line names an order line and how much of it; the
// fulfillment linkage is resolved server-side, so a caller cannot point a refund at somebody else's
// delivery.
func readRefundLines(
	params dmodel.DynamicFields,
) ([]services.CreateReturnLine, *ft.ClientErrors) {
	raw, present := params[paramRefundLines]
	if !present || raw == nil {
		return nil, refundViolation("name the lines to refund")
	}
	list, ok := raw.([]any)
	if !ok {
		return nil, refundViolation("lines must be a list of objects")
	}

	lines := make([]services.CreateReturnLine, 0, len(list))
	for _, entry := range list {
		fields, ok := entry.(map[string]any)
		if !ok {
			return nil, refundViolation("each line must be an object")
		}

		lineId, _ := fields["sales_order_line_id"].(string)
		if lineId == "" {
			return nil, refundViolation("each line must name a sales_order_line_id")
		}
		quantity, ok := readDecimalValue(fields[paramQuantity])
		if !ok || !quantity.IsPositive() {
			return nil, refundViolation("line " + lineId + " must carry a positive quantity")
		}

		lines = append(lines, services.CreateReturnLine{
			SalesOrderLineId: lineId,
			Quantity:         quantity,
			RequestedQty:     quantity,
		})
	}
	return lines, nil
}

func refundViolation(message string) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(
		models.SalesReturnSchemaName, reasonRefundMalformed, message))
	return vErrs
}
