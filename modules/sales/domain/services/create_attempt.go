package services

import (
	"time"

	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	lock "github.com/sky-as-code/nikki-erp/modules/core/infra/distributedlock"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// Asking an executor to hand goods over, and recording that we asked.
//
// The attempt row is written BEFORE the executor is told anything. That ordering is what makes a
// dispense accountable: a machine that takes a command and then goes silent leaves a `pending`
// attempt naming what it was asked for and when, which an operator can find. Telling it first and
// recording afterwards would lose exactly the cases that matter.
//
// Everything here runs under the per-order distributed lock, because `attempt_no` counts tries at one
// delivery and two concurrent requests would otherwise both read the same count and both claim to be
// attempt 2 — making max_attempts uncountable and letting two machines act on one reservation.

const (
	ReasonAttemptNotAttemptable     = "sales.fulfillment.not_attemptable"
	ReasonAttemptWrongExecutor      = "sales.fulfillment.executor_not_target"
	ReasonAttemptAlreadyOutstanding = "sales.fulfillment.attempt_already_outstanding"
	ReasonAttemptExhausted          = "sales.fulfillment.max_attempts_reached"
	ReasonAttemptReservationLapsed  = "sales.fulfillment.reservation_expired"
	ReasonAttemptQuantityExceeded   = "sales.fulfillment.quantity_exceeds_fulfillable"
	ReasonAttemptNoItems            = "sales.fulfillment.no_items_requested"
	ReasonAttemptItemNotFound       = "sales.fulfillment.item_not_in_fulfillment"
	ReasonAttemptNotFound           = "sales.fulfillment.attempt_not_found"
	ReasonFulfillmentNotFound       = "sales.fulfillment.not_found"
)

// CreateAttemptParams is what a caller asks for. The executor is named explicitly rather than read
// from the fulfillment so that a request aimed at the wrong machine is REFUSED rather than silently
// redirected to the right one — a kiosk asking to dispense somebody else's order is a bug worth
// hearing about, not something to quietly correct.
type CreateAttemptParams struct {
	FulfillmentId    string
	ExecutorOutletId string
	Items            []CreateAttemptItem
}

// CreateAttemptItem is one product and how much of it to try.
type CreateAttemptItem struct {
	FulfillmentItemId string
	Quantity          decimal.Decimal
}

// CreateAttemptResult is what the caller needs to drive the executor and match its reply.
type CreateAttemptResult struct {
	AttemptId string
	AttemptNo int32

	// ExternalCorrelationId is what the executor must echo back. Devices that only know an order
	// code carry it in the command payload rather than as their own key.
	ExternalCorrelationId string
}

// CreateAttempt records a try and returns what the caller needs to command the executor.
func CreateAttempt(
	ctx corectx.Context,
	params CreateAttemptParams,
	dLock lock.DistributedLock,
) (*CreateAttemptResult, *ft.ClientErrors, error) {
	if dLock == nil {
		// Without the lock two requests can both become attempt 2, which makes the max_attempts
		// policy uncountable and lets two machines act on one reservation.
		return nil, nil, errors.New(
			"the distributed lock is not available; a fulfillment attempt cannot be created without it")
	}

	fulfillment, err := loadRecord(ctx, models.SalesOrderFulfillmentSchemaName,
		models.SalesOrderFulfillmentFieldId, params.FulfillmentId)
	if err != nil {
		return nil, nil, err
	}
	if fulfillment == nil {
		return nil, attemptRefusal(ReasonFulfillmentNotFound,
			"no fulfillment with id "+params.FulfillmentId), nil
	}

	orderId := stringOf(fulfillment, models.SalesOrderFulfillmentFieldSalesOrderId)
	key := confirmLockKeyOf(orderId)
	acquired, err := dLock.AcquireWithRetry(
		ctx, key, confirmLockTtl, confirmLockRetryCount, confirmLockRetryDelay)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "acquiring the lock of order '%s'", orderId)
	}
	if !acquired {
		return nil, attemptRefusal(ReasonLockUnavailable,
			"this order is being changed by another request; try again"), nil
	}
	defer func() { _ = dLock.Release(ctx, key) }()

	// Re-read under the lock: the record fetched while queuing describes the world before the other
	// holder finished, and its status is exactly what the guards below turn on.
	fulfillment, err = loadRecord(ctx, models.SalesOrderFulfillmentSchemaName,
		models.SalesOrderFulfillmentFieldId, params.FulfillmentId)
	if err != nil {
		return nil, nil, err
	}
	if fulfillment == nil {
		return nil, attemptRefusal(ReasonFulfillmentNotFound,
			"no fulfillment with id "+params.FulfillmentId), nil
	}
	return createAttemptUnderLock(ctx, params, fulfillment)
}

