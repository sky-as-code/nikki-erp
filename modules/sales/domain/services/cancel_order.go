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
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// Cancelling a sales order. What a cancel means depends on how far the sale got:
//
// 	draft                       -> cancel directly, nothing to undo
// 	confirmed, unpaid, unfilled -> cancel, release the stock reservation and pending payments
// 	paid                        -> REFUSED. A paid order needs a refund, not a cancellation.
// 	fulfilled                   -> REFUSED. Goods are with the customer; that is a return.
//
// The two refusals are the point, and both are reported as a redirection to the right workflow.
// Nothing is ever deleted: a cancelled order keeps its lines, adjustments, events and payments.

type CancelOrderResult struct {
	SalesOrderId string
	Status       string
	CancelledAt  string

	ReleasedVoucherIds []string

	// ReleasedFulfillmentIds are the fulfillments whose held stock went back to the sellable pool.
	// Empty for an order that reserved nothing, which is every order that never named a kiosk.
	ReleasedFulfillmentIds []string

	// Pending names the steps this cancel did not perform. Payments are always among them: cancel
	// does not cancel them, and a caller believing otherwise would stop chasing money that is still
	// committed against a sale nobody will complete.
	Pending []string
}

const (
	ReasonAlreadyCancelled = "sales_order.already_cancelled"
	ReasonRequiresRefund   = "sales_order.requires_refund"
	ReasonRequiresReturn   = "sales_order.requires_return"
	ReasonNotCancellable   = "sales_order.not_cancellable"
)

// CancelOrder cancels an order if its state allows it, under the same distributed lock as confirm:
// cancel releases voucher redemptions on another table, and a cancel interleaved with a confirm
// could release a reservation the confirm had just redeemed.
func CancelOrder(
	ctx corectx.Context,
	orderId, reason string,
	dLock lock.DistributedLock,
	reservations itExt.FulfillmentReservationExtService,
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

	return cancelUnderLock(ctx, orderId, reason, reservations)
}

func cancelUnderLock(
	ctx corectx.Context,
	orderId, reason string,
	reservations itExt.FulfillmentReservationExtService,
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

	// Vouchers go back BEFORE the status moves, so a failure to release leaves the order cancellable
	// again rather than cancelled with a use still held against it.
	released, err := releaseOrderVouchers(ctx, orderId)
	if err != nil {
		return nil, nil, err
	}

	// The stock goes back BEFORE the status moves, for the same reason the vouchers do: a failure
	// here leaves the order cancellable again rather than cancelled with goods still held against a
	// sale nobody is going to complete.
	releasedFulfillments, err := ReleaseOrderFulfillments(ctx, orderId, reservations)
	if err != nil {
		return nil, nil, err
	}

	cancelledAt := time.Now().UTC()
	if err := stampCancelled(ctx, orderId, record, fromStatus, reason, cancelledAt); err != nil {
		return nil, nil, err
	}

	return &CancelOrderResult{
		SalesOrderId:           orderId,
		Status:                 string(models.SalesOrderStatusCancelled),
		CancelledAt:            cancelledAt.Format(time.RFC3339),
		ReleasedVoucherIds:     released,
		ReleasedFulfillmentIds: releasedFulfillments,
		Pending:                pendingCancelSteps(fromStatus, reservations),
	}, nil, nil
}

// pendingCancelSteps: a draft has nothing to undo; saying so for a confirmed order stops a caller
// assuming the money came back.
func pendingCancelSteps(
	fromStatus string, reservations itExt.FulfillmentReservationExtService,
) []string {
	if fromStatus == string(models.SalesOrderStatusDraft) {
		return nil
	}

	pending := make([]string, 0, 2)
	// The stock IS released now, through ReleaseOrderFulfillments — unless no port is bound, in
	// which case there was nothing to release it through and the caller must know the goods are
	// still held.
	if reservations == nil {
		pending = append(pending,
			"release_stock_reservation (no inventory port bound)")
	}
	// Payments remain genuinely undone: cancel does not cancel them, so the money is still
	// committed after a successful cancel and a caller told otherwise would stop chasing it.
	pending = append(pending, "cancel_pending_payments (cancel does not cancel payments)")
	return pending
}

// assertCancellable: the two refusals name the workflow to use instead, so an operator is not left
// with a paid order and no next step.
func assertCancellable(record dmodel.DynamicFields) *ft.ClientErrors {
	refuse := func(field, reason, message string) *ft.ClientErrors {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation(field, reason, message))
		return vErrs
	}

	status := stringOf(record, models.SalesOrderFieldStatus)

	if status == string(models.SalesOrderStatusCancelled) {
		// Already cancelled. Refused rather than a no-op, because a cancel releases vouchers: a
		// silent second success would give back a use the order never held.
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

// releaseOrderVouchers hands back every use this order was holding. Only RESERVED redemptions:
// a redeemed one belongs to a confirmed sale, and a released one is already given back, so
// releasing it twice would credit a use the code never lost.
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
			// A release the transition table refuses means the row moved under us despite the lock. Not
			// fatal: the use is not held by a reservation that is no longer reserved.
			continue
		}
		released = append(released,
			stringOf(record, models.SalesVoucherRedemptionFieldVoucherCodeId))
	}
	return released, nil
}

// stampCancelled moves the status and records why. The reason travels into the audit event rather
// than onto the order, which would hold only the most recent answer. All three writes are ONE
// transaction: an event announcing a cancellation that then rolled back would leave consumers
// releasing stock and reversing money for a live sale.
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
			// draft releases nothing, a confirmed sale releases a reservation.
			"from_status": fromStatus,
			"reason":      reason,

			"cancelled_at": at.Unix(),
		},
	})
	return err
}
