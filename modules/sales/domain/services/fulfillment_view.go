package services

import (
	"github.com/shopspring/decimal"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// What a caller is shown about an order's deliveries, and which places could take one on.
//
// Two rules govern the shape of the first view. Refund state never appears inside
// fulfillment_status — a refund has its own lifecycle, and folding it in would make a failed refund
// silently reopen a delivery that was never going to happen again. And the quantities distinguish
// what is still owed from what may be attempted RIGHT NOW: a quantity with a refund in flight is
// owed but not attemptable, and a client that could not tell them apart would offer a customer a
// retry that is about to be refunded underneath them.

// FulfillmentView is one delivery as a caller sees it.
type FulfillmentView struct {
	FulfillmentId  string
	MethodId       string
	Type           string
	TargetOutletId string
	Status         string

	// ReservationExpiresAt is RFC3339, or empty when the policy sets no TTL.
	ReservationExpiresAt string

	Items []FulfillmentItemView
}

// FulfillmentItemView is one product within it.
type FulfillmentItemView struct {
	ItemId           string
	SalesOrderLineId string
	ProductVariantId string
	Status           string

	// SourceLocationId is where this item's stock is held, when the target has addressable places
	// rather than one pool. Empty means the fulfillment's single target location.
	SourceLocationId string

	// InventorySourceId is the key THIS item's hold is found by in Inventory, so an executor can
	// consume or release exactly the hold its result concerns.
	InventorySourceId string

	// InventoryReservationRef is the warehouse reservation holding this item, when the goods are
	// held at warehouse level rather than at a slot. An executor reporting a dispense consumes
	// it, naming the slot the goods actually left from.
	InventoryReservationRef string

	OrderedQty   decimal.Decimal
	FulfilledQty decimal.Decimal
	RefundedQty  decimal.Decimal

	// RemainingQty is what the customer is still owed: ordered minus delivered minus successfully
	// refunded. Completion turns on this alone.
	RemainingQty decimal.Decimal

	// PendingRefundQty is quantity with a refund raised but not yet settled. It is owed and NOT
	// attemptable: the money may be about to go back, and dispensing against it would hand over
	// goods that are simultaneously being paid for in reverse.
	//
	// Always zero until the refund lifecycle lands; the field exists now so the response shape does
	// not change under clients when it does.
	PendingRefundQty decimal.Decimal

	// FulfillableQty is what a new attempt may ask for: remaining minus pending refund, never below
	// zero. This is the number a retry is validated against, not RemainingQty.
	FulfillableQty decimal.Decimal
}

// ViewOrderFulfillments assembles the deliveries of one order.
func ViewOrderFulfillments(
	ctx corectx.Context, orderId string,
) ([]FulfillmentView, error) {
	records, err := FulfillmentsOfOrder(ctx, orderId)
	if err != nil {
		return nil, err
	}

	views := make([]FulfillmentView, 0, len(records))
	for _, record := range records {
		fulfillmentId := stringOf(record, models.SalesOrderFulfillmentFieldId)

		itemRecords, err := ItemsOfFulfillment(ctx, fulfillmentId)
		if err != nil {
			return nil, err
		}

		pendingRefunds, err := PendingRefundQuantities(ctx, fulfillmentId)
		if err != nil {
			return nil, err
		}

		items := make([]FulfillmentItemView, 0, len(itemRecords))
		for _, itemRecord := range itemRecords {
			item := models.NewSalesOrderFulfillmentItemFrom(itemRecord)
			remaining := item.RemainingQuantity()
			pendingRefund := pendingRefunds[stringOf(itemRecord, models.SalesOrderFulfillmentItemFieldId)]

			fulfillable := remaining.Sub(pendingRefund)
			if fulfillable.IsNegative() {
				fulfillable = decimal.Zero
			}

			sourceLocationId := stringOf(itemRecord,
				models.SalesOrderFulfillmentItemFieldSourceLocationId)
			inventorySourceId := reservationSourceId(fulfillmentId, sourceLocationId, "")

			items = append(items, FulfillmentItemView{
				ItemId:            stringOf(itemRecord, models.SalesOrderFulfillmentItemFieldId),
				SalesOrderLineId:  stringOf(itemRecord, models.SalesOrderFulfillmentItemFieldSalesOrderLineId),
				ProductVariantId:  stringOf(itemRecord, models.SalesOrderFulfillmentItemFieldProductVariantId),
				Status:            stringOf(itemRecord, models.SalesOrderFulfillmentItemFieldItemStatus),
				OrderedQty:        decimalOf(itemRecord, models.SalesOrderFulfillmentItemFieldOrderedQty),
				FulfilledQty:      decimalOf(itemRecord, models.SalesOrderFulfillmentItemFieldFulfilledQty),
				RefundedQty:       decimalOf(itemRecord, models.SalesOrderFulfillmentItemFieldRefundedQty),
				RemainingQty:      remaining,
				PendingRefundQty:  pendingRefund,
				FulfillableQty:    fulfillable,
				SourceLocationId:  sourceLocationId,
				InventorySourceId: inventorySourceId,
				InventoryReservationRef: stringOf(itemRecord,
					models.SalesOrderFulfillmentItemFieldInventoryReservationRef),
			})
		}

		view := FulfillmentView{
			FulfillmentId:  fulfillmentId,
			MethodId:       stringOf(record, models.SalesOrderFulfillmentFieldFulfillmentMethodId),
			Type:           stringOf(record, models.SalesOrderFulfillmentFieldFulfillmentType),
			TargetOutletId: stringOf(record, models.SalesOrderFulfillmentFieldTargetOutletId),
			Status:         stringOf(record, models.SalesOrderFulfillmentFieldFulfillmentStatus),
			Items:          items,
		}
		if expiresAt := dateTimeOf(record, models.SalesOrderFulfillmentFieldReservationExpiresAt); expiresAt != nil {
			view.ReservationExpiresAt = expiresAt.String()
		}
		views = append(views, view)
	}
	return views, nil
}

// FulfillmentTargetOption is one place that could take a delivery on.
type FulfillmentTargetOption struct {
	SalesPointId      string
	InventoryLocation string
	CanFulfillAll     bool
	Shortages         []itExt.AvailabilityShortage
}

// SearchFulfillmentTargets lists the fulfillment-enabled sales points of an org and says which of
// them could supply a basket.
//
// ADVISORY, and the response says so: it takes no lock, so a kiosk reported able to supply may be
// emptied by another sale before the customer chooses. Only a reservation secures stock. Points
// with no inventory location are dropped rather than reported unable — they are not candidates at
// all, and listing them would invite a caller to retry against something that can never work.
func SearchFulfillmentTargets(
	ctx corectx.Context,
	orgId string,
	items []itExt.AvailabilityItem,
	availability itExt.FulfillmentReservationExtService,
) ([]FulfillmentTargetOption, error) {
	points, err := searchBy(ctx, models.SalesPointSchemaName,
		models.SalesPointFieldFulfillmentEnabled, "true")
	if err != nil {
		return nil, err
	}

	locationByPoint := make(map[string]string, len(points))
	locationIds := make([]string, 0, len(points))
	for _, point := range points {
		locationId := stringOf(point, models.SalesPointFieldInventoryLocationId)
		if locationId == "" {
			continue
		}
		locationByPoint[stringOf(point, models.SalesPointFieldId)] = locationId
		locationIds = append(locationIds, locationId)
	}
	if len(locationIds) == 0 || availability == nil || len(items) == 0 {
		return []FulfillmentTargetOption{}, nil
	}

	result, err := availability.CheckAvailabilityByLocations(ctx, itExt.AvailabilityQuery{
		OrgId:       orgId,
		LocationIds: locationIds,
		Items:       items,
	})
	if err != nil {
		return nil, err
	}

	byLocation := make(map[string]itExt.LocationAvailability, len(result.Locations))
	for _, location := range result.Locations {
		byLocation[location.LocationId] = location
	}

	options := make([]FulfillmentTargetOption, 0, len(locationByPoint))
	for pointId, locationId := range locationByPoint {
		location := byLocation[locationId]
		options = append(options, FulfillmentTargetOption{
			SalesPointId:      pointId,
			InventoryLocation: locationId,
			CanFulfillAll:     location.CanFulfillAll,
			Shortages:         location.Shortages,
		})
	}
	return options, nil
}

// RefundView is one refund as a caller sees it, and the shape doc 05 §16/§17 asks for.
type RefundView struct {
	RefundId     string
	RefundStatus string
	RefundReason string
	ReturnType   string
	RefundTotal  decimal.Decimal

	Items []RefundItemView
}

// RefundItemView is one line of it. Requested and refunded are both reported because they answer
// different questions: what the customer asked for, and what they have actually been paid.
type RefundItemView struct {
	SalesOrderLineId  string
	FulfillmentId     string
	FulfillmentItemId string

	RequestedQty decimal.Decimal
	RefundedQty  decimal.Decimal
}

// ViewOrderRefunds lists an order's refunds and what each has actually paid back.
func ViewOrderRefunds(ctx corectx.Context, orderId string) ([]RefundView, error) {
	returns, err := searchBy(ctx, models.SalesReturnSchemaName,
		models.SalesReturnFieldSalesOrderId, orderId)
	if err != nil {
		return nil, err
	}

	views := make([]RefundView, 0, len(returns))
	for _, record := range returns {
		returnId := stringOf(record, models.SalesReturnFieldId)

		lines, err := returnLinesOf(ctx, returnId)
		if err != nil {
			return nil, err
		}

		items := make([]RefundItemView, 0, len(lines))
		for _, line := range lines {
			items = append(items, RefundItemView{
				SalesOrderLineId:  stringOf(line, models.SalesReturnLineFieldSalesOrderLineId),
				FulfillmentId:     stringOf(line, models.SalesReturnLineFieldFulfillmentId),
				FulfillmentItemId: stringOf(line, models.SalesReturnLineFieldFulfillmentItemId),
				RequestedQty:      models.NewSalesReturnLineFrom(line).RefundQuantityRequested(),
				RefundedQty:       decimalOf(line, models.SalesReturnLineFieldRefundedQty),
			})
		}

		views = append(views, RefundView{
			RefundId:     returnId,
			RefundStatus: stringOf(record, models.SalesReturnFieldRefundStatus),
			RefundReason: stringOf(record, models.SalesReturnFieldRefundReason),
			ReturnType:   stringOf(record, models.SalesReturnFieldReturnType),
			RefundTotal:  decimalOf(record, models.SalesReturnFieldRefundTotal),
			Items:        items,
		})
	}
	return views, nil
}

// ViewFulfillment assembles ONE delivery, for a caller that already knows which one it means — a
// kiosk about to dispense, rather than an agent looking at an order. Nil when there is no such
// fulfillment, which a caller reads as "nothing to attempt" rather than as a fault.
func ViewFulfillment(ctx corectx.Context, fulfillmentId string) (*FulfillmentView, error) {
	record, err := loadRecord(ctx, models.SalesOrderFulfillmentSchemaName,
		models.SalesOrderFulfillmentFieldId, fulfillmentId)
	if err != nil || record == nil {
		return nil, err
	}

	pendingRefunds, err := PendingRefundQuantities(ctx, fulfillmentId)
	if err != nil {
		return nil, err
	}
	itemRecords, err := ItemsOfFulfillment(ctx, fulfillmentId)
	if err != nil {
		return nil, err
	}

	items := make([]FulfillmentItemView, 0, len(itemRecords))
	for _, itemRecord := range itemRecords {
		item := models.NewSalesOrderFulfillmentItemFrom(itemRecord)
		itemId := stringOf(itemRecord, models.SalesOrderFulfillmentItemFieldId)

		remaining := item.RemainingQuantity()
		pendingRefund := pendingRefunds[itemId]
		fulfillable := remaining.Sub(pendingRefund)
		if fulfillable.IsNegative() {
			fulfillable = decimal.Zero
		}

		sourceLocationId := stringOf(itemRecord,
			models.SalesOrderFulfillmentItemFieldSourceLocationId)
		inventorySourceId := reservationSourceId(fulfillmentId, sourceLocationId, "")

		items = append(items, FulfillmentItemView{
			ItemId:            itemId,
			SalesOrderLineId:  stringOf(itemRecord, models.SalesOrderFulfillmentItemFieldSalesOrderLineId),
			ProductVariantId:  stringOf(itemRecord, models.SalesOrderFulfillmentItemFieldProductVariantId),
			Status:            stringOf(itemRecord, models.SalesOrderFulfillmentItemFieldItemStatus),
			OrderedQty:        decimalOf(itemRecord, models.SalesOrderFulfillmentItemFieldOrderedQty),
			FulfilledQty:      decimalOf(itemRecord, models.SalesOrderFulfillmentItemFieldFulfilledQty),
			RefundedQty:       decimalOf(itemRecord, models.SalesOrderFulfillmentItemFieldRefundedQty),
			RemainingQty:      remaining,
			PendingRefundQty:  pendingRefund,
			FulfillableQty:    fulfillable,
			SourceLocationId:  sourceLocationId,
			InventorySourceId: inventorySourceId,
			InventoryReservationRef: stringOf(itemRecord,
				models.SalesOrderFulfillmentItemFieldInventoryReservationRef),
		})
	}

	return &FulfillmentView{
		FulfillmentId:  fulfillmentId,
		MethodId:       stringOf(record, models.SalesOrderFulfillmentFieldFulfillmentMethodId),
		Type:           stringOf(record, models.SalesOrderFulfillmentFieldFulfillmentType),
		TargetOutletId: stringOf(record, models.SalesOrderFulfillmentFieldTargetOutletId),
		Status:         stringOf(record, models.SalesOrderFulfillmentFieldFulfillmentStatus),
		Items:          items,
	}, nil
}
