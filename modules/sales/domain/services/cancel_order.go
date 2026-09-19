package services

import (
	"github.com/shopspring/decimal"
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

	// RefundRequestId is the refund request raised for what a paid order had collected, and
	// RefundStatus where it stands: draft awaiting a confirmer, or already confirmed and
	// dispatched when the order's snapshot says refunds confirm themselves.
	RefundRequestId string
	RefundStatus    string
}

// CancelOrderOptions carries what a cancel records and may do beyond the order id.
type CancelOrderOptions struct {
	Reason string

	// CancellationNote is stored only when the caller gave one; the system never fills it in.
	CancellationNote string

	// SystemInitiated marks a cancel the system decided (stock could not be held after payment),
	// which may cancel a paid order whatever its snapshot says.
	SystemInitiated bool

	// Policy and RefundDeps are what raising and dispatching the refund of a paid order needs.
	Policy     SalesPolicy
	RefundDeps RefundProcessingDeps
}

const (
	ReasonAlreadyCancelled      = "sales_order.already_cancelled"
	ReasonRequiresRefund        = "sales_order.requires_refund"
	ReasonRequiresReturn        = "sales_order.requires_return"
	ReasonNotCancellable        = "sales_order.not_cancellable"
	ReasonFulfillmentInProgress = "sales_order.fulfillment_in_progress"
)

// CancelOrder cancels an order with a reason and nothing else. See CancelOrderWith.
func CancelOrder(
	ctx corectx.Context,
	orderId, reason string,
	dLock lock.DistributedLock,
	reservations itExt.FulfillmentReservationExtService,
) (*CancelOrderResult, *ft.ClientErrors, error) {
	return CancelOrderWith(ctx, orderId, CancelOrderOptions{Reason: reason}, dLock, reservations)
}

// CancelOrderWith cancels an order if its state allows it, under the same distributed lock as
// confirm: cancel releases voucher redemptions on another table, and a cancel interleaved with a
// confirm could release a reservation the confirm had just redeemed.
//
// A paid order is accepted when its snapshot says the channel confirms orders automatically, or
// when the system itself is cancelling (CR-INV-SALES-WH-RESERVATION §8.1): a refund request is
// raised for what was collected and, per the order's other snapshot, confirmed at once or left
// for a confirmer. An order whose kiosk may still be dispensing is refused until the attempt has
// reported: releasing that stock or refunding those goods would be guessing.
func CancelOrderWith(
	ctx corectx.Context,
	orderId string,
	opts CancelOrderOptions,
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

	return cancelUnderLock(ctx, orderId, opts, reservations)
}

func cancelUnderLock(
	ctx corectx.Context,
	orderId string,
	opts CancelOrderOptions,
	reservations itExt.FulfillmentReservationExtService,
) (*CancelOrderResult, *ft.ClientErrors, error) {
	record, err := loadRecord(ctx, models.SalesOrderSchemaName, models.SalesOrderFieldId, orderId)
	if err != nil {
		return nil, nil, err
	}
	if record == nil {
		return nil, OrderNotFoundErrors(orderId), nil
	}

	if vErrs := assertCancellableWith(record, opts); vErrs != nil {
		return nil, vErrs, nil
	}

	fromStatus := stringOf(record, models.SalesOrderFieldStatus)
	reason := opts.Reason

	if fromStatus != string(models.SalesOrderStatusDraft) {
		outstanding, err := hasOutstandingAttempt(ctx, orderId)
		if err != nil {
			return nil, nil, err
		}
		if outstanding {
			vErrs := ft.NewClientErrors()
			vErrs.Append(*ft.NewBusinessViolation("fulfillment_status", ReasonFulfillmentInProgress,
				"a kiosk may still be dispensing this order; wait for its result before cancelling"))
			return nil, vErrs, nil
		}
	}

	// The refund request for a paid order is raised while the order is still returnable, before
	// the status moves. A cancel that then fails leaves a draft request beside a live order, which
	// an operator can see and drop; the other order would leave a cancelled order with money and
	// no request at all.
	refundId, refundStatus := "", ""
	if isPaid(record) {
		refundId, refundStatus, err = raiseCancellationRefund(ctx, record, opts)
		if err != nil {
			return nil, nil, err
		}
	}

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
	if err := stampCancelled(ctx, orderId, record, fromStatus, reason, normaliseNote(opts.CancellationNote), cancelledAt); err != nil {
		return nil, nil, err
	}

	// Confirmed and dispatched at once when the order's snapshot says so; the confirm runs under
	// this same lock, and its own dispatch reports the gateway's answer rather than assuming one.
	if refundId != "" && boolOf(record, models.SalesOrderFieldAutoConfirmRefund) {
		confirmed, vErrs, err := confirmRefundRequestUnderLock(ctx, ConfirmRefundRequestParams{
			SalesReturnId: refundId, Automatic: true,
		}, opts.RefundDeps)
		if err != nil {
			return nil, nil, err
		}
		if vErrs == nil && confirmed != nil {
			refundStatus = confirmed.Status
		}
	}

	return &CancelOrderResult{
		SalesOrderId:           orderId,
		Status:                 string(models.SalesOrderStatusCancelled),
		CancelledAt:            cancelledAt.Format(time.RFC3339),
		ReleasedVoucherIds:     released,
		ReleasedFulfillmentIds: releasedFulfillments,
		Pending:                pendingCancelSteps(fromStatus, reservations),
		RefundRequestId:        refundId,
		RefundStatus:           refundStatus,
	}, nil, nil
}

