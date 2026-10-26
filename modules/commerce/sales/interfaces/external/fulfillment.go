package external

import (
	"github.com/shopspring/decimal"

	"github.com/sky-as-code/nikki-erp/common/model"
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

	// SourceLocationId is the exact location this line's goods come from, for a target whose stock
	// sits in addressable places rather than one pool — a vending slot being the case that needs it.
	// Empty means the fulfillment's single target location applies, which is every other target.
	SourceLocationId string
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

// The reservation methods below are the fulfillment seam, distinct from the four intents above.
// The difference is who names the demand: the older methods hand Inventory a sales order and let it
// pick the location from an operation type's defaults, while these name a location and a source
// reference of their own, because a kiosk fulfillment is a promise about ONE machine and stock held
// anywhere else would not satisfy it.
//
// Every one of them is keyed by that source reference rather than by whatever document Inventory
// created. Sales therefore never has to store and re-present an inventory id to release or move a
// hold, and a retry after a lost reply finds the existing reservation instead of making a second.
type FulfillmentReservationExtService interface {
	// ReserveForFulfillment holds stock for one fulfillment at one location. Idempotent by source
	// reference: called twice for the same fulfillment it returns the existing hold rather than
	// doubling it, which is what makes a confirm safe to retry.
	ReserveForFulfillment(
		ctx corectx.Context, request FulfillmentReservationRequest,
	) (*FulfillmentReservationResponse, error)

	// ReallocateReservation moves a hold to another location atomically — claimed at the new one
	// before released at the old, inside a single Inventory transaction. Never release-then-reserve:
	// a customer who is promised a different kiosk must not lose the goods to another buyer in the
	// gap, so a destination that cannot supply leaves the original hold exactly as it was.
	ReallocateReservation(
		ctx corectx.Context, request FulfillmentReservationRequest,
	) (*FulfillmentReservationResponse, error)

	// ReleaseFulfillmentReservation gives back everything held for a fulfillment. Idempotent, and
	// releasing a reservation that is already gone is success rather than a refusal: a cancel must
	// not fail because a sweep got there first.
	ReleaseFulfillmentReservation(
		ctx corectx.Context, sourceId string,
	) (*FulfillmentReservationResponse, error)

	// CheckAvailabilityByLocations answers which of these places could supply these items. It is
	// ADVISORY and takes no lock: the answer is true of the instant it was read, and only a
	// reservation actually secures stock. Callers use it to offer a customer a shortlist, never to
	// decide that a fulfillment will succeed.
	CheckAvailabilityByLocations(
		ctx corectx.Context, query AvailabilityQuery,
	) (*AvailabilityResult, error)
}

// FulfillmentReservationRequest names a demand and where it should be met. SourceId is the
// fulfillment's own id and SourceItemId each item's, so Inventory attributes a hold to the item that
// asked for it — never to the product, because one fulfillment may name the same variant twice and
// owe them to different order lines.
type FulfillmentReservationRequest struct {
	FulfillmentId string
	OrgId         string

	// LocationId is the inventory location behind the target sales point. Sales resolves it from
	// the point rather than letting Inventory guess, because the target is a commercial promise
	// about a specific machine.
	LocationId string

	// ExpiresAt is when the hold lapses if nothing claims it. Nil means it does not expire on a
	// timer, which suits goods dispensed seconds after payment.
	ExpiresAt *model.ModelDateTime

	Items []FulfillmentReservationItem
}

// FulfillmentReservationItem is one fulfillment item's demand.
type FulfillmentReservationItem struct {
	FulfillmentItemId string
	ProductVariantId  string
	UomId             string
	Quantity          decimal.Decimal

	// SourceLocationId is where this item's stock is held. Empty defers to the request's LocationId.
	// Sales groups items by this before calling Inventory, because a hold names one location.
	SourceLocationId string
}

// FulfillmentReservationResponse is what Inventory answered about a hold.
type FulfillmentReservationResponse struct {
	// Accepted says whether Inventory took the request. False is a normal outcome, not a fault.
	Accepted bool

	// InventoryReference is the reservation document Inventory created or found. Opaque to Sales,
	// stored per item so an operator can trace a hold without searching by origin.
	InventoryReference string

	// FullyReserved distinguishes the whole demand being held from part of it. A partial hold is
	// not an acceptance of the sale: a caller that requires all of it refuses on this flag rather
	// than on Accepted, which only says Inventory understood the request.
	FullyReserved bool

	// FailureReason explains a refusal in Inventory's words, so an operator can tell whether to
	// wait for a restock, offer another kiosk, or refund.
	FailureReason string
}

// AvailabilityQuery asks whether a set of places could supply a set of items.
type AvailabilityQuery struct {
	OrgId       string
	LocationIds []string
	Items       []AvailabilityItem
}

// AvailabilityItem is one product and how much of it a caller is looking for.
type AvailabilityItem struct {
	ProductVariantId string
	Quantity         decimal.Decimal
}

// AvailabilityResult reports per location, in the order Inventory answered. It carries no guarantee
// and no hold: see the advisory note on CheckAvailabilityByLocations.
type AvailabilityResult struct {
	Locations []LocationAvailability
}

// LocationAvailability is one place's answer.
type LocationAvailability struct {
	LocationId string

	// CanFulfillAll is the whole question most callers are asking, precomputed so that every caller
	// does not re-derive it from the shortages and get the empty-list edge case wrong.
	CanFulfillAll bool

	// Shortages lists only what is missing, so an empty slice and CanFulfillAll agree by
	// construction rather than by convention.
	Shortages []AvailabilityShortage
}

// AvailabilityShortage is one item this location cannot fully supply.
type AvailabilityShortage struct {
	ProductVariantId string
	Requested        decimal.Decimal
	Available        decimal.Decimal
}
