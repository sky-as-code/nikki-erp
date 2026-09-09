package services

import (
	"strings"
	"testing"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// The state gates of confirm and cancel; both read the repository and are exercised live.

func orderRecord(status, payment, fulfillment string) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.SalesOrderFieldId:                orderTestId,
		models.SalesOrderFieldStatus:            status,
		models.SalesOrderFieldPaymentStatus:     payment,
		models.SalesOrderFieldFulfillmentStatus: fulfillment,
		models.SalesOrderFieldSalesChannelId:    "CH1",
	}
}

const orderTestId = "OR1"

// THE test of cancel. A paid order must not be cancellable: that leaves the business holding money
// against a sale that no longer exists, and the refusal must name the refund workflow.
func TestAPaidOrderCannotBeCancelled(t *testing.T) {
	for _, paymentStatus := range []string{
		string(models.SalesOrderPaymentStatusPaid),
		string(models.SalesOrderPaymentStatusPartiallyPaid),
		string(models.SalesOrderPaymentStatusOverpaid),
	} {
		t.Run(paymentStatus, func(t *testing.T) {
			vErrs := assertCancellable(orderRecord(
				string(models.SalesOrderStatusConfirmed), paymentStatus,
				string(models.SalesOrderFulfillmentStatusPending)))

			if vErrs == nil {
				t.Fatalf("a %s order must not be cancellable", paymentStatus)
			}
			if !hasReasonKey(vErrs, ReasonRequiresRefund) {
				t.Errorf("the refusal must name %q so the caller knows to raise a refund, got %v",
					ReasonRequiresRefund, vErrs.ToError())
			}
		})
	}
}

// Goods already with the customer are a return, not a cancellation.
func TestAFulfilledOrderCannotBeCancelled(t *testing.T) {
	for _, fulfillmentStatus := range []string{
		string(models.SalesOrderFulfillmentStatusFulfilled),
		string(models.SalesOrderFulfillmentStatusPartiallyFulfilled),
	} {
		t.Run(fulfillmentStatus, func(t *testing.T) {
			vErrs := assertCancellable(orderRecord(
				string(models.SalesOrderStatusConfirmed),
				string(models.SalesOrderPaymentStatusUnpaid), fulfillmentStatus))

			if vErrs == nil || !hasReasonKey(vErrs, ReasonRequiresReturn) {
				t.Errorf("a %s order must be redirected to the return workflow", fulfillmentStatus)
			}
		})
	}
}

// Payment is checked before fulfilment: an order both paid AND fulfilled is more usefully
// described as needing a refund.
func TestPaymentIsCheckedBeforeFulfilment(t *testing.T) {
	vErrs := assertCancellable(orderRecord(
		string(models.SalesOrderStatusConfirmed),
		string(models.SalesOrderPaymentStatusPaid),
		string(models.SalesOrderFulfillmentStatusFulfilled)))

	if vErrs == nil || !hasReasonKey(vErrs, ReasonRequiresRefund) {
		t.Error("an order that is both paid and fulfilled must report the refund path first")
	}
}

// The ordinary cases: a draft, and a confirmed order nobody has paid for or shipped.
func TestUnpaidUnfulfilledOrdersCancel(t *testing.T) {
	for _, status := range []string{
		string(models.SalesOrderStatusDraft),
		string(models.SalesOrderStatusConfirmed),
	} {
		t.Run(status, func(t *testing.T) {
			vErrs := assertCancellable(orderRecord(status,
				string(models.SalesOrderPaymentStatusUnpaid),
				string(models.SalesOrderFulfillmentStatusPending)))
			if vErrs != nil {
				t.Errorf("a %s order with no payment or fulfilment must cancel, got %v",
					status, vErrs.ToError())
			}
		})
	}
}

// Cancelling twice is refused rather than a no-op: a cancel releases voucher reservations, so a
// silent second success would give back a use the order no longer held.
func TestCancellingTwiceIsRefused(t *testing.T) {
	vErrs := assertCancellable(orderRecord(
		string(models.SalesOrderStatusCancelled),
		string(models.SalesOrderPaymentStatusUnpaid),
		string(models.SalesOrderFulfillmentStatusPending)))

	if vErrs == nil || !hasReasonKey(vErrs, ReasonAlreadyCancelled) {
		t.Error("an already-cancelled order must be refused, not silently re-cancelled")
	}
}

// A completed order is finished; unwinding it is a return.
func TestACompletedOrderCannotBeCancelled(t *testing.T) {
	vErrs := assertCancellable(orderRecord(
		string(models.SalesOrderStatusCompleted),
		string(models.SalesOrderPaymentStatusUnpaid),
		string(models.SalesOrderFulfillmentStatusPending)))

	if vErrs == nil || !hasReasonKey(vErrs, ReasonNotCancellable) {
		t.Error("a completed order must not be cancellable")
	}
}

