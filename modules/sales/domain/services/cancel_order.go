package services

import (
	"time"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	lock "github.com/sky-as-code/nikki-erp/modules/core/infra/distributedlock"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// Cancelling a sales order (BR 43, BR 82, SALES-014).
//
// What a cancel MEANS depends entirely on how far the sale got, and BR 43 gives a different answer
// for each state:
//
//	draft                       -> cancel directly, nothing to undo
//	confirmed, unpaid, unfilled -> cancel, release the stock reservation, cancel pending payments
//	paid                        -> REFUSED. A paid order needs a refund, not a cancellation.
//	fulfilled                   -> REFUSED. Goods are with the customer; that is a return.
//
// The two refusals are the point of this operation. Silently cancelling a paid order would leave the
// business holding money against a sale that no longer exists, and cancelling a fulfilled one would
// leave goods with a customer against no record. Both must be told to the caller as a redirection to
// the right workflow rather than reported as a generic failure.
//
// Nothing is ever deleted (BR 68). A cancelled order keeps its lines, adjustments, events and - once
// they exist - its bills and payments. The evidence of what was attempted is the record.

// CancelOrderResult is what a cancel concluded.
type CancelOrderResult struct {
	SalesOrderId string
	Status       string
	CancelledAt  string

	// ReleasedVoucherIds are the reservations handed back by this cancel (BR 82).
	ReleasedVoucherIds []string

	// Pending names the steps this build cannot perform. A confirmed order's stock reservation and
	// pending payments cannot be released until SALES-029 and SALES-027 exist, and a caller that
	// believed otherwise would leave stock reserved against a cancelled sale.
	Pending []string
}

// The refusal reasons cancel can produce.
const (
	ReasonAlreadyCancelled = "sales_order.already_cancelled"
	ReasonRequiresRefund   = "sales_order.requires_refund"
	ReasonRequiresReturn   = "sales_order.requires_return"
	ReasonNotCancellable   = "sales_order.not_cancellable"
)

// CancelOrder cancels an order if its state allows it.
//
// Under the same distributed lock as confirm (D-30) and for the same reason: cancel is not a
// single-row update - it releases voucher redemptions on another table - and a cancel interleaved
// with a confirm could release a reservation the confirm had just redeemed.
func CancelOrder(
	ctx corectx.Context, orderId, reason string, dLock lock.DistributedLock,
) (*CancelOrderResult, *ft.ClientErrors, error) {
	if dLock == nil {
		return nil, nil, errors.New(
			"the distributed lock is not available; a sales order cannot be cancelled without it")
	}

	key := confirmLockKeyOf(orderId)
	acquired, err := dLock.AcquireWithRetry(
		ctx, key, confirmLockTtl, confirmLockRetryCount, confirmLockRetryDelay)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "acquiring the lock of order '%s'", orderId)
	}
	if !acquired {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation("id", ReasonLockUnavailable,
			"this order is being changed by another request; try again"))
		return nil, vErrs, nil
	}
	defer func() { _ = dLock.Release(ctx, key) }()

	return cancelUnderLock(ctx, orderId, reason)
}

func cancelUnderLock(
	ctx corectx.Context, orderId, reason string,
) (*CancelOrderResult, *ft.ClientErrors, error) {
	record, err := loadRecord(ctx, models.SalesOrderSchemaName, models.SalesOrderFieldId, orderId)
	if err != nil {
		return nil, nil, err
	}
	if record == nil {
		return nil, OrderNotFoundErrors(orderId), nil
	}

	if vErrs := assertCancellable(record); vErrs != nil {
		return nil, vErrs, nil
	}

	fromStatus := stringOf(record, models.SalesOrderFieldStatus)

	// Vouchers go back BEFORE the status moves, so that a failure to release leaves the order
	// cancellable again rather than cancelled with a use still held against it.
	released, err := releaseOrderVouchers(ctx, orderId)
	if err != nil {
		return nil, nil, err
	}

	cancelledAt := time.Now().UTC()
	if err := stampCancelled(ctx, orderId, record, fromStatus, reason, cancelledAt); err != nil {
		return nil, nil, err
	}

	return &CancelOrderResult{
		SalesOrderId:       orderId,
		Status:             string(models.SalesOrderStatusCancelled),
		CancelledAt:        cancelledAt.Format(time.RFC3339),
		ReleasedVoucherIds: released,
		Pending:            pendingCancelSteps(fromStatus),
	}, nil, nil
}

// pendingCancelSteps names what a cancel of this order cannot yet undo.
//
// A draft has nothing to undo, so its list is empty and the cancel really is complete. A confirmed
// order does, and saying so is what stops a caller assuming the stock came back.
func pendingCancelSteps(fromStatus string) []string {
	if fromStatus == string(models.SalesOrderStatusDraft) {
		return nil
	}
	return []string{
		"release_stock_reservation (SALES-029: no inventory port bound)",
		"cancel_pending_payments (SALES-027: sales_payments does not exist)",
	}
}

