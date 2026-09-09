package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// Giving held stock back, and the shared reads every fulfillment flow needs.
//
// Releasing is idempotent from end to end, and every caller depends on that: a cancel, an expiry
// sweep and an operator can all decide the same hold should go, and none of them should fail because
// another got there first. A release that finds nothing to release has achieved what it was asked to.

// FulfillmentsOfOrder lists an order's fulfillments. An order commonly has exactly one; it may hold
// several when a sale was split across targets, or when an expired fulfillment was replaced rather
// than revived, so callers iterate rather than assuming.
func FulfillmentsOfOrder(
	ctx corectx.Context, orderId string,
) ([]dmodel.DynamicFields, error) {
	return searchBy(ctx, models.SalesOrderFulfillmentSchemaName,
		models.SalesOrderFulfillmentFieldSalesOrderId, orderId)
}

// ItemsOfFulfillment lists one fulfillment's items.
func ItemsOfFulfillment(
	ctx corectx.Context, fulfillmentId string,
) ([]dmodel.DynamicFields, error) {
	return searchBy(ctx, models.SalesOrderFulfillmentItemSchemaName,
		models.SalesOrderFulfillmentItemFieldFulfillmentId, fulfillmentId)
}

// ReleaseOrderFulfillments gives back the stock held for every non-terminal fulfillment of an order
// and marks them cancelled. It is what a cancel calls, and it names the fulfillments it released so
// the caller can report them rather than leaving a customer to guess whether the goods came back.
//
// A nil port releases nothing and reports nothing, matching how the rest of the module treats an
// unbound Inventory: the sale is real either way, and refusing a cancel because a port is missing
// would leave the order stuck in a state nobody can leave.
func ReleaseOrderFulfillments(
	ctx corectx.Context,
	orderId string,
	reservations itExt.FulfillmentReservationExtService,
) (releasedIds []string, err error) {
	fulfillments, err := FulfillmentsOfOrder(ctx, orderId)
	if err != nil {
		return nil, err
	}

	releasedIds = make([]string, 0, len(fulfillments))
	for _, record := range fulfillments {
		fulfillment := models.NewSalesOrderFulfillmentFrom(record)
		if fulfillment.IsTerminal() {
			// Already completed or cancelled. A completed fulfillment's stock has left the building
			// and there is nothing to give back; releasing against it would be a no-op that
			// nonetheless reads, in the result, as if goods returned.
			continue
		}

		fulfillmentId := stringOf(record, models.SalesOrderFulfillmentFieldId)
		if reservations != nil {
			// The refusal is deliberately not propagated. Inventory answering "no such reservation"
			// means the hold is gone, which is the end state this asked for; any other refusal still
			// must not keep an order the customer cancelled in a state they cannot leave.
			if _, err := reservations.ReleaseFulfillmentReservation(ctx, fulfillmentId); err != nil {
				return nil, err
			}
		}

		if err := setFulfillmentStatus(ctx, fulfillmentId, models.FulfillmentStatusCancelled); err != nil {
			return nil, err
		}
		if err := clearReservationRefs(ctx, fulfillmentId); err != nil {
			return nil, err
		}
		releasedIds = append(releasedIds, fulfillmentId)
	}
	return releasedIds, nil
}

// ExpireFulfillment releases a lapsed hold and marks the fulfillment expired.
//
// Expiry is NOT a cancellation and never refunds: the customer keeps what they paid for, and may
// reserve again at the same or another target. Only the stock goes back, so that goods nobody has
// claimed do not sit unsellable forever.
func ExpireFulfillment(
	ctx corectx.Context,
	fulfillmentId string,
	reservations itExt.FulfillmentReservationExtService,
) error {
	if reservations != nil {
		if _, err := reservations.ReleaseFulfillmentReservation(ctx, fulfillmentId); err != nil {
			return err
		}
	}
	if err := setFulfillmentStatus(ctx, fulfillmentId, models.FulfillmentStatusExpired); err != nil {
		return err
	}
	return clearReservationRefs(ctx, fulfillmentId)
}

// clearReservationRefs drops the inventory references from a fulfillment's items once its hold is
// gone. Left in place they would point at a released document, and an operator tracing an item
// would be shown a reservation that no longer holds anything.
func clearReservationRefs(ctx corectx.Context, fulfillmentId string) error {
	items, err := ItemsOfFulfillment(ctx, fulfillmentId)
	if err != nil {
		return err
	}
	for _, item := range items {
		if stringOf(item, models.SalesOrderFulfillmentItemFieldInventoryReservationRef) == "" {
			continue
		}
		err := writeChanges(ctx, models.SalesOrderFulfillmentItemSchemaName, item, dmodel.DynamicFields{
			models.SalesOrderFulfillmentItemFieldInventoryReservationRef: nil,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
