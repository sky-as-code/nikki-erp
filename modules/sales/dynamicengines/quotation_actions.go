package dynamicengines

import (
	stdErr "errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
)

// The quotation actions (BR 87.1, SALES-038).
//
// # Convert is its own permission, `convert`, and not `update`
//
// Accepting a quotation creates a sales order — it commits the business to a sale. A role that may
// draft and correct an offer should not thereby be able to turn one into a binding order, which is
// exactly what folding this into `update` would do.
//
// Sending and cancelling ride on `update`, because both are ordinary handling of a document by
// whoever owns it: sending is showing the customer what you wrote, cancelling is withdrawing it.
// Neither creates anything.

const (
	// PermissionConvertQuotation commits the business to a sale. See above.
	PermissionConvertQuotation = "convert"

	ActionConvertQuotation = "convert"

	// PermissionTransitionQuotation is `update`: sending and cancelling are ordinary handling of a
	// document, not powers of their own.
	PermissionTransitionQuotation = "update"

	ActionSendQuotation   = "send"
	ActionCancelQuotation = "cancel"
)

// defineSalesQuotationActions adds convert, send and cancel to the quotation engine.
func defineSalesQuotationActions(engine drif.DynamicResourceEngine) error {
	return stdErr.Join(
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionConvertQuotation,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/convert",
			Permission:  PermissionConvertQuotation,
			MainProcess: processConvertQuotation,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionSendQuotation,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/send",
			Permission:  PermissionTransitionQuotation,
			MainProcess: processSendQuotation,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionCancelQuotation,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/cancel",
			Permission:  PermissionTransitionQuotation,
			MainProcess: processCancelQuotation,
		}),
	)
}

// processConvertQuotation turns an accepted offer into a sales order.
func processConvertQuotation(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	policy := services.ResolveSalesPolicy(ctx, effectiveSettings)

	result, vErrs, err := services.ConvertQuotation(ctx, services.ConvertQuotationParams{
		SalesQuotationId: readStringParam(input.Params, paramId),
		SalesPointId:     readStringParam(input.Params, "sales_point_id"),
		IdempotencyKey:   readStringParam(input.Params, "idempotency_key"),
	}, taxCalculation, productVariants, policy)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
	}

	return &drif.ActionResult{
		HasData: true,
		Data: map[string]any{
			"sales_quotation_id": result.SalesQuotationId,
			"sales_order_id":     result.SalesOrderId,
			"order_number":       result.OrderNumber,
			"already_converted":  result.AlreadyConverted,

			// BOTH totals, so a caller can see whether repricing moved the number rather than
			// taking it on trust. An operator handing an order to a customer holding the quotation
			// needs to know before the customer does.
			"quoted_total": result.QuotedTotal,
			"order_total":  result.OrderTotal,
		},
	}, nil
}

func processSendQuotation(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	return transitionQuotationResult(ctx,
		readStringParam(input.Params, paramId), string(models.SalesQuotationStatusSent))
}

func processCancelQuotation(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	return transitionQuotationResult(ctx,
		readStringParam(input.Params, paramId), string(models.SalesQuotationStatusCancelled))
}

// transitionQuotationResult applies one status move and reports the outcome.
func transitionQuotationResult(
	ctx corectx.Context, quotationId, toStatus string,
) (*drif.ActionResult, error) {
	vErrs, err := services.TransitionQuotation(ctx, quotationId, toStatus)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
	}
	return &drif.ActionResult{
		HasData: true,
		Data: map[string]any{
			"sales_quotation_id": quotationId,
			"status":             toStatus,
		},
	}, nil
}
