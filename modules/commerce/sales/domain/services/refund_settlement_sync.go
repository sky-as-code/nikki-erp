package services

import (
	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// Turning settled money into settled quantities.
//
// A refund is asked for in quantities and paid in amounts, across however many payment legs funded
// the order. This is where the two meet: once the legs have settled, each refund line learns how much
// of what it asked for was actually paid back, and the fulfillment it points at learns that the
// customer no longer owes those goods.
//
// The rule that governs all of it: creation is not success. A raised refund changes nothing about
// what a customer is owed, and only money that actually moved reduces it. Doing otherwise would close
// an order nobody has been paid back for, and leave nothing owed if the refund then failed.

// SyncRefundedQuantities records how much of each refund line was actually paid, and propagates the
// result to any fulfillment the line names.
//
// The proportion settled is applied uniformly across the return's lines rather than line by line,
// because a payment leg funds the ORDER and not any particular line: when one of two legs fails
// there is no fact of the matter about which line went unpaid, and inventing one would tell a
// customer they were refunded for the wrong item.
func SyncRefundedQuantities(ctx corectx.Context, salesReturn dmodel.DynamicFields) error {
	returnId := stringOf(salesReturn, models.SalesReturnFieldId)

	lines, err := returnLinesOf(ctx, returnId)
	if err != nil {
		return err
	}
	if len(lines) == 0 {
		return nil
	}

	settledShare, err := refundSettledShare(ctx, salesReturn)
	if err != nil {
		return err
	}

	fulfillments := map[string]struct{}{}
	for _, line := range lines {
		requested := models.NewSalesReturnLineFrom(line).RefundQuantityRequested()
		refunded := requested.Mul(settledShare)

		// Guard against a rounding artefact reporting more paid back than was ever asked for.
		if refunded.GreaterThan(requested) {
			refunded = requested
		}

		if !decimalOf(line, models.SalesReturnLineFieldRefundedQty).Equal(refunded) {
			err := writeChanges(ctx, models.SalesReturnLineSchemaName, line, dmodel.DynamicFields{
				models.SalesReturnLineFieldRefundedQty: refunded,
			})
			if err != nil {
				return err
			}
		}

		if fulfillmentId := stringOf(line, models.SalesReturnLineFieldFulfillmentId); fulfillmentId != "" {
			fulfillments[fulfillmentId] = struct{}{}
		}
	}

	for fulfillmentId := range fulfillments {
		if err := SyncFulfillmentRefundQuantities(ctx, fulfillmentId); err != nil {
			return err
		}
		if err := SettleFulfillmentAfterRefund(ctx, fulfillmentId); err != nil {
			return err
		}
	}
	return nil
}

// refundSettledShare is the fraction of the refund that actually reached the customer, as a ratio of
// completed leg amounts to the total asked for.
//
// It reads the legs rather than the return's step status because the status is a summary: a return
// can be `partially_succeeded` at several different proportions, and a customer owed three of five
// units back is not the same as one owed four.
func refundSettledShare(
	ctx corectx.Context, salesReturn dmodel.DynamicFields,
) (decimal.Decimal, error) {
	requested := decimalOf(salesReturn, models.SalesReturnFieldRefundTotal)
	if !requested.IsPositive() {
		return decimal.Zero, nil
	}

	legs, err := searchBy(ctx, models.SalesRefundPaymentSchemaName,
		models.SalesRefundPaymentFieldSalesReturnId,
		stringOf(salesReturn, models.SalesReturnFieldId))
	if err != nil {
		return decimal.Zero, err
	}

	settled := decimal.Zero
	for _, leg := range legs {
		if models.SalesRefundPaymentStatus(
			stringOf(leg, models.SalesRefundPaymentFieldStatus),
		) != models.SalesRefundPaymentStatusCompleted {
			// Pending money has not moved and failed money never will. Neither reduces what the
			// customer is owed.
			continue
		}
		settled = settled.Add(decimalOf(leg, models.SalesRefundPaymentFieldAmount))
	}

	if settled.GreaterThanOrEqual(requested) {
		return decimal.NewFromInt(1), nil
	}
	return settled.Div(requested), nil
}

// SettleFulfillmentAfterRefund moves a fulfillment to whatever its quantities now say, once a refund
// against it has settled or failed.
//
// Completion turns on outstanding quantity ALONE (doc 05 §13): an item whose shortfall was refunded
// owes nothing, exactly as one that was delivered owes nothing, and the fulfillment completes on
// either. A REFUND's own failure never makes a fulfillment `failed` (doc 05 §12/§15) — the goods are
// simply owed again, the quantity becomes fulfillable, and the fulfillment goes back to waiting for
// somebody to decide what happens next.
func SettleFulfillmentAfterRefund(ctx corectx.Context, fulfillmentId string) error {
	fulfillment, err := loadRecord(ctx, models.SalesOrderFulfillmentSchemaName,
		models.SalesOrderFulfillmentFieldId, fulfillmentId)
	if err != nil || fulfillment == nil {
		return err
	}
	if models.NewSalesOrderFulfillmentFrom(fulfillment).IsTerminal() {
		// Already completed or cancelled. A late refund settlement does not reopen it.
		return nil
	}

	items, err := ItemsOfFulfillment(ctx, fulfillmentId)
	if err != nil {
		return err
	}

	outstanding := decimal.Zero
	for _, record := range items {
		outstanding = outstanding.Add(
			models.NewSalesOrderFulfillmentItemFrom(record).RemainingQuantity())
	}

	if outstanding.IsZero() {
		if err := setFulfillmentStatus(ctx, fulfillmentId, models.FulfillmentStatusCompleted); err != nil {
			return err
		}
	} else if stringOf(fulfillment, models.SalesOrderFulfillmentFieldFulfillmentStatus) ==
		string(models.FulfillmentStatusCompleted) {
		// A failed refund put quantity back into play on a fulfillment that had been settled by it.
		// It waits rather than being attempted automatically: somebody has to decide whether to try
		// the goods again or refund once more.
		if err := setFulfillmentStatus(
			ctx, fulfillmentId, models.FulfillmentStatusWaitingCustomerAction); err != nil {
			return err
		}
	}

	// The order's own rollup last, so cancel-vs-return gating and invoice eligibility see the same
	// picture the fulfillment does.
	return SyncOrderFulfillmentRollup(ctx,
		stringOf(fulfillment, models.SalesOrderFulfillmentFieldSalesOrderId))
}