func createAttemptUnderLock(
	ctx corectx.Context, params CreateAttemptParams, fulfillment dmodel.DynamicFields,
) (*CreateAttemptResult, *ft.ClientErrors, error) {
	// The order behind the fulfillment must be paid, alive and inside its deadline before a
	// machine is told to hand anything over (CR-INV-SALES-WH-RESERVATION §9.2).
	if vErrs, err := assertOrderDispensable(ctx, stringOf(fulfillment, models.SalesOrderFulfillmentFieldSalesOrderId)); err != nil || vErrs != nil {
		return nil, vErrs, err
	}
	if vErrs := assertAttemptable(fulfillment, params, time.Now().UTC()); vErrs != nil {
		return nil, vErrs, nil
	}
	if len(params.Items) == 0 {
		return nil, attemptRefusal(ReasonAttemptNoItems,
			"name at least one item to attempt"), nil
	}

	items, err := ItemsOfFulfillment(ctx, params.FulfillmentId)
	if err != nil {
		return nil, nil, err
	}
	itemsById := make(map[string]dmodel.DynamicFields, len(items))
	for _, item := range items {
		itemsById[stringOf(item, models.SalesOrderFulfillmentItemFieldId)] = item
	}

	attempts, err := attemptsOfFulfillment(ctx, params.FulfillmentId)
	if err != nil {
		return nil, nil, err
	}
	if vErrs := assertNoOutstandingAttempt(attempts); vErrs != nil {
		return nil, vErrs, nil
	}

	attemptNo := int32(len(attempts) + 1)
	if vErrs := assertAttemptsRemain(fulfillment, attemptNo); vErrs != nil {
		return nil, vErrs, nil
	}
	fulfillable, err := FulfillableQuantities(ctx, params.FulfillmentId)
	if err != nil {
		return nil, nil, err
	}
	if vErrs := assertRequestedWithinFulfillable(params.Items, itemsById, fulfillable); vErrs != nil {
		return nil, vErrs, nil
	}

	attemptId, correlationId, err := writeAttempt(ctx, params, fulfillment, attemptNo, itemsById)
	if err != nil {
		return nil, nil, err
	}

	// Only now does the fulfillment move: a machine is about to be told to dispense, and the status
	// is what stops anything else acting on the same reservation while it does.
	if err := setFulfillmentStatus(ctx, params.FulfillmentId, models.FulfillmentStatusInProgress); err != nil {
		return nil, nil, err
	}

	return &CreateAttemptResult{
		AttemptId:             attemptId,
		AttemptNo:             attemptNo,
		ExternalCorrelationId: correlationId,
	}, nil, nil
}

// assertAttemptable covers the guards that read only the fulfillment itself.
func assertAttemptable(
	fulfillment dmodel.DynamicFields, params CreateAttemptParams, now time.Time,
) *ft.ClientErrors {
	record := models.NewSalesOrderFulfillmentFrom(fulfillment)

	if !record.ExecutesKioskDispense() {
		return attemptRefusal(ReasonMethodTypeUnsupported,
			"only a kiosk-dispense fulfillment can be attempted")
	}
	if !record.IsAttemptable() {
		return attemptRefusal(ReasonAttemptNotAttemptable,
			"a fulfillment in status '"+
				stringOf(fulfillment, models.SalesOrderFulfillmentFieldFulfillmentStatus)+
				"' cannot be attempted")
	}

	// The executor must be the current target. After a reassignment the old machine still holds a
	// stale command; letting it dispense would hand over goods reserved somewhere else entirely.
	target := stringOf(fulfillment, models.SalesOrderFulfillmentFieldTargetOutletId)
	if params.ExecutorOutletId == "" || params.ExecutorOutletId != target {
		return attemptRefusal(ReasonAttemptWrongExecutor,
			"this fulfillment is targeted at another sales point")
	}

	// An expired reservation holds no stock. Dispensing against it would hand over goods that were
	// returned to the sellable pool and may already have been sold to somebody else.
	if record.HasLapsed(now) {
		return attemptRefusal(ReasonAttemptReservationLapsed,
			"this fulfillment's stock reservation has expired; reserve again before attempting")
	}
	return nil
}

