package stock

import (
	"time"

	"github.com/shopspring/decimal"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// The warehouse-level reservation contract (CR-INV-SALES-WH-RESERVATION). A caller commits a
// quantity of a variant at a warehouse and names no location; the location is decided when the
// goods actually leave, in ConsumeReservation. Every operation runs under the warehouse guard and
// decides expiry on the database clock read after the lock, never on a stored status.

// WarehouseReservationLine is one demand line of a reserve request.
type WarehouseReservationLine struct {
	// SourceLineId is the caller's line identifier, stored verbatim. Two lines of one request
	// may name the same variant; they are summed for the availability check and held separately.
	SourceLineId     string   `json:"source_line_id"`
	ProductVariantId model.Id `json:"product_variant_id"`

	// UomId is the unit Quantity is expressed in. Empty means the variant's base unit; anything
	// else is converted before the quantity is compared with stock.
	UomId    model.Id        `json:"uom_id"`
	Quantity decimal.Decimal `json:"quantity"`
}

type ReserveWarehouseStockRequest struct {
	OrgId       model.Id                   `json:"org_id"`
	WarehouseId model.Id                   `json:"warehouse_id"`
	Lines       []WarehouseReservationLine `json:"lines"`

	// ReservedUntil is when the hold lapses; nil means never. It must lie in the future at the
	// moment the guard is held.
	ReservedUntil *time.Time `json:"reserved_until"`

	// The demand the hold serves. SourceModule is stamped by the calling port rather than taken
	// from a client. SourceRevision distinguishes a fresh reserve for a changed demand from a
	// retry of the old one; zero means 1.
	SourceModule   string `json:"source_module"`
	SourceType     string `json:"source_type"`
	SourceId       string `json:"source_id"`
	SourceRevision int32  `json:"source_revision"`

	IdempotencyKey string `json:"idempotency_key"`
}

// ReservedLine is one reservation row as it stands after the call: freshly written, or replayed
// with its current figures.
type ReservedLine struct {
	ReservationId     model.Id        `json:"reservation_id"`
	SourceLineId      string          `json:"source_line_id"`
	ProductVariantId  model.Id        `json:"product_variant_id"`
	BaseUomId         model.Id        `json:"base_uom_id"`
	Quantity          decimal.Decimal `json:"quantity"`
	RemainingQuantity decimal.Decimal `json:"remaining_quantity"`
	Status            string          `json:"status"`
	EffectiveStatus   string          `json:"effective_status"`
	ReservedUntil     *time.Time      `json:"reserved_until"`
}

type ReserveWarehouseStockResult struct {
	Reservations []ReservedLine `json:"reservations"`

	// Replayed reports that the request was already served and these are the stored rows, with
	// their current status; a replay never re-activates a lapsed or released hold.
	Replayed bool `json:"replayed"`

	ClientErrors ft.ClientErrors `json:"client_errors,omitempty"`
}

func (this ReserveWarehouseStockResult) Refused() bool {
	return this.ClientErrors.Count() > 0
}

type CheckWarehouseAvailabilityQuery struct {
	OrgId            model.Id `json:"org_id"`
	WarehouseId      model.Id `json:"warehouse_id"`
	ProductVariantId model.Id `json:"product_variant_id"`

	// Requested is what the caller would like to reserve; Shortage answers how much of it cannot
	// be met. Zero is allowed and simply yields no shortage.
	Requested decimal.Decimal `json:"requested"`
}

// WarehouseAvailabilityReport is advisory: it promises nothing about the next reserve.
type WarehouseAvailabilityReport struct {
	EligibleOnHand    decimal.Decimal `json:"eligible_on_hand"`
	EffectiveReserved decimal.Decimal `json:"effective_reserved"`
	Available         decimal.Decimal `json:"available"`
	Shortage          decimal.Decimal `json:"shortage"`
	AsOf              time.Time       `json:"as_of"`

	ClientErrors ft.ClientErrors `json:"client_errors,omitempty"`
}

// ActualSource is where goods actually came from when a reservation was consumed: the location
// the executor took them from and, when the stock is tracked, the lot, package and owner. The
// location must belong to the reservation's warehouse.
type ActualSource struct {
	LocationId model.Id        `json:"location_id"`
	LotRef     string          `json:"lot_ref"`
	PackageRef string          `json:"package_ref"`
	OwnerRef   string          `json:"owner_ref"`
	Quantity   decimal.Decimal `json:"quantity"`
}

type ConsumeReservationRequest struct {
	OrgId         model.Id `json:"org_id"`
	ReservationId model.Id `json:"reservation_id"`

	// ExecutionId names the physical attempt this consumption records; IdempotencyKey makes a
	// resent report of the same attempt harmless. Both are required.
	ExecutionId    string `json:"execution_id"`
	IdempotencyKey string `json:"idempotency_key"`

	// ActualSources sum to the quantity consumed. Nothing is consumed without one: a consumed
	// figure that no movement backs would be a phantom issue.
	ActualSources []ActualSource `json:"actual_sources"`

	// OperationTypeId records the movement under the caller's outgoing operation type; empty
	// means the module's correction type. DestinationLocationId is where the goods went; empty
	// means the org's customer location.
	OperationTypeId       model.Id `json:"operation_type_id"`
	DestinationLocationId model.Id `json:"destination_location_id"`
	OriginReference       string   `json:"origin_reference"`
}

type ConsumeReservationResult struct {
	ReservationId     model.Id        `json:"reservation_id"`
	ConsumedNow       decimal.Decimal `json:"consumed_now"`
	ConsumedTotal     decimal.Decimal `json:"consumed_total"`
	RemainingQuantity decimal.Decimal `json:"remaining_quantity"`
	Status            string          `json:"status"`
	EffectiveStatus   string          `json:"effective_status"`

	// InventoryResultRef is the movement document that recorded the issue, for the caller to
	// keep as proof the ledger moved.
	InventoryResultRef string `json:"inventory_result_ref"`

	// Replayed reports that the execution was already recorded and nothing moved this time.
	Replayed bool `json:"replayed"`

	ClientErrors ft.ClientErrors `json:"client_errors,omitempty"`
}

func (this ConsumeReservationResult) Refused() bool {
	return this.ClientErrors.Count() > 0
}

type ReleaseReservationRequest struct {
	OrgId         model.Id `json:"org_id"`
	ReservationId model.Id `json:"reservation_id"`

	// Reason is recorded on the row; empty is allowed.
	Reason string `json:"reason"`

	// Quantity is how much to give back. Zero means the whole remainder, the ordinary release; a
	// positive value is the internal partial release, which needs IdempotencyKey so a retry does
	// not release twice. A full release is idempotent on its own: nothing is left to release.
	Quantity       decimal.Decimal `json:"quantity"`
	IdempotencyKey string          `json:"idempotency_key"`
}

type ReleaseReservationResult struct {
	ReservationId     model.Id        `json:"reservation_id"`
	ReleasedNow       decimal.Decimal `json:"released_now"`
	ReleasedTotal     decimal.Decimal `json:"released_total"`
	RemainingQuantity decimal.Decimal `json:"remaining_quantity"`
	Status            string          `json:"status"`
	EffectiveStatus   string          `json:"effective_status"`

	ClientErrors ft.ClientErrors `json:"client_errors,omitempty"`
}

func (this ReleaseReservationResult) Refused() bool {
	return this.ClientErrors.Count() > 0
}

type ProtectPaidReservationsRequest struct {
	OrgId          model.Id `json:"org_id"`
	SourceModule   string   `json:"source_module"`
	SourceType     string   `json:"source_type"`
	SourceId       string   `json:"source_id"`
	SourceRevision int32    `json:"source_revision"`

	// PaymentReference is the evidence the protection rests on, recorded in the event.
	PaymentReference string `json:"payment_reference"`
	IdempotencyKey   string `json:"idempotency_key"`
}

type ProtectPaidReservationsResult struct {
	// Protected are the reservations whose deadline was cleared by this call; AlreadyProtected
	// had none to clear. Expired had lapsed before the lock was taken and were not revived: the
	// caller decides whether to reserve afresh.
	Protected        []model.Id `json:"protected"`
	AlreadyProtected []model.Id `json:"already_protected"`
	Expired          []model.Id `json:"expired"`

	ClientErrors ft.ClientErrors `json:"client_errors,omitempty"`
}

func (this ProtectPaidReservationsResult) Refused() bool {
	return this.ClientErrors.Count() > 0
}

// AllHeld reports that every reservation of the source is now protected.
func (this ProtectPaidReservationsResult) AllHeld() bool {
	return len(this.Expired) == 0 && (len(this.Protected)+len(this.AlreadyProtected)) > 0
}

// WarehouseReservationService is the port through which the reservation operations are reached,
// by the module's own application service and by the ports other modules bind.
type WarehouseReservationService interface {
	ReserveWarehouseStock(ctx corectx.Context, request ReserveWarehouseStockRequest) (*ReserveWarehouseStockResult, error)
	CheckWarehouseAvailability(ctx corectx.Context, query CheckWarehouseAvailabilityQuery) (*WarehouseAvailabilityReport, error)
	ConsumeReservation(ctx corectx.Context, request ConsumeReservationRequest) (*ConsumeReservationResult, error)
	ReleaseReservation(ctx corectx.Context, request ReleaseReservationRequest) (*ReleaseReservationResult, error)
	ProtectPaidReservations(ctx corectx.Context, request ProtectPaidReservationsRequest) (*ProtectPaidReservationsResult, error)

	// ExpireLapsedWarehouseReservations materializes expiry on up to limit lapsed rows, for the
	// bookkeeping sweep. Availability never waits for it.
	ExpireLapsedWarehouseReservations(ctx corectx.Context, asOf time.Time, limit int) (int, error)

	// ResolveWarehouseOfLocation answers which warehouse a location belongs to; empty when none.
	ResolveWarehouseOfLocation(ctx corectx.Context, orgId, locationId model.Id) (model.Id, error)

	// ReservationsOfSource lists every reservation row of one demand revision, as it stands now.
	ReservationsOfSource(ctx corectx.Context, orgId model.Id, sourceModule, sourceType, sourceId string, revision int32) ([]ReservedLine, error)
}
