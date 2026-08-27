package external

import (
	"github.com/shopspring/decimal"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// FulfillmentExtService is Sales' port onto Inventory's goods movement.
//
// **Sales sends intent and never touches stock** (BR 44, BR 3.2). Every method here says what
// commercially happened - these goods were sold, these came back - and Inventory decides
// availability, reservation, warehouse, location and the movements that follow. Acceptance criterion
// BR 94.6 tests exactly this separation, and the port is shaped to make violating it awkward: there
// is no method for adjusting a quantity, moving between locations, or naming a warehouse.
//
// # This port is DECLARED but NOT YET BOUND
//
// Inventory publishes readers only. Its mutation surface -
// `services.StockTransferDomainServiceImpl` with Create, Confirm, Reserve, Unreserve, Validate and
// CreateReturn - is a **concrete struct with no interface**, reachable only by type-asserting the
// resource service inside one of Inventory's own action callbacks. Sales cannot reach it, and this
// module must not reach across and construct one.
//
// Two routes out, and the choice is Inventory's to make rather than Sales':
//
//  1. Extract a StockTransferDomainService interface into `inventory/interfaces/stock/` and register
//     it in deps. Clean, matches how Essential and Contacts publish their capabilities, and is the
//     same fix SALES-048 needs for the product port.
//  2. Drive the engine actions by schema name, which is what paymentinvoice does internally - but it
//     leaves Sales shaping params and carrying permission concerns that belong upstream.
//
// Until one lands, `sales_fulfillment_requests` rows are written and stay `pending`. That is the
// honest state: the sale really has asked for goods and Inventory really has not answered. See the
// SALES-029 note in 02-progress.md.

// FulfillmentExtService is the capability Sales needs from Inventory.
type FulfillmentExtService interface {
	// RequestReservation asks Inventory to hold stock for a confirmed sale, without moving it.
	//
	// Separate from the goods issue because they happen at different moments and can fail
	// differently: a reservation that cannot be met means the sale should not have confirmed, while
	// an issue that fails after a successful reservation is BR 7.3's case and needs a compensating
	// movement rather than a refusal.
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

// FulfillmentRequest is one commercial intent, expressed in Sales' own terms.
//
// Note what it does NOT carry: no warehouse, no location, no movement type, no stock quant. Those
// are Inventory's decisions, and a field for one here would be an invitation to make it in Sales.
type FulfillmentRequest struct {
	// SalesOrderId and SalesFulfillmentRequestId let Inventory trace what it was asked and by whom.
	// Opaque to Inventory: it echoes them back and never resolves them against Sales.
	SalesOrderId              string
	SalesFulfillmentRequestId string

	// IdempotencyKey makes a retry safe. A fulfilment request that arrives twice must not issue the
	// goods twice, and the key is what lets Inventory recognise the second call.
	IdempotencyKey string

	Lines []FulfillmentLine
}

// FulfillmentLine is one product and how much of it.
type FulfillmentLine struct {
	// SalesOrderLineId is echoed back so Sales can attribute what Inventory reports to the right
	// line without matching on product and quantity, which are not unique within an order.
	SalesOrderLineId string

	ProductVariantId string
	UomId            string
	Quantity         decimal.Decimal
}

// FulfillmentResponse is what Inventory answered.
type FulfillmentResponse struct {
	// Accepted says whether Inventory took the request. False is a normal outcome, not a fault:
	// stock runs out.
	Accepted bool

	// InventoryReference is whatever Inventory created - a stock transfer id. Sales stores it and
	// uses it to ask about or release the request later; it never interprets it.
	InventoryReference string

	// FailureReason explains a refusal in Inventory's words. A rejected fulfilment is the case an
	// operator must act on, and "rejected" alone tells them nothing about whether to wait, re-route
	// or refund.
	FailureReason string

	// Completed distinguishes goods actually moved from stock merely held. BR 7.3's failure lives
	// between the two, so a single "done" flag could not express it.
	Completed bool
}
