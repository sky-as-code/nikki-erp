package app

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/services"
	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

func NewStockReservationApplicationService(base dynamicresource.CrudApplicationService) itStock.StockReservationApplicationService {
	reservationSvc, ok := base.DomainService().(*services.StockReservationDomainServiceImpl)
	if !ok {
		panic(errors.New("the stock reservation onion must be built with NewStockReservationDomainService"))
	}
	return &StockReservationApplicationServiceImpl{CrudApplicationService: base, reservationSvc: reservationSvc}
}

// StockReservationApplicationServiceImpl serves the reads of the reservation resource. The
// reservation operations reachable by a client are added as custom actions on this type, each
// opening with assertRecordAction; the built-in writes are withheld at the engine.
type StockReservationApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
	reservationSvc *services.StockReservationDomainServiceImpl
}