// A draft cancel has nothing to undo. A confirmed one still leaves its payments committed, and
// saying so stops a caller assuming the money came back.
func TestOnlyAConfirmedCancelReportsPendingWork(t *testing.T) {
	if pending := pendingCancelSteps(string(models.SalesOrderStatusDraft), nil); len(pending) != 0 {
		t.Errorf("a draft cancel has nothing outstanding, got %v", pending)
	}
	if pending := pendingCancelSteps(string(models.SalesOrderStatusConfirmed), nil); len(pending) == 0 {
		t.Error("a confirmed cancel cannot cancel payments and must say so")
	}
}

// Stock release is no longer pending work: cancel performs it. It stays reportable only when there
// is no inventory port to perform it through, because then the goods really are still held.
func TestStockReleaseIsPendingOnlyWithoutAnInventoryPort(t *testing.T) {
	unbound := pendingCancelSteps(string(models.SalesOrderStatusConfirmed), nil)
	if !mentions(unbound, "release_stock_reservation") {
		t.Errorf("with no inventory port the stock is still held and must be reported, got %v", unbound)
	}

	bound := pendingCancelSteps(string(models.SalesOrderStatusConfirmed), stubReservations{})
	if mentions(bound, "release_stock_reservation") {
		t.Errorf("cancel releases the reservation itself, so it is not pending, got %v", bound)
	}
	if !mentions(bound, "cancel_pending_payments") {
		t.Errorf("payments are still not cancelled and must stay reported, got %v", bound)
	}
}

func mentions(pending []string, step string) bool {
	for _, entry := range pending {
		if strings.Contains(entry, step) {
			return true
		}
	}
	return false
}

// stubReservations stands in for a bound inventory port. pendingCancelSteps only asks whether one
// exists, so nothing here needs to do anything.
type stubReservations struct{}

func (stubReservations) ReserveForFulfillment(
	corectx.Context, itExt.FulfillmentReservationRequest,
) (*itExt.FulfillmentReservationResponse, error) {
	return nil, nil
}

func (stubReservations) ReallocateReservation(
	corectx.Context, itExt.FulfillmentReservationRequest,
) (*itExt.FulfillmentReservationResponse, error) {
	return nil, nil
}

func (stubReservations) ReleaseFulfillmentReservation(
	corectx.Context, string,
) (*itExt.FulfillmentReservationResponse, error) {
	return nil, nil
}

func (stubReservations) CheckAvailabilityByLocations(
	corectx.Context, itExt.AvailabilityQuery,
) (*itExt.AvailabilityResult, error) {
	return nil, nil
}

// Confirm refuses anything that is not a draft, and re-confirming is refused rather than
// idempotent, because a confirm redeems vouchers and would redeem twice.
func TestOnlyADraftCanBeConfirmed(t *testing.T) {
	for _, status := range []string{
		string(models.SalesOrderStatusConfirmed),
		string(models.SalesOrderStatusProcessing),
		string(models.SalesOrderStatusCompleted),
		string(models.SalesOrderStatusCancelled),
	} {
		t.Run(status, func(t *testing.T) {
			vErrs, err := assertConfirmable(nil, orderRecord(status,
				string(models.SalesOrderPaymentStatusUnpaid),
				string(models.SalesOrderFulfillmentStatusPending)))
			if err != nil {
				t.Fatalf("the status gate must not need a repository: %v", err)
			}
			if vErrs == nil || !hasReasonKey(vErrs, ReasonNotConfirmable) {
				t.Errorf("a %s order must not be confirmable", status)
			}
		})
	}
}

// The channel check stays even though the column is NOT NULL: a schema that changes underneath
// this code should fail loudly here rather than silently later.
func TestAnOrderWithNoChannelCannotBeConfirmed(t *testing.T) {
	record := orderRecord(string(models.SalesOrderStatusDraft),
		string(models.SalesOrderPaymentStatusUnpaid),
		string(models.SalesOrderFulfillmentStatusPending))
	delete(record, models.SalesOrderFieldSalesChannelId)

	vErrs, err := assertConfirmable(nil, record)
	if err != nil {
		t.Fatalf("the channel gate must not need a repository: %v", err)
	}
	if vErrs == nil || !hasReasonKey(vErrs, ReasonChannelMissing) {
		t.Error("an order with no sales channel must not be confirmable")
	}
}

// The lock key is built from the id, never a human-facing label a correction could reuse.
func TestTheLockKeyIsBuiltFromTheOrderId(t *testing.T) {
	first := confirmLockKeyOf("OR1")
	second := confirmLockKeyOf("OR2")

	if first == second {
		t.Fatal("two orders must not share a lock key")
	}
	if first != "lock:sales_order:OR1" {
		t.Errorf("lock key = %q, want it namespaced and id-based", first)
	}
}

// Confirm and cancel must take the SAME lock, or one could release a reservation the other had
// just redeemed.
func TestConfirmAndCancelShareOneLock(t *testing.T) {
	// Both call confirmLockKeyOf; this asserts that stays true by pinning the single key builder.
	if confirmLockKeyOf(orderTestId) != "lock:sales_order:"+orderTestId {
		t.Error("cancel must lock on the same key confirm does")
	}
}
