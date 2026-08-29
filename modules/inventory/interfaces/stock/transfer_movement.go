package stock

import (
	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// The goods-movement port other modules bind to.
//
// Narrowed deliberately rather than publishing the implementing struct, which also embeds
// drif.DynamicResourceService: full CRUD would let a consumer set a status to done without moving
// anything, or edit a validated document, making the lifecycle rules optional.
//
// Every method takes a transfer id, and the transfer already carries its operation type, locations
// and policies. A consumer sequences the document's life but does not decide where goods sit, so
// nothing here accepts a location.
type StockTransferMovementService interface {
	// Create raises a draft transfer. An operation type is required; Stock stamps the number,
	// status and policies from it. Writes the HEADER only — the move engine is not published, so a
	// consumer outside Inventory must use CreateWithMoves instead.
	Create(
		ctx corectx.Context, params dmodel.DynamicFields,
	) (*dyn.OpResult[dmodel.DynamicFields], error)

	// CreateWithMoves raises a draft transfer together with the lines it moves, in ONE transaction.
	// The halves must not be separable: a header created without its moves would validate
	// successfully and report goods moved that were never named. The caller says what and how much;
	// sequence numbers, base quantities and location defaults are derived here.
	CreateWithMoves(
		ctx corectx.Context, params dmodel.DynamicFields, moves []TransferMoveRequest,
	) (*dyn.OpResult[dmodel.DynamicFields], error)

	// Confirm moves a draft into the flow, which is what makes its moves eligible to reserve.
	Confirm(ctx corectx.Context, transferId string) (*dyn.OpResult[dyn.MutateResultData], error)

	// Reserve claims stock for the transfer's moves without moving it. A partial claim is a NORMAL
	// outcome, not an error: a caller needing all-or-nothing must compare the reserved quantity
	// against what it asked for rather than treating a nil error as success.
	Reserve(ctx corectx.Context, transferId string) (*dyn.OpResult[dyn.MutateResultData], error)

	// Unreserve gives back what Reserve claimed.
	Unreserve(ctx corectx.Context, transferId string) (*dyn.OpResult[dyn.MutateResultData], error)

	// Validate moves the goods, irreversibly: balances change and the movements are recorded as
	// fact, so a correction must be its own document. The idempotency key makes a retry safe —
	// validating twice would move the goods twice, and a timed-out caller cannot tell whether the
	// first call landed. createBackorder decides the fate of the shortfall when only part of the
	// demand was reserved; nil defers to the transfer's own backorder policy.
	Validate(
		ctx corectx.Context, transferId string, idempotencyKey string, createBackorder *bool,
	) (*dyn.OpResult[dyn.MutateResultData], error)

	// CreateReturn raises a DRAFT reverse transfer: raising it states an intent, validating it
	// states the goods are physically back. The original transfer is never touched, so history
	// reads as original movement then reverse movement.
	CreateReturn(
		ctx corectx.Context, transferId string, request TransferReturnRequest,
	) (*dyn.OpResult[dyn.MutateResultData], error)
}

// TransferMoveRequest is one line of a transfer a consumer is raising.
//
// Deliberately three fields: location, status, reserved quantity and valuation are Stock's to
// derive, and a consumer cannot know them. Base quantity is absent for the same reason —
// converting to the variant's base unit needs a conversion factor the consumer does not hold.
//
// No per-line origin reference: inventory_stock_move has no column for one, and the transfer's
// own origin_reference is the granularity Stock records.
type TransferMoveRequest struct {
	ProductVariantId string

	// UomId is the unit Quantity is expressed in. Empty means the variant's own base unit.
	UomId string

	Quantity decimal.Decimal
}

// TransferReturnRequest names how much of each move to send back. Empty Lines means everything
// still returnable.
type TransferReturnRequest struct {
	Lines []TransferReturnLine
}

// TransferReturnLine is one move's requested return quantity.
type TransferReturnLine struct {
	MoveId   string
	Quantity decimal.Decimal
}