// assertNoOutstandingAttempt is the single-flight guard. Two executors acting on one reservation
// hand over goods that were paid for once.
func assertNoOutstandingAttempt(attempts []dmodel.DynamicFields) *ft.ClientErrors {
	for _, record := range attempts {
		if models.NewSalesFulfillmentAttemptFrom(record).IsOutstanding() {
			return attemptRefusal(ReasonAttemptAlreadyOutstanding,
				"an attempt on this fulfillment is still awaiting its result")
		}
	}
	return nil
}

// assertAttemptsRemain enforces the snapshotted counter. A NULL max_attempts means the customer
// decides when to stop rather than a counter, so it never exhausts.
func assertAttemptsRemain(fulfillment dmodel.DynamicFields, attemptNo int32) *ft.ClientErrors {
	maxAttempts := optionalInt32Of(fulfillment, models.SalesOrderFulfillmentFieldMaxAttempts)
	if maxAttempts == nil {
		return nil
	}
	if attemptNo > *maxAttempts {
		return attemptRefusal(ReasonAttemptExhausted,
			"this fulfillment has used all the attempts its method allows")
	}
	return nil
}

// assertRequestedWithinFulfillable checks each requested quantity against what may be attempted NOW,
// which is remaining minus anything with a refund in flight — not remaining alone. Dispensing
// against a quantity that is being refunded would hand over goods while paying for them in reverse.
func assertRequestedWithinFulfillable(
	requested []CreateAttemptItem,
	itemsById map[string]dmodel.DynamicFields,
	fulfillable map[string]decimal.Decimal,
) *ft.ClientErrors {
	seen := make(map[string]struct{}, len(requested))
	for _, item := range requested {
		if _, known := itemsById[item.FulfillmentItemId]; !known {
			return attemptRefusal(ReasonAttemptItemNotFound,
				"item "+item.FulfillmentItemId+" does not belong to this fulfillment")
		}
		if _, repeated := seen[item.FulfillmentItemId]; repeated {
			return attemptRefusal(ReasonAttemptItemNotFound,
				"item "+item.FulfillmentItemId+" is named twice in one attempt")
		}
		seen[item.FulfillmentItemId] = struct{}{}

		if !item.Quantity.IsPositive() {
			return attemptRefusal(ReasonQuantityNotPositive,
				"item "+item.FulfillmentItemId+" must be attempted with a positive quantity")
		}
		if item.Quantity.GreaterThan(fulfillable[item.FulfillmentItemId]) {
			// Deliberately not "more than it still owes": a quantity with a refund in flight is owed
			// and NOT attemptable, and a message about what is owed would send a caller to look at a
			// number that says the attempt should have worked.
			return attemptRefusal(ReasonAttemptQuantityExceeded,
				"item "+item.FulfillmentItemId+" cannot be attempted for more than may be "+
					"fulfilled now; a refund in flight blocks the quantity it covers")
		}
	}
	return nil
}