// isPaid is the paid latch: paid or overpaid, and nothing a later refund does moves it back.
func isPaid(record dmodel.DynamicFields) bool {
	switch stringOf(record, models.SalesOrderFieldPaymentStatus) {
	case string(models.SalesOrderPaymentStatusPaid), string(models.SalesOrderPaymentStatusOverpaid),
		string(models.SalesOrderPaymentStatusRefunded), string(models.SalesOrderPaymentStatusPartiallyRefunded):
		return true
	}
	return false
}

// hasOutstandingAttempt reports whether any fulfillment of the order has an attempt awaiting its
// result: a machine that may still hand goods over.
func hasOutstandingAttempt(ctx corectx.Context, orderId string) (bool, error) {
	fulfillments, err := FulfillmentsOfOrder(ctx, orderId)
	if err != nil {
		return false, err
	}
	for _, fulfillment := range fulfillments {
		attempts, err := attemptsOfFulfillment(ctx, stringOf(fulfillment, models.SalesOrderFulfillmentFieldId))
		if err != nil {
			return false, err
		}
		for _, attempt := range attempts {
			if models.NewSalesFulfillmentAttemptFrom(attempt).IsOutstanding() {
				return true, nil
			}
		}
	}
	return false, nil
}

// raiseCancellationRefund raises a refund-only request for what a paid order still owes: every
// line's ordered quantity less what was handed over and what earlier requests already claimed.
// Delivered goods are not refunded here; they go through the ordinary return.
func raiseCancellationRefund(
	ctx corectx.Context, order dmodel.DynamicFields, opts CancelOrderOptions,
) (refundId, refundStatus string, err error) {
	orderId := stringOf(order, models.SalesOrderFieldId)
	lines, err := searchBy(ctx, models.SalesOrderLineSchemaName, models.SalesOrderLineFieldSalesOrderId, orderId)
	if err != nil {
		return "", "", err
	}
	claimed, err := refundOnlyClaimsOf(ctx, orderId)
	if err != nil {
		return "", "", err
	}

	requests := make([]CreateReturnLine, 0, len(lines))
	for _, line := range lines {
		lineId := stringOf(line, models.SalesOrderLineFieldId)
		owed := decimalOf(line, models.SalesOrderLineFieldOrderedQuantity).
			Sub(decimalOf(line, models.SalesOrderLineFieldFulfilledQuantity)).
			Sub(claimed[lineId])
		if !owed.IsPositive() {
			continue
		}
		requests = append(requests, CreateReturnLine{SalesOrderLineId: lineId, RequestedQty: owed})
	}
	if len(requests) == 0 {
		return "", "", nil
	}

	created, vErrs, err := createReturnUnderLock(ctx, CreateReturnParams{
		SalesOrderId: orderId,
		Reason:       "Cancelled before delivery: " + opts.Reason,
		RefundReason: models.SalesRefundReasonCustomerRequested,
		ReturnType:   models.SalesReturnTypeRefundOnly,
		Lines:        requests,
	}, order, opts.Policy)
	if err != nil {
		return "", "", err
	}
	if vErrs != nil && vErrs.Count() > 0 {
		// Nothing refundable was found (an unpaid balance, a line already claimed). The cancel
		// still stands; what was collected is reconciled through the ordinary refund path.
		return "", "", nil
	}
	return created.SalesReturnId, created.Status, nil
}

