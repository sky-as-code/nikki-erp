package dynamicengines

import (
	stdErr "errors"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
)

// Commanding a dispense and hearing back about it.
//
// Both actions hang off the sales_order_fulfillment engine, because both are about ONE delivery, and
// both write: creating an attempt commits the single-flight slot and moves the fulfillment, while a
// result moves quantities that money depends on. They therefore answer `update` rather than `read`,
// even though neither edits the fulfillment record through ordinary CRUD.

const (
	ActionOrderRefunds   = "refunds"
	ActionCreateAttempt  = "attempts"
	ActionAttemptResult  = "attempt_result"
	ActionReassignTarget = "reassign_target"

	paramToOutletId = "to_outlet_id"
	paramReason     = "reason"

	paramExecutorOutletId       = "executor_outlet_id"
	paramFulfillmentItemId      = "fulfillment_item_id"
	paramAttemptId              = "attempt_id"
	paramResultEventId          = "result_event_id"
	paramCorrelationId          = "external_correlation_id"
	paramInventoryResultRef     = "inventory_result_ref"
	paramDispensedQty           = "dispensed_qty"
	paramFailedQty              = "failed_qty"
	paramFailureCode            = "failure_code"
	paramFailureMessage         = "failure_message"
	reasonAttemptItemsMalformed = "sales_fulfillment_attempt.items_malformed"
)

func defineSalesFulfillmentAttemptActions(engine drif.DynamicResourceEngine) error {
	return stdErr.Join(
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName: ActionReassignTarget,
			ActionType: drif.ActionTypeGeneric,

			// An action, never a PATCH of target_outlet_id. Writing that field directly would leave
			// the goods held at the old machine while the order promised them from the new one, and
			// no amount of validation on a field update can move stock.
			RestPath:    ":id/reassign_target",
			Permission:  drif.PermissionUpdate,
			MainProcess: processReassignTarget,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionCreateAttempt,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/attempts",
			Permission:  drif.PermissionUpdate,
			MainProcess: processCreateAttempt,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName: ActionAttemptResult,
			ActionType: drif.ActionTypeGeneric,

			// Addressed by the fulfillment with the attempt named in the body, rather than by the
			// attempt id in the path: the caller reporting a result is a device or a bridge that
			// knows the delivery it was told to make, and the org scoping the pipeline applies is
			// keyed on the fulfillment either way.
			RestPath:    ":id/attempt_result",
			Permission:  drif.PermissionUpdate,
			MainProcess: processAttemptResult,
		}),
	)
}

func processCreateAttempt(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	items, vErrs := readAttemptItems(input.Params)
	if vErrs != nil {
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
	}

	result, vErrs, err := services.CreateAttempt(ctx, services.CreateAttemptParams{
		FulfillmentId:    readStringParam(input.Params, paramId),
		ExecutorOutletId: readStringParam(input.Params, paramExecutorOutletId),
		Items:            items,
	}, orderLock)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
	}

	return &drif.ActionResult{
		HasData: true,
		Data: map[string]any{
			"attempt_id": result.AttemptId,
			"attempt_no": result.AttemptNo,

			// The executor must echo this back, which is how a reply is tied to the try that caused
			// it. A device that only knows an order code carries it in the command payload.
			"external_correlation_id": result.ExternalCorrelationId,
		},
	}, nil
}

func processAttemptResult(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	items, vErrs := readResultItems(input.Params)
	if vErrs != nil {
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
	}

	outcome, vErrs, err := services.ApplyAttemptResult(ctx, services.ApplyAttemptResultParams{
		AttemptId:             readStringParam(input.Params, paramAttemptId),
		ResultEventId:         readStringParam(input.Params, paramResultEventId),
		ExternalCorrelationId: readStringParam(input.Params, paramCorrelationId),
		InventoryResultRef:    readStringParam(input.Params, paramInventoryResultRef),
		ExecutorOutletId:      readStringParam(input.Params, paramExecutorOutletId),
		Items:                 items,
	}, orderLock, services.ResolveSalesPolicy(ctx, effectiveSettings))
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
	}

	data := map[string]any{
		"attempt_id":         outcome.AttemptId,
		"attempt_status":     outcome.AttemptStatus,
		"fulfillment_id":     outcome.FulfillmentId,
		"fulfillment_status": outcome.FulfillmentStatus,

		// True when this delivery was a replay of one already applied. Reported rather than hidden:
		// a bridge retrying after a lost reply should be able to tell that its retry changed nothing.
		"already_applied": outcome.AlreadyApplied,
	}
	if outcome.Refund != nil {
		// Creation is not success. The quantity is not refunded until settlement says so, which is
		// why a status travels with the id.
		data["refund"] = map[string]any{
			"refund_id":     outcome.Refund.RefundId,
			"refund_status": outcome.Refund.RefundStatus,
		}
	}
	return &drif.ActionResult{HasData: true, Data: data}, nil
}

