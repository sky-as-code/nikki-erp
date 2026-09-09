package services

import (
	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// Where a refund meets a delivery.
//
// Two quantities live at this seam and they are not the same thing. `refunded_qty` counts money that
// has actually gone back, and it alone reduces what a customer is owed. `pending_refund_qty` counts
// money that has been asked for and not yet moved — the customer is still owed something, but not
// the goods, because dispensing against a quantity that is being refunded would hand goods over
// while paying for them in reverse.
//
// Everything here is RECOMPUTED from the refund lines rather than incremented, for the same reason
// the dispensed quantities are: a settlement that ran twice, or a roll-up that failed halfway, is
// corrected by the next recount instead of compounding.

// SyncFulfillmentRefundQuantities recomputes every fulfillment item's refunded quantity from the
// refunds that name it, and moves the item and its fulfillment to whatever the new totals say.
//
// Only SETTLED refunds count. A refund that was raised, or is in flight, has not paid anybody back:
// treating the request as the outcome would close an order nobody has received money for, and would
// leave nothing owed if the refund then failed.
func SyncFulfillmentRefundQuantities(ctx corectx.Context, fulfillmentId string) error {
	refunded, err := settledRefundsByFulfillmentItem(ctx, fulfillmentId)
	if err != nil {
		return err
	}

	items, err := ItemsOfFulfillment(ctx, fulfillmentId)
	if err != nil {
		return err
	}
	for _, record := range items {
		itemId := stringOf(record, models.SalesOrderFulfillmentItemFieldId)
		total := refunded[itemId]

		if decimalOf(record, models.SalesOrderFulfillmentItemFieldRefundedQty).Equal(total) {
			continue
		}
		err := writeChanges(ctx, models.SalesOrderFulfillmentItemSchemaName, record,
			dmodel.DynamicFields{
				models.SalesOrderFulfillmentItemFieldRefundedQty: total,
			})
		if err != nil {
			return err
		}
	}
	return nil
}

// settledRefundsByFulfillmentItem totals what has actually been paid back, keyed by fulfillment item.
//
// A refund line contributes its REFUNDED quantity, not its requested one, and only from a return
// whose refund step reached a settled state. The two conditions are separate on purpose: a return
// can be complete overall while one of its payment legs failed, and the per-line figure is what
// records which money genuinely moved.
func settledRefundsByFulfillmentItem(
	ctx corectx.Context, fulfillmentId string,
) (map[string]decimal.Decimal, error) {
	lines, err := searchBy(ctx, models.SalesReturnLineSchemaName,
		models.SalesReturnLineFieldFulfillmentId, fulfillmentId)
	if err != nil {
		return nil, err
	}

	refunded := map[string]decimal.Decimal{}
	for _, line := range lines {
		itemId := stringOf(line, models.SalesReturnLineFieldFulfillmentItemId)
		if itemId == "" {
			continue
		}
		refunded[itemId] = refunded[itemId].Add(
			decimalOf(line, models.SalesReturnLineFieldRefundedQty))
	}
	return refunded, nil
}

// PendingRefundQuantities totals what has been ASKED for and not yet settled, per fulfillment item.
//
// This is what blocks a retry. A quantity with a refund in flight is still owed to the customer, so
// it is not deducted from what they are due — but it must not be dispensed either, because the money
// may be about to go back and the customer would end up with both.
func PendingRefundQuantities(
	ctx corectx.Context, fulfillmentId string,
) (map[string]decimal.Decimal, error) {
	lines, err := searchBy(ctx, models.SalesReturnLineSchemaName,
		models.SalesReturnLineFieldFulfillmentId, fulfillmentId)
	if err != nil {
		return nil, err
	}

	// Cached per return, because several lines of one refund commonly name the same fulfillment and
	// re-reading the header for each would multiply the queries for one answer.
	terminal := map[string]bool{}
	pending := map[string]decimal.Decimal{}

	for _, line := range lines {
		itemId := stringOf(line, models.SalesReturnLineFieldFulfillmentItemId)
		if itemId == "" {
			continue
		}

		returnId := stringOf(line, models.SalesReturnLineFieldSalesReturnId)
		settled, known := terminal[returnId]
		if !known {
			salesReturn, err := loadRecord(ctx, models.SalesReturnSchemaName,
				models.SalesReturnFieldId, returnId)
			if err != nil {
				return nil, err
			}
			settled = isRefundTerminal(salesReturn)
			terminal[returnId] = settled
		}
		if settled {
			// Settled either way: what succeeded is already in refunded_qty, and what failed released
			// its hold on the quantity, which is then fulfillable again.
			continue
		}

		outstanding := models.NewSalesReturnLineFrom(line).RefundQuantityRequested().
			Sub(decimalOf(line, models.SalesReturnLineFieldRefundedQty))
		if outstanding.IsPositive() {
			pending[itemId] = pending[itemId].Add(outstanding)
		}
	}
	return pending, nil
}

// isRefundTerminal reports that a refund will not move again on its own.
//
// A failed refund is terminal, and deliberately so: the money did not go back, the customer is owed
// the goods again, and the quantity must return to being fulfillable rather than staying blocked
// forever behind a refund that is never going to succeed.
func isRefundTerminal(salesReturn dmodel.DynamicFields) bool {
	if salesReturn == nil {
		// A refund line pointing at a return that no longer exists blocks nothing; treating it as
		// live would strand the quantity permanently.
		return true
	}
	switch models.SalesReturnStepStatus(
		stringOf(salesReturn, models.SalesReturnFieldRefundStatus),
	) {
	case models.SalesReturnStepCompleted,
		models.SalesReturnStepPartiallySucceeded,
		models.SalesReturnStepFailed,
		models.SalesReturnStepCancelled,
		models.SalesReturnStepNotRequired:
		return true
	}
	return false
}

// FulfillableQuantities answers what each item of a fulfillment may be attempted for right now:
// still owed, minus anything with a refund in flight, never below zero.
func FulfillableQuantities(
	ctx corectx.Context, fulfillmentId string,
) (map[string]decimal.Decimal, error) {
	pending, err := PendingRefundQuantities(ctx, fulfillmentId)
	if err != nil {
		return nil, err
	}
	items, err := ItemsOfFulfillment(ctx, fulfillmentId)
	if err != nil {
		return nil, err
	}

	fulfillable := make(map[string]decimal.Decimal, len(items))
	for _, record := range items {
		itemId := stringOf(record, models.SalesOrderFulfillmentItemFieldId)
		remaining := models.NewSalesOrderFulfillmentItemFrom(record).RemainingQuantity()

		available := remaining.Sub(pending[itemId])
		if available.IsNegative() {
			available = decimal.Zero
		}
		fulfillable[itemId] = available
	}
	return fulfillable, nil
}
