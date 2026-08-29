package external

import (
	"github.com/shopspring/decimal"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// FulfillmentExtService is Sales' port onto inventory's goods movement: Sales sends intent and never
// touches stock. Every method says what commercially happened, and inventory decides availability,
// reservation, warehouse, location and the movements that follow — so there is deliberately no
// method for adjusting a quantity, moving between locations, or naming a warehouse.
type FulfillmentExtService interface {
	// RequestReservation asks inventory to hold stock for a confirmed sale without moving it.
	// Separate from the goods issue because they fail differently: an unmeetable reservation means
	// the sale should not have confirmed, while an issue failing after a successful reservation
	// needs a compensating movement rather than a refusal.
	RequestReservation(
		ctx corectx.Context, request FulfillmentRequest,
	) (*FulfillmentResponse, error)

	// RequestGoodsIssue asks Inventory to move the goods out.
	RequestGoodsIssue(
		ctx corectx.Context, request FulfillmentRequest,
	) (*FulfillmentResponse, error)

	// RequestReturnReceipt asks Inventory to take returned goods back in.
	RequestReturnReceipt(
		ctx corectx.Context, request FulfillmentRequest,
	) (*FulfillmentResponse, error)

	// ReleaseReservation gives back a hold a cancelled sale no longer needs.
	ReleaseReservation(
		ctx corectx.Context, inventoryReference string,
	) (*FulfillmentResponse, error)
}

// FulfillmentRequest is one commercial intent in Sales' own terms. It carries no warehouse,
// location, movement type or stock quant: those are inventory's decisions.
type FulfillmentRequest struct {
	// Opaque to inventory: it echoes them back and never resolves them against Sales.
	SalesOrderId              string
	SalesFulfillmentRequestId string

	// IdempotencyKey is what lets inventory recognise a second call, so a request that arrives
	// twice does not issue the goods twice.
	IdempotencyKey string

	Lines []FulfillmentLine
}

// FulfillmentLine is one product and how much of it.
type FulfillmentLine struct {
	// SalesOrderLineId is echoed back so Sales attributes what inventory reports to the right line
	// without matching on product and quantity, which are not unique within an order.
	SalesOrderLineId string

	ProductVariantId string
	UomId            string
	Quantity         decimal.Decimal
}

// FulfillmentResponse is what Inventory answered.
type FulfillmentResponse struct {
	// Accepted says whether inventory took the request. False is a normal outcome, not a fault.
	Accepted bool

	// InventoryReference is whatever inventory created — a stock transfer id. Sales stores it to ask
	// about or release the request later, and never interprets it.
	InventoryReference string

	// FailureReason explains a refusal in inventory's words: "rejected" alone would not tell an
	// operator whether to wait, re-route or refund.
	FailureReason string

	// Completed distinguishes goods actually moved from stock merely held; a failure can land
	// between the two, so a single "done" flag could not express it.
	Completed bool
}