// assertCancellable decides whether this order may be cancelled at all (BR 43).
//
// The two refusals name the workflow the caller should use instead. A message saying only "cannot
// cancel" would leave an operator with a paid order and no idea what to do with it.
func assertCancellable(record dmodel.DynamicFields) *ft.ClientErrors {
	refuse := func(field, reason, message string) *ft.ClientErrors {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation(field, reason, message))
		return vErrs
	}

	status := stringOf(record, models.SalesOrderFieldStatus)

	if status == string(models.SalesOrderStatusCancelled) {
		// Already cancelled. Refused rather than treated as a no-op, because a cancel releases
		// vouchers: a silent second success would release a reservation twice, and the second
		// release would give back a use the order never held.
		return refuse("status", ReasonAlreadyCancelled, "this order is already cancelled")
	}
	if status == string(models.SalesOrderStatusCompleted) {
		return refuse("status", ReasonNotCancellable,
			"a completed order cannot be cancelled; raise a return instead")
	}

	// Money first: a paid order is the case that costs real money to get wrong.
	paymentStatus := stringOf(record, models.SalesOrderFieldPaymentStatus)
	switch paymentStatus {
	case string(models.SalesOrderPaymentStatusPaid),
		string(models.SalesOrderPaymentStatusPartiallyPaid),
		string(models.SalesOrderPaymentStatusOverpaid):
		return refuse("payment_status", ReasonRequiresRefund,
			"this order has been paid and cannot be cancelled on its own; it needs a refund")
	}

	fulfillmentStatus := stringOf(record, models.SalesOrderFieldFulfillmentStatus)
	switch fulfillmentStatus {
	case string(models.SalesOrderFulfillmentStatusFulfilled),
		string(models.SalesOrderFulfillmentStatusPartiallyFulfilled):
		return refuse("fulfillment_status", ReasonRequiresReturn,
			"goods have been delivered against this order; it needs a return, not a cancellation")
	}

	return nil
}

// releaseOrderVouchers hands back every use this order was holding (BR 82).
//
// Only RESERVED redemptions are released. A redeemed one belongs to a confirmed sale, and this
// operation cannot reach a confirmed-and-paid order anyway; a released or reversed one is already
// given back, and releasing it twice would credit a use the code never lost.
func releaseOrderVouchers(ctx corectx.Context, orderId string) ([]string, error) {
	redemptions, err := searchBy(ctx,
		models.SalesVoucherRedemptionSchemaName,
		models.SalesVoucherRedemptionFieldSalesOrderId, orderId)
	if err != nil {
		return nil, err
	}

	released := make([]string, 0, len(redemptions))
	for _, record := range redemptions {
		if stringOf(record, models.SalesVoucherRedemptionFieldStatus) !=
			string(models.VoucherRedemptionStatusReserved) {
			continue
		}

		redemptionId := stringOf(record, models.SalesVoucherRedemptionFieldId)
		vErrs, err := SettleRedemption(ctx, redemptionId,
			string(models.VoucherRedemptionStatusReleased))
		if err != nil {
			return nil, err
		}
		if vErrs != nil {
			// A release that the transition table refuses means the row moved under us despite the
			// lock. Not fatal to the cancel: the use is not held by a reservation that is no longer
			// reserved, so carry on rather than blocking a cancellation on it.
			continue
		}
		released = append(released,
			stringOf(record, models.SalesVoucherRedemptionFieldVoucherCodeId))
	}
	return released, nil
}

// stampCancelled moves the status and records why.
//
// The reason travels into the audit event rather than onto the order. BR 86 wants the trail to say
// what happened and why; a column on the order would hold only the most recent answer, and a
// cancelled order is cancelled exactly once.
// All three writes are ONE transaction, for the reason stampConfirmed is: an integration event
// announcing a cancellation that then rolled back would leave consumers releasing stock and
// reversing money for a sale that is still live.
func stampCancelled(
	ctx corectx.Context, orderId string, record dmodel.DynamicFields,
	fromStatus, reason string, at time.Time,
) error {
	return withTransaction(ctx, models.SalesOrderSchemaName, func(tranxCtx corectx.Context) error {
		return stampCancelledInTranx(tranxCtx, orderId, record, fromStatus, reason, at)
	})
}

func stampCancelledInTranx(
	ctx corectx.Context, orderId string, record dmodel.DynamicFields,
	fromStatus, reason string, at time.Time,
) error {
	engine, err := engineFor(models.SalesOrderSchemaName)
	if err != nil {
		return err
	}

	update := dmodel.DynamicFields{
		models.SalesOrderFieldId:          orderId,
		models.SalesOrderFieldStatus:      string(models.SalesOrderStatusCancelled),
		models.SalesOrderFieldCancelledAt: model.ModelDateTime(at),
	}
	if _, err := engine.ResourceRepository().Update(ctx, update); err != nil {
		return err
	}

	orgId := stringOf(record, basemodel.FieldOrgId)
	if err := WriteSalesAuditEvent(ctx, SalesAuditEntry{
		SalesOrderId: orderId,
		EntityType:   models.SalesOrderSchemaName,
		EntityId:     orderId,
		Action:       models.SalesOrderActionCancel,
		FromStatus:   fromStatus,
		ToStatus:     string(models.SalesOrderStatusCancelled),
		Reason:       reason,
		OrgId:        orgId,
	}); err != nil {
		return err
	}

	_, err = RecordEvent(ctx, RecordEventParams{
		EventType:   models.EventSalesOrderCancelled,
		AggregateId: orderId,
		OrgId:       orgId,
		OccurredAt:  at.Unix(),
		Payload: map[string]any{
			"sales_order_id": orderId,
			"order_number":   stringOf(record, models.SalesOrderFieldOrderNumber),

			// The status it came FROM, because what a consumer must undo depends on it: cancelling a
			// draft releases nothing, while cancelling a confirmed sale releases a reservation.
			"from_status": fromStatus,
			"reason":      reason,

			"cancelled_at": at.Unix(),
		},
	})
	return err
}