// writeAttempt stores the attempt and its items in one transaction: an attempt row with no items
// would command a machine to dispense nothing while occupying the single-flight slot.
func writeAttempt(
	ctx corectx.Context,
	params CreateAttemptParams,
	fulfillment dmodel.DynamicFields,
	attemptNo int32,
	itemsById map[string]dmodel.DynamicFields,
) (attemptId string, correlationId string, err error) {
	id, err := model.NewId()
	if err != nil {
		return "", "", err
	}
	attemptId = string(*id)

	// The correlation id is the attempt's own id. A second identifier would have to be kept in step
	// with it for no gain, and this way a reply naming a correlation is already naming the record.
	correlationId = attemptId
	orgId := stringOf(fulfillment, basemodel.FieldOrgId)
	startedAt := model.ModelDateTime(time.Now().UTC())

	err = withTransaction(ctx, models.SalesFulfillmentAttemptSchemaName, func(tranxCtx corectx.Context) error {
		engineRepo, err := repoFor(models.SalesFulfillmentAttemptSchemaName)
		if err != nil {
			return err
		}
		fields := dmodel.DynamicFields{
			models.SalesFulfillmentAttemptFieldId:                    attemptId,
			models.SalesFulfillmentAttemptFieldFulfillmentId:         params.FulfillmentId,
			models.SalesFulfillmentAttemptFieldAttemptNo:             attemptNo,
			models.SalesFulfillmentAttemptFieldExecutorOutletId:      params.ExecutorOutletId,
			models.SalesFulfillmentAttemptFieldAttemptStatus:         string(models.FulfillmentAttemptStatusPending),
			models.SalesFulfillmentAttemptFieldExternalCorrelationId: correlationId,
			models.SalesFulfillmentAttemptFieldStartedAt:             startedAt,
			basemodel.FieldOrgId:                                     orgId,
		}
		if _, err := engineRepo.Insert(tranxCtx, fields); err != nil {
			return errors.Wrap(err, "writing the fulfillment attempt")
		}
		return writeAttemptItems(tranxCtx, attemptId, orgId, params.Items)
	})
	if err != nil {
		return "", "", err
	}
	return attemptId, correlationId, nil
}

func writeAttemptItems(
	ctx corectx.Context, attemptId, orgId string, items []CreateAttemptItem,
) error {
	engineRepo, err := repoFor(models.SalesFulfillmentAttemptItemSchemaName)
	if err != nil {
		return err
	}
	for _, item := range items {
		itemId, err := model.NewId()
		if err != nil {
			return err
		}
		fields := dmodel.DynamicFields{
			models.SalesFulfillmentAttemptItemFieldId:                string(*itemId),
			models.SalesFulfillmentAttemptItemFieldAttemptId:         attemptId,
			models.SalesFulfillmentAttemptItemFieldFulfillmentItemId: item.FulfillmentItemId,
			models.SalesFulfillmentAttemptItemFieldAttemptedQty:      item.Quantity,

			// Zeroed until a result arrives. Nothing has been dispensed at the moment a machine is
			// merely asked, and the result is `pending` for the same reason.
			models.SalesFulfillmentAttemptItemFieldDispensedQty: decimal.Zero,
			models.SalesFulfillmentAttemptItemFieldFailedQty:    decimal.Zero,
			models.SalesFulfillmentAttemptItemFieldItemResult:   string(models.FulfillmentAttemptItemResultPending),

			basemodel.FieldOrgId: orgId,
		}
		if _, err := engineRepo.Insert(ctx, fields); err != nil {
			return errors.Wrap(err, "writing a fulfillment attempt item")
		}
	}
	return nil
}

// attemptsOfFulfillment lists every try at one delivery, newest included: the count is what the next
// attempt_no is derived from, so a filtered subset would reuse a number.
func attemptsOfFulfillment(
	ctx corectx.Context, fulfillmentId string,
) ([]dmodel.DynamicFields, error) {
	return searchBy(ctx, models.SalesFulfillmentAttemptSchemaName,
		models.SalesFulfillmentAttemptFieldFulfillmentId, fulfillmentId)
}

// itemsOfAttempt lists what one try reported on.
func itemsOfAttempt(ctx corectx.Context, attemptId string) ([]dmodel.DynamicFields, error) {
	return searchBy(ctx, models.SalesFulfillmentAttemptItemSchemaName,
		models.SalesFulfillmentAttemptItemFieldAttemptId, attemptId)
}

func attemptRefusal(reason, message string) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(
		models.SalesFulfillmentAttemptSchemaName, reason, message))
	return vErrs
}
