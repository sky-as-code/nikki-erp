package app

import (
	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/distributedlock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
	itFulfillment "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/fulfillment"
)

// The fulfillment's authorized surface: the dispense loop.
//
// All three actions reuse the update permission. Commanding a dispense, recording its result and
// moving a target each change what a delivery owes, which is the same power over the same sale --
// splitting them would let a role start a dispense it could not then report on.
//
// The distributed lock is handed to the operation services untouched; each acquires it before
// opening its transaction and releases it after the commit.

type SalesOrderFulfillmentApplicationServiceImpl struct {
	composable.CrudApplicationService

	orderLock               distributedlock.DistributedLock
	effectiveSettings       itExt.EffectiveSettingsExtService
	fulfillmentReservations itExt.FulfillmentReservationExtService
}

func NewSalesOrderFulfillmentApplicationService(
	base composable.CrudApplicationService,
	dLock distributedlock.DistributedLock,
	settings itExt.EffectiveSettingsExtService,
	reservations itExt.FulfillmentReservationExtService,
) itFulfillment.SalesOrderFulfillmentApplicationService {
	return &SalesOrderFulfillmentApplicationServiceImpl{
		CrudApplicationService:  base,
		orderLock:               dLock,
		effectiveSettings:       settings,
		fulfillmentReservations: reservations,
	}
}

func (this *SalesOrderFulfillmentApplicationServiceImpl) CreateAttempt(
	ctx corectx.Context, cmd itFulfillment.FulfillmentActionCommand,
) (*dyn.OpResult[any], error) {
	if cErrs, err := assertRecordAction(this, ctx, composable.PermissionUpdate, cmd); cErrs != nil || err != nil {
		return anyFailure(cErrs, err)
	}
	return this.runCreateAttempt(ctx, cmd)
}

func (this *SalesOrderFulfillmentApplicationServiceImpl) ApplyAttemptResult(
	ctx corectx.Context, cmd itFulfillment.FulfillmentActionCommand,
) (*dyn.OpResult[any], error) {
	if cErrs, err := assertRecordAction(this, ctx, composable.PermissionUpdate, cmd); cErrs != nil || err != nil {
		return anyFailure(cErrs, err)
	}
	return this.runAttemptResult(ctx, cmd)
}

func (this *SalesOrderFulfillmentApplicationServiceImpl) ReassignTarget(
	ctx corectx.Context, cmd itFulfillment.FulfillmentActionCommand,
) (*dyn.OpResult[any], error) {
	if cErrs, err := assertRecordAction(this, ctx, composable.PermissionUpdate, cmd); cErrs != nil || err != nil {
		return anyFailure(cErrs, err)
	}
	return this.runReassignTarget(ctx, cmd)
}

// The read-only fulfillment resources.

func NewSalesOrderFulfillmentItemApplicationService(base composable.CrudApplicationService) itFulfillment.SalesOrderFulfillmentItemApplicationService {
	return &SalesOrderFulfillmentItemApplicationServiceImpl{CrudApplicationService: base}
}

type SalesOrderFulfillmentItemApplicationServiceImpl struct {
	composable.CrudApplicationService
}

func NewSalesFulfillmentAttemptApplicationService(base composable.CrudApplicationService) itFulfillment.SalesFulfillmentAttemptApplicationService {
	return &SalesFulfillmentAttemptApplicationServiceImpl{CrudApplicationService: base}
}

type SalesFulfillmentAttemptApplicationServiceImpl struct {
	composable.CrudApplicationService
}

func NewSalesFulfillmentAttemptItemApplicationService(base composable.CrudApplicationService) itFulfillment.SalesFulfillmentAttemptItemApplicationService {
	return &SalesFulfillmentAttemptItemApplicationServiceImpl{CrudApplicationService: base}
}

type SalesFulfillmentAttemptItemApplicationServiceImpl struct {
	composable.CrudApplicationService
}

func NewSalesFulfillmentTargetChangeApplicationService(base composable.CrudApplicationService) itFulfillment.SalesFulfillmentTargetChangeApplicationService {
	return &SalesFulfillmentTargetChangeApplicationServiceImpl{CrudApplicationService: base}
}

type SalesFulfillmentTargetChangeApplicationServiceImpl struct {
	composable.CrudApplicationService
}

// The parameter names the moved bodies read.
const (
	paramToOutletId             = "to_outlet_id"
	paramReason                 = "reason"
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

// The moved action bodies follow.

func (this *SalesOrderFulfillmentApplicationServiceImpl) runCreateAttempt(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[any], error) {
	items, vErrs := readAttemptItems(params)
	if vErrs != nil {
		return &dyn.OpResult[any]{ClientErrors: *vErrs}, nil
	}

	result, vErrs, err := services.CreateAttempt(ctx, services.CreateAttemptParams{
		FulfillmentId:    readStringParam(params, paramRecordId),
		ExecutorOutletId: readStringParam(params, paramExecutorOutletId),
		Items:            items,
	}, this.orderLock)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &dyn.OpResult[any]{ClientErrors: *vErrs}, nil
	}

	return &dyn.OpResult[any]{
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

func (this *SalesOrderFulfillmentApplicationServiceImpl) runAttemptResult(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[any], error) {
	items, vErrs := readResultItems(params)
	if vErrs != nil {
		return &dyn.OpResult[any]{ClientErrors: *vErrs}, nil
	}

	outcome, vErrs, err := services.ApplyAttemptResult(ctx, services.ApplyAttemptResultParams{
		AttemptId:             readStringParam(params, paramAttemptId),
		ResultEventId:         readStringParam(params, paramResultEventId),
		ExternalCorrelationId: readStringParam(params, paramCorrelationId),
		InventoryResultRef:    readStringParam(params, paramInventoryResultRef),
		ExecutorOutletId:      readStringParam(params, paramExecutorOutletId),
		Items:                 items,
	}, this.orderLock, services.ResolveSalesPolicy(ctx, this.effectiveSettings))
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &dyn.OpResult[any]{ClientErrors: *vErrs}, nil
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
	return &dyn.OpResult[any]{HasData: true, Data: data}, nil
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
func (this *SalesOrderFulfillmentApplicationServiceImpl) runReassignTarget(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[any], error) {
	result, vErrs, err := services.ReassignTarget(ctx, services.ReassignTargetParams{
		FulfillmentId: readStringParam(params, paramRecordId),
		ToOutletId:    readStringParam(params, paramToOutletId),
		Reason:        models.TargetChangeReason(readStringParam(params, paramReason)),
	}, this.orderLock, this.fulfillmentReservations)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &dyn.OpResult[any]{ClientErrors: *vErrs}, nil
	}

	return &dyn.OpResult[any]{
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
