package external

import (
	"time"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// The warehouse-level reservation seam (CR-INV-SALES-WH-RESERVATION). Unlike the location holds
// above, a warehouse reservation names no machine slot: a kiosk sale commits a quantity at the
// warehouse behind the sales point, and the slot the goods actually leave from is reported when
// they leave. The types are Inventory's own, aliased here so the two modules describe one
// contract; a module split into its own process rebinds infra/external and keeps these names.

type (
	WarehouseReservationLine       = itStock.WarehouseReservationLine
	ReserveWarehouseStockRequest   = itStock.ReserveWarehouseStockRequest
	ReserveWarehouseStockResult    = itStock.ReserveWarehouseStockResult
	ReservedLine                   = itStock.ReservedLine
	ReleaseReservationRequest      = itStock.ReleaseReservationRequest
	ReleaseReservationResult       = itStock.ReleaseReservationResult
	ProtectPaidReservationsRequest = itStock.ProtectPaidReservationsRequest
	ProtectPaidReservationsResult  = itStock.ProtectPaidReservationsResult
)

// How Sales names its demands to Inventory. Plain strings on both sides: Inventory stores them to
// attribute a hold and never resolves them against a Sales table.
const (
	SourceModuleSales          = "sales"
	SourceTypeOrderFulfillment = "sales_order_fulfillment"
)

// WarehouseReservationExtService is Sales' port onto warehouse-level reservations.
type WarehouseReservationExtService interface {
	// ReserveWarehouseStock holds every line at the warehouse or nothing; a retry under the same
	// key and source is answered from the rows already written.
	ReserveWarehouseStock(ctx corectx.Context, request ReserveWarehouseStockRequest) (*ReserveWarehouseStockResult, error)

	// ReleaseReservation gives back a hold's remainder; a hold already gone is success.
	ReleaseReservation(ctx corectx.Context, request ReleaseReservationRequest) (*ReleaseReservationResult, error)

	// ProtectPaidReservations clears the deadline of a paid demand's holds; a hold that lapsed
	// first is reported back, never revived.
	ProtectPaidReservations(ctx corectx.Context, request ProtectPaidReservationsRequest) (*ProtectPaidReservationsResult, error)

	// ReservationsOfSource lists a demand revision's holds as they stand now.
	ReservationsOfSource(ctx corectx.Context, orgId model.Id, sourceModule, sourceType, sourceId string, revision int32) ([]ReservedLine, error)

	// ResolveWarehouseOfLocation answers the warehouse behind a sales point's location; empty
	// when the location belongs to none.
	ResolveWarehouseOfLocation(ctx corectx.Context, orgId, locationId model.Id) (model.Id, error)
}

// deadlineOrNil turns an optional model timestamp into the port's own deadline shape.
func DeadlineOrNil(value *model.ModelDateTime) *time.Time {
	if value == nil {
		return nil
	}
	goTime := value.GoTime().UTC()
	return &goTime
}
