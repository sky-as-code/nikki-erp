package services

import (
	"github.com/shopspring/decimal"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// Comparing what a client thought a sale cost against what Sales charged.
//
// A client may send the prices it calculated locally. They are never charged and never corrected:
// Sales prices the order from its own pricelists, and the estimate is kept only so a device showing
// stale prices can be found before customers start noticing the difference.
//
// A divergence therefore does NOT fail the sale. It is recorded and nothing more, because the
// alternative — refusing an order the customer is standing in front of — would turn a reporting
// problem into a lost sale.

// RecordEstimatedPriceDivergence writes an integration event when the client's own total differs
// from the authoritative total by more than the tolerance.
//
// Must be called inside the caller's transaction, like every other outbox write. A zero tolerance
// means the check is off, which is the default: an organization that has not set one has not asked
// to be told, and emitting on every rounding difference would drown the signal.
func RecordEstimatedPriceDivergence(
	ctx corectx.Context, params EstimatedPriceCheckParams,
) error {
	if !exceedsTolerance(params.Estimated, params.Actual, params.Tolerance) {
		return nil
	}

	_, err := RecordEvent(ctx, RecordEventParams{
		EventType:   models.EventSalesOrderPriceDiverged,
		AggregateId: params.SalesOrderId,
		OrgId:       params.OrgId,
		Payload: map[string]any{
			"sales_order_id":        params.SalesOrderId,
			"estimated_total_price": params.Estimated.String(),
			"grand_total":           params.Actual.String(),
			"difference":            params.Actual.Sub(*params.Estimated).Abs().String(),
			"tolerance":             params.Tolerance.String(),
		},
	})
	return err
}

// EstimatedPriceCheckParams is one order's estimate against its authoritative total.
type EstimatedPriceCheckParams struct {
	SalesOrderId string
	OrgId        string

	// Estimated is what the client sent. Nil means it sent nothing, and nothing is not a divergence.
	Estimated *decimal.Decimal

	// Actual is the grand total Sales calculated, which is what the customer owes.
	Actual decimal.Decimal

	Tolerance decimal.Decimal
}

// exceedsTolerance answers only when there is something to compare and a tolerance to compare it
// against. A non-positive tolerance switches the check off rather than making every order divergent.
func exceedsTolerance(estimated *decimal.Decimal, actual, tolerance decimal.Decimal) bool {
	if estimated == nil || !tolerance.IsPositive() {
		return false
	}
	return actual.Sub(*estimated).Abs().GreaterThan(tolerance)
}
