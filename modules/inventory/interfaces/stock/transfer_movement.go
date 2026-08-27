package stock

import (
	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// The goods-movement port other modules bind to (SALES-049).
//
// Until this landed, Stock published readers only. Its movement operations — create, confirm,
// reserve, unreserve, validate, return — lived on `services.StockTransferDomainServiceImpl`, a
// concrete struct reachable only by type-asserting the resource service inside one of Stock's own
// engine actions. A selling module could therefore record that it owed a customer goods but had no
// way to ask for them, so its requests sat unanswered.
//
// # Why this is a NARROWED interface rather than the whole struct
//
// The struct also embeds `drif.DynamicResourceService`, which is full CRUD over the transfer
// resource. Publishing it would hand every consumer the power to write a transfer's fields
// directly — to set a status to `done` without moving anything, or to edit a validated document.
// Those are exactly the operations the lifecycle methods exist to police, so a port that offered
// both would make the rules optional. This interface lists the six operations and nothing else.
//
// # Consumers still do not choose warehouses
//
// Every method here takes a transfer id, and the transfer already carries its operation type,
// locations and policies — chosen when it was created, by Stock's own rules. A consumer sequences
// the document's life; it does not decide where goods sit. Nothing here accepts a location.
type StockTransferMovementService interface {
	// Create raises a draft transfer. The params are the transfer's own fields: an operation type
	// is required, and Stock stamps the number, status and policies from it.
	//
	// Writes the HEADER only. A consumer outside Inventory cannot follow it with moves of its own —
	// the move engine is not published, and reaching for it would put the shape of Stock's tables
	// in another module. Use CreateWithMoves instead.
	Create(
		ctx corectx.Context, params dmodel.DynamicFields,
	) (*dyn.OpResult[dmodel.DynamicFields], error)

	// CreateWithMoves raises a draft transfer together with the lines it moves, in ONE transaction.
	//
	// This is the entry point for another module, and it exists because the two halves must not be
	// separable from outside. A consumer that created a header and then failed to add its moves
	// would leave an empty transfer that validates successfully and reports goods moved that were
	// never named — a silent, wrong success. Committing both together makes that unreachable.
	//
	// It also keeps the move's shape inside Inventory: the caller says what and how much, and the
	// sequence numbers, base quantities and location defaults are derived here.
	CreateWithMoves(
		ctx corectx.Context, params dmodel.DynamicFields, moves []TransferMoveRequest,
	) (*dyn.OpResult[dmodel.DynamicFields], error)

	// Confirm moves a draft into the flow, which is what makes its moves eligible to reserve.
	Confirm(ctx corectx.Context, transferId string) (*dyn.OpResult[dyn.MutateResultData], error)

	// Reserve claims stock for the transfer's moves without moving it.
	//
	// A partial claim is a NORMAL outcome, not an error: stock runs out, and the transfer records
	// how much it managed to hold. A caller that needs all-or-nothing must compare the reserved
	// quantity against what it asked for rather than treating a nil error as success.
	Reserve(ctx corectx.Context, transferId string) (*dyn.OpResult[dyn.MutateResultData], error)

	// Unreserve gives back what Reserve claimed, for a transfer that is no longer wanted.
	Unreserve(ctx corectx.Context, transferId string) (*dyn.OpResult[dyn.MutateResultData], error)

	// Validate moves the goods. This is the irreversible one — balances change and the movements
	// are recorded as fact, so an edit cannot undo it and a correction must be its own document.
	//
	// The idempotency key is what makes a retry safe: validating twice would move the goods twice,
	// and a caller that timed out cannot tell whether the first call landed.
	//
	// createBackorder decides what happens to the shortfall when only part of the demand was
	// reserved. Nil defers to the transfer's own backorder policy, which is the right default —
	// the policy was chosen by whoever configured the operation type.
	Validate(
		ctx corectx.Context, transferId string, idempotencyKey string, createBackorder *bool,
	) (*dyn.OpResult[dyn.MutateResultData], error)

	// CreateReturn raises a draft reverse transfer for goods that came back.
	//
	// It returns a DRAFT, deliberately: raising a return states an intent, and validating it states
	// that the goods are physically back. The original transfer is never touched, so history reads
	// as original movement then reverse movement.
	CreateReturn(
		ctx corectx.Context, transferId string, request TransferReturnRequest,
	) (*dyn.OpResult[dyn.MutateResultData], error)
}

// TransferMoveRequest is one line of a transfer a consumer is raising.
//
// Deliberately three fields. There is no location, no status, no reserved quantity and no
// valuation: those are Stock's to derive, and a field for one here would let a consumer assert
// something it cannot know. Base quantity is absent for the same reason — converting to the
// variant's base unit needs a conversion factor the consumer does not hold, so Stock derives it.
//
// There is no per-line origin reference because inventory_stock_move has no column for one. The
// transfer's own origin_reference ties the whole document back to what caused it, which is the
// granularity Stock actually records.
type TransferMoveRequest struct {
	ProductVariantId string

	// UomId is the unit Quantity is expressed in. Empty means the variant's own base unit.
	UomId string

	Quantity decimal.Decimal
}

// TransferReturnRequest names how much of each move to send back.
//
// An empty Lines means everything still returnable, which is the common case.
type TransferReturnRequest struct {
	Lines []TransferReturnLine
}

// TransferReturnLine is one move's requested return quantity.
type TransferReturnLine struct {
	MoveId   string
	Quantity decimal.Decimal
}
