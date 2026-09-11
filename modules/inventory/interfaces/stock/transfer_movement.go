package stock

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"time"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
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
		ctx corectx.Context, params dmodel.DynamicFields, options ...composable.CreateOptions,
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

	// ReserveForSource holds stock for a demand that lives in another module, remembering whose it
	// is so the caller can release or move it later without having stored the transfer id.
	//
	// Distinct from Create+Confirm+Reserve, which a consumer can already sequence itself: those
	// three leave no record of WHY the hold exists, so a caller that lost the id could only find it
	// by guessing from locations and quantities. Idempotent by the source pair.
	ReserveForSource(
		ctx corectx.Context, request SourceReservationRequest,
	) (*SourceReservationResult, error)

	// ReleaseReservationBySource gives back every hold taken for a demand. Releasing a demand that
	// holds nothing is success: the caller asked for it to hold nothing, and it does not.
	ReleaseReservationBySource(
		ctx corectx.Context, sourceType string, sourceId string,
	) (*dyn.OpResult[dyn.MutateResultData], error)

	// ReallocateReservation moves a demand's hold from one location to another, ATOMICALLY: the
	// stock is claimed at the destination before it is given back at the origin, inside a single
	// transaction. If the destination cannot supply it, nothing changes and the original hold is
	// untouched — the customer keeps the guarantee they already had.
	//
	// Never release-then-reserve. Doing it in that order would leave a window in which the customer
	// holds nothing, and a concurrent sale could take the stock they were promised.
	ReallocateReservation(
		ctx corectx.Context, request ReservationReallocationRequest,
	) (*SourceReservationResult, error)

	// ExpireLapsedReservations releases every hold whose expiry has passed, up to limit of them,
	// and reports how many it released. It only ever releases stock: expiring a hold says the goods
	// are sellable again, never that the demand is cancelled or that anybody should be refunded.
	ExpireLapsedReservations(
		ctx corectx.Context, asOf time.Time, limit int,
	) (int, error)

	// ApplyFulfillmentResult records what an executor physically managed to hand over against a
	// demand's reservation: the quantities that left are consumed, and the hold on the rest is
	// given back.
	//
	// The caller reports facts — this much came out, this much did not — and never the consequence.
	// What becomes of goods a machine failed to dispense (still sellable, jammed, needing a human)
	// is Inventory's judgement, because it is a statement about stock.
	//
	// Idempotent by event id: a broker that delivers the same event twice must not take the goods
	// off the shelf twice.
	ApplyFulfillmentResult(
		ctx corectx.Context, request FulfillmentResultRequest,
	) (*FulfillmentResultResponse, error)
}

// FulfillmentResultRequest reports one executor attempt against one demand.
type FulfillmentResultRequest struct {
	// EventId identifies the report itself, and is the idempotency key. Required.
	EventId string

	// SourceType and SourceId name the demand whose reservation this is a result for.
	SourceType string
	SourceId   string

	OrgId string

	// ExecutorLocationId is where the goods were handed over from. Recorded for reconciliation; the
	// stock actually consumed is whatever the reservation held, which is the authoritative answer.
	ExecutorLocationId string

	// Both lists are required to be meaningful together: a result saying only that something failed
	// leaves Inventory unable to tell a total failure from a partial one, and the difference decides
	// how much stock stays held.
	SuccessfulItems []FulfillmentResultItem
	FailedItems     []FulfillmentResultItem
}

// FulfillmentResultItem is what happened to one line of the demand.
type FulfillmentResultItem struct {
	// SourceItemId names the caller's own line. Quantities are attributed by it rather than by
	// product, because one demand may name the same variant on two lines.
	SourceItemId string

	ProductVariantId string
	Quantity         decimal.Decimal

	// FailureCode and FailureMessage carry the executor's diagnosis on a failed line. Recorded, not
	// interpreted: Inventory does not branch on a machine's error strings.
	FailureCode    string
	FailureMessage string
}

// FulfillmentResultResponse is Inventory's acknowledgement.
type FulfillmentResultResponse struct {
	// InventoryResultRef is what the caller quotes to prove Inventory processed the physical result
	// before it acts on it commercially.
	InventoryResultRef string

	// AlreadyApplied says the event had been processed before, so the caller can tell a replay from
	// a first delivery without treating either as a failure.
	AlreadyApplied bool

	ClientErrors ft.ClientErrors
}

// SourceReservationRequest asks for stock to be held at one location for one demand.
type SourceReservationRequest struct {
	// SourceType and SourceId name the demand, as '{module}_{concept}' plus its id. Together they
	// are the key the hold is found by afterwards, so both are required.
	SourceType string
	SourceId   string

	// OrgId is the organization the demand and the stock both belong to.
	OrgId string

	// LocationId is where the stock is held. Unlike the other movement operations this one accepts
	// a location, because choosing WHICH kiosk holds the goods is the caller's commercial decision
	// and Inventory has no basis to pick one.
	LocationId string

	// OperationTypeId is the document type the hold is raised under, which supplies the policies
	// and the number series. The caller reads it from its own configuration.
	OperationTypeId string

	// OriginReference is free text naming the upstream document, for a human reading the transfer.
	OriginReference string

	// ExpiresAt is when the hold lapses, in UTC. Nil means it does not expire on a timer.
	ExpiresAt *model.ModelDateTime

	Items []SourceReservationItem
}

// SourceReservationItem is one line of a demand.
type SourceReservationItem struct {
	// SourceItemId names the caller's own line, so a partial outcome is attributable.
	SourceItemId string

	ProductVariantId string
	UomId            string
	Quantity         decimal.Decimal
}

// ReservationReallocationRequest moves an existing hold to a different location.
type ReservationReallocationRequest struct {
	SourceType string
	SourceId   string
	OrgId      string

	// ToLocationId is where the stock should be held instead. The origin is not named: it is
	// whatever the demand currently holds, and asking the caller to repeat it would let the two
	// disagree.
	ToLocationId string

	OperationTypeId string
	OriginReference string
	ExpiresAt       *model.ModelDateTime

	// Items is what must be held at the new location — typically the quantity still outstanding,
	// not the whole original demand, since what was already handed over stays handed over.
	Items []SourceReservationItem
}

// SourceReservationResult is what Inventory did with a hold request.
type SourceReservationResult struct {
	// InventoryReference is the transfer that holds the stock. The caller stores it for reading
	// back and never interprets it.
	InventoryReference string

	// FullyReserved says whether the whole requested quantity is held. A partial hold is a normal
	// outcome, so a caller that promised a customer a specific location must check this rather than
	// treating the absence of an error as success.
	FullyReserved bool

	// ClientErrors carries a refusal the caller can act on, with no Go error, so the REST layer
	// answers 400 rather than 500.
	ClientErrors ft.ClientErrors
}

// TransferMoveRequest is one line of a transfer a consumer is raising.
//
// Deliberately three fields: location, status, reserved quantity and valuation are Stock's to
// derive, and a consumer cannot know them. Base quantity is absent for the same reason —
// converting to the variant's base unit needs a conversion factor the consumer does not hold.
type TransferMoveRequest struct {
	ProductVariantId string

	// UomId is the unit Quantity is expressed in. Empty means the variant's own base unit.
	UomId string

	Quantity decimal.Decimal

	// SourceItemId names the line of the caller's own demand this move serves, so a partial
	// outcome can be attributed line by line rather than guessed at from product and quantity,
	// which are not unique within a transfer. Opaque to Stock: it is echoed back, never resolved.
	// Empty when the caller tracks nothing finer than the whole document.
	SourceItemId string
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
