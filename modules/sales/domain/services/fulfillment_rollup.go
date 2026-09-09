package services

import (
	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// Propagating what a delivery did up to the order that owns it.
//
// The fulfillment tables are where a kiosk sale's truth lives, but three older consumers ask the
// ORDER instead: cancel-vs-return gating reads sales_orders.fulfillment_status, e-invoice eligibility
// reads it too, and return validation reads sales_order_lines.fulfilled_quantity. Without this,
// goods that physically left a machine leave every one of those reading zero — so a dispensed order
// could be cancelled outright instead of returned, would never have its invoice issued, and would
// refuse a customer the return of goods they are holding.
//
// Both quantities are RECOMPUTED from their sources rather than incremented, so a result delivered
// twice cannot double-count and a roll-up that failed halfway is corrected by the next one.

// SyncOrderFulfillmentRollup recomputes an order's delivered quantities and fulfillment status from
// whatever actually delivered them, and writes both.
func SyncOrderFulfillmentRollup(ctx corectx.Context, orderId string) error {
	if err := SyncFulfilledQuantities(ctx, orderId); err != nil {
		return err
	}

	status, err := DeriveOrderFulfillmentStatus(ctx, orderId)
	if err != nil {
		return err
	}
	if status == "" {
		return nil
	}

	order, err := loadRecord(ctx, models.SalesOrderSchemaName, models.SalesOrderFieldId, orderId)
	if err != nil || order == nil {
		return err
	}
	if stringOf(order, models.SalesOrderFieldFulfillmentStatus) == status {
		return nil
	}
	return writeChanges(ctx, models.SalesOrderSchemaName, order, dmodel.DynamicFields{
		models.SalesOrderFieldFulfillmentStatus: status,
	})
}

// deliveredByOrderLine totals what the fulfillment tables say reached the customer, keyed by order
// line. This is the kiosk half of the roll-up: a kiosk sale raises no fulfillment REQUEST, so the
// older source below sees nothing for it.
func deliveredByOrderLine(
	ctx corectx.Context, orderId string,
) (map[string]decimal.Decimal, error) {
	fulfillments, err := FulfillmentsOfOrder(ctx, orderId)
	if err != nil {
		return nil, err
	}

	delivered := map[string]decimal.Decimal{}
	for _, fulfillment := range fulfillments {
		if models.FulfillmentStatus(
			stringOf(fulfillment, models.SalesOrderFulfillmentFieldFulfillmentStatus),
		) == models.FulfillmentStatusCancelled {
			// A cancelled fulfillment delivered nothing; its items keep their quantities only as
			// history, and counting them would report goods that never left.
			continue
		}

		items, err := ItemsOfFulfillment(ctx,
			stringOf(fulfillment, models.SalesOrderFulfillmentFieldId))
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			lineId := stringOf(item, models.SalesOrderFulfillmentItemFieldSalesOrderLineId)
			if lineId == "" {
				continue
			}
			delivered[lineId] = delivered[lineId].Add(
				decimalOf(item, models.SalesOrderFulfillmentItemFieldFulfilledQty))
		}
	}
	return delivered, nil
}