// readAttemptItems parses what to try. A malformed quantity is refused rather than defaulted: zero
// would command a machine to dispense nothing and then report the try as complete.
func readAttemptItems(
	params dmodel.DynamicFields,
) ([]services.CreateAttemptItem, *ft.ClientErrors) {
	entries, vErrs := readItemObjects(params)
	if vErrs != nil {
		return nil, vErrs
	}

	items := make([]services.CreateAttemptItem, 0, len(entries))
	for _, fields := range entries {
		itemId, _ := fields[paramFulfillmentItemId].(string)
		if itemId == "" {
			return nil, attemptItemsViolation("each item must name a fulfillment_item_id")
		}
		quantity, ok := readDecimalValue(fields[paramQuantity])
		if !ok {
			return nil, attemptItemsViolation(
				"item " + itemId + " must carry a numeric quantity")
		}
		items = append(items, services.CreateAttemptItem{
			FulfillmentItemId: itemId,
			Quantity:          quantity,
		})
	}
	return items, nil
}

// readResultItems parses what happened. Both quantities default to zero when absent, which is
// meaningful here and not a guess: a report naming only what was dispensed is saying the rest failed,
// and the service refuses the pair if they do not add up to what was attempted.
func readResultItems(
	params dmodel.DynamicFields,
) ([]services.ApplyAttemptResultItem, *ft.ClientErrors) {
	entries, vErrs := readItemObjects(params)
	if vErrs != nil {
		return nil, vErrs
	}

	items := make([]services.ApplyAttemptResultItem, 0, len(entries))
	for _, fields := range entries {
		itemId, _ := fields[paramFulfillmentItemId].(string)
		if itemId == "" {
			return nil, attemptItemsViolation("each item must name a fulfillment_item_id")
		}

		dispensed, vErrs := readResultQuantity(fields, paramDispensedQty, itemId)
		if vErrs != nil {
			return nil, vErrs
		}
		failed, vErrs := readResultQuantity(fields, paramFailedQty, itemId)
		if vErrs != nil {
			return nil, vErrs
		}

		code, _ := fields[paramFailureCode].(string)
		message, _ := fields[paramFailureMessage].(string)
		items = append(items, services.ApplyAttemptResultItem{
			FulfillmentItemId: itemId,
			DispensedQty:      dispensed,
			FailedQty:         failed,
			FailureCode:       code,
			FailureMessage:    message,
		})
	}
	return items, nil
}

// readResultQuantity treats an absent quantity as zero but a malformed one as a violation. The
// difference matters: omitting `failed_qty` says nothing failed, while sending "abc" says the
// reporter is confused, and reading the second as zero would record a dispense that never happened.
func readResultQuantity(
	fields map[string]any, key, itemId string,
) (decimal.Decimal, *ft.ClientErrors) {
	value, present := fields[key]
	if !present || value == nil {
		return decimal.Zero, nil
	}
	quantity, ok := readDecimalValue(value)
	if !ok {
		return decimal.Zero, attemptItemsViolation(
			"item " + itemId + " carries a malformed " + key)
	}
	return quantity, nil
}

func readItemObjects(params dmodel.DynamicFields) ([]map[string]any, *ft.ClientErrors) {
	raw, present := params[paramItems]
	if !present || raw == nil {
		return nil, attemptItemsViolation("name the items this call is about")
	}
	list, ok := raw.([]any)
	if !ok {
		return nil, attemptItemsViolation("items must be a list of objects")
	}

	entries := make([]map[string]any, 0, len(list))
	for _, entry := range list {
		fields, ok := entry.(map[string]any)
		if !ok {
			return nil, attemptItemsViolation("each item must be an object")
		}
		entries = append(entries, fields)
	}
	return entries, nil
}

func attemptItemsViolation(message string) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(
		models.SalesFulfillmentAttemptSchemaName, reasonAttemptItemsMalformed, message))
	return vErrs
}

// processReassignTarget moves a delivery to another machine, stock and all.
func processReassignTarget(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	result, vErrs, err := services.ReassignTarget(ctx, services.ReassignTargetParams{
		FulfillmentId: readStringParam(input.Params, paramId),
		ToOutletId:    readStringParam(input.Params, paramToOutletId),
		Reason:        models.TargetChangeReason(readStringParam(input.Params, paramReason)),
	}, orderLock, fulfillmentReservations)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
	}

	return &drif.ActionResult{
		HasData: true,
		Data: map[string]any{
			"fulfillment_id":         result.FulfillmentId,
			"from_outlet_id":         result.FromOutletId,
			"to_outlet_id":           result.ToOutletId,
			"fulfillment_status":     result.Status,
			"inventory_reference":    result.InventoryReference,
			"reservation_expires_at": result.ReservationExpiresAt,

			// Whether the hold followed the customer or had to be taken again. Reported because a
			// lapsed reservation means the goods were genuinely back in the pool in between, which
			// is a different story to tell a customer than an uninterrupted move.
			"fresh_reservation": result.FreshReservation,
		},
	}, nil
}