// refundOnlyClaimsOf sums, per order line, what refund requests not yet cancelled have already
// asked for, so a retried cancel never claims the same goods twice.
func refundOnlyClaimsOf(ctx corectx.Context, orderId string) (map[string]decimal.Decimal, error) {
	claims := map[string]decimal.Decimal{}
	returns, err := searchBy(ctx, models.SalesReturnSchemaName, models.SalesReturnFieldSalesOrderId, orderId)
	if err != nil {
		return nil, err
	}
	for _, salesReturn := range returns {
		if stringOf(salesReturn, models.SalesReturnFieldStatus) == string(models.SalesReturnStatusCancelled) {
			continue
		}
		lines, err := searchBy(ctx, models.SalesReturnLineSchemaName, models.SalesReturnLineFieldSalesReturnId,
			stringOf(salesReturn, models.SalesReturnFieldId))
		if err != nil {
			return nil, err
		}
		for _, line := range lines {
			lineId := stringOf(line, models.SalesReturnLineFieldSalesOrderLineId)
			claims[lineId] = claims[lineId].Add(decimalOf(line, models.SalesReturnLineFieldRequestedQty))
		}
	}
	return claims, nil
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

// assertCancellable: the refusals name the workflow to use instead, so an operator is not left
// with a paid order and no next step. A paid order passes when the order's own snapshot allows
// it or the system is cancelling; a partly delivered one passes on the same condition, with only
// the undelivered part refunded.
func assertCancellable(record dmodel.DynamicFields) *ft.ClientErrors {
	return assertCancellableWith(record, CancelOrderOptions{})
}

func assertCancellableWith(record dmodel.DynamicFields, opts CancelOrderOptions) *ft.ClientErrors {
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

	paidCancelAllowed := opts.SystemInitiated || boolOf(record, models.SalesOrderFieldAutoConfirmOrder)

	// Money first: a paid order is the case that costs real money to get wrong.
	paymentStatus := stringOf(record, models.SalesOrderFieldPaymentStatus)
	switch paymentStatus {
	case string(models.SalesOrderPaymentStatusPaid),
		string(models.SalesOrderPaymentStatusPartiallyPaid),
		string(models.SalesOrderPaymentStatusOverpaid):
		if !paidCancelAllowed {
			return refuse("payment_status", ReasonRequiresRefund,
				"this order has been paid and cannot be cancelled on its own; it needs a refund")
		}
	}

	fulfillmentStatus := stringOf(record, models.SalesOrderFieldFulfillmentStatus)
	switch fulfillmentStatus {
	case string(models.SalesOrderFulfillmentStatusFulfilled):
		return refuse("fulfillment_status", ReasonRequiresReturn,
			"goods have been delivered against this order; it needs a return, not a cancellation")
	case string(models.SalesOrderFulfillmentStatusPartiallyFulfilled):
		if !paidCancelAllowed {
			return refuse("fulfillment_status", ReasonRequiresReturn,
				"goods have been delivered against this order; it needs a return, not a cancellation")
		}
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
	fromStatus, reason, note string, at time.Time,
) error {
	return withTransaction(ctx, models.SalesOrderSchemaName, func(tranxCtx corectx.Context) error {
		return stampCancelledInTranx(tranxCtx, orderId, record, fromStatus, reason, note, at)
	})
}

func stampCancelledInTranx(
	ctx corectx.Context, orderId string, record dmodel.DynamicFields,
	fromStatus, reason, note string, at time.Time,
) error {
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

	// The status change and its announcement, as one step. previous_stage carries what it came FROM,
	// because what a consumer must undo depends on it: cancelling a draft releases nothing, a
	// confirmed sale releases a reservation.
	extra := dmodel.DynamicFields{
		models.SalesOrderFieldCancelledAt: model.ModelDateTime(at),
	}
	if note != "" {
		extra[models.SalesOrderFieldCancellationNote] = note
	}
	return TransitionOrderStage(ctx, record, string(models.SalesOrderStatusCancelled), extra,
		StageEventExtras{Reason: reason, OccurredAt: at})
}
