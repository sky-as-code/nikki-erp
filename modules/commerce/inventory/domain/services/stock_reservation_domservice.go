package services

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itExt "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/external"
)

// The reservation resource's domain service. The composable defaults serve every read; the
// writes a client could reach (create, update, delete) are withheld by the engine's action list,
// because a reservation changes only through the operations in reserve_warehouse_stock.go,
// consume_reservation.go, release_reservation.go and protect_paid_reservations.go, each of which
// runs under the warehouse guard. Those operations write through the repository directly rather
// than through this service so that the no_update fields the schema declares stay closed to
// clients while remaining writable to the engine.
func NewStockReservationDomainService(
	base composable.CrudDomainService, uom itExt.UomConversionExtService,
) *StockReservationDomainServiceImpl {
	return &StockReservationDomainServiceImpl{CrudDomainService: base, uom: uom}
}

type StockReservationDomainServiceImpl struct {
	composable.CrudDomainService

	// uom converts a caller's quantity into the variant's base unit before it meets a balance.
	uom itExt.UomConversionExtService
}

var _ composable.CrudDomainService = (*StockReservationDomainServiceImpl)(nil)
