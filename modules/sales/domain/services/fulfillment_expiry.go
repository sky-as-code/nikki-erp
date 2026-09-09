package services

import (
	"time"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// Letting a reservation lapse, so that goods nobody has claimed do not sit unsellable forever.
//
// Expiry is the gentlest of the three ways a hold ends. A cancel undoes the sale; a dispense
// consumes the stock; an expiry does neither — it only stops holding, and the customer keeps
// everything they paid for. So there is NO refund here and no invoice: the entitlement survives, and
// the customer may reserve again at the same or another target whenever they come back.
//
// A fulfillment with no TTL never lapses, which is what a sale dispensed seconds after payment
// wants: it would be absurd to expire an order that was already complete.

// FulfillmentExpiryResult names what one sweep did, so the job can log a count rather than a shrug.
type FulfillmentExpiryResult struct {
	ExpiredFulfillmentIds []string
}

// ExpireLapsedFulfillments releases the stock of every fulfillment whose deadline has passed.
//
// Each is handled on its own and a failure on one does not stop the rest: a reservation Inventory
// could not release this hour is retried next hour, whereas abandoning the sweep would strand every
// fulfillment behind it. Re-running is safe — a fulfillment already moved out of a reservable state
// is skipped, so a second sweep over the same window releases nothing twice.
func ExpireLapsedFulfillments(
	ctx corectx.Context,
	now time.Time,
	limit int,
	reservations itExt.FulfillmentReservationExtService,
) (*FulfillmentExpiryResult, error) {
	result := &FulfillmentExpiryResult{}

	// Only the two states that actually hold stock are swept. `ready` and `in_progress` are excluded
	// deliberately: a paid order clear to dispense, or one with a machine mid-dispense, must not have
	// the goods pulled out from under it by a clock.
	for _, status := range []models.FulfillmentStatus{
		models.FulfillmentStatusReserved,
		models.FulfillmentStatusPendingReservation,
	} {
		records, err := searchBy(ctx, models.SalesOrderFulfillmentSchemaName,
			models.SalesOrderFulfillmentFieldFulfillmentStatus, string(status))
		if err != nil {
			return nil, err
		}

		for _, record := range records {
			fulfillment := models.NewSalesOrderFulfillmentFrom(record)
			if !fulfillment.HasLapsed(now) {
				continue
			}

			fulfillmentId := stringOf(record, models.SalesOrderFulfillmentFieldId)
			if err := ExpireFulfillment(ctx, fulfillmentId, reservations); err != nil {
				// Logged by the caller through the count it does not get. Skipping keeps one
				// unreachable Inventory from stranding every other lapsed hold.
				continue
			}
			result.ExpiredFulfillmentIds = append(result.ExpiredFulfillmentIds, fulfillmentId)

			if limit > 0 && len(result.ExpiredFulfillmentIds) >= limit {
				return result, nil
			}
		}
	}
	return result, nil
}
