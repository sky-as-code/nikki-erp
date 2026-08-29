package dynamicengines

import (
	stdErr "errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
)

const (
	// Convert has its own permission because it creates a sales order and commits the business to a
	// sale; a role that may draft and correct an offer must not thereby be able to bind one.
	PermissionConvertQuotation = "convert"

	ActionConvertQuotation = "convert"

	// Sending and cancelling are ordinary handling of a document, so they ride on `update`.
	PermissionTransitionQuotation = "update"

	ActionSendQuotation   = "send"
	ActionCancelQuotation = "cancel"
)

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

func processConvertQuotation(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	policy := services.ResolveSalesPolicy(ctx, effectiveSettings)

	result, vErrs, err := services.ConvertQuotation(ctx, services.ConvertQuotationParams{
		SalesQuotationId: readStringParam(input.Params, paramId),
		SalesPointId:     readStringParam(input.Params, "sales_point_id"),
		IdempotencyKey:   readStringParam(input.Params, "idempotency_key"),
	}, taxCalculation, productVariants, pricingBasis, policy)
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

			// Both totals, so the caller can see whether repricing moved the number.
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
