package v1

import (
	"go.uber.org/dig"

	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// StockReservationEngineName is the container name of the stock_reservation onion, declared once
// for the struct tag and checked against dynamicresource.EngineDependencyName by a test.
const StockReservationEngineName = "dynengine_inventory_stock_reservation"

type stockReservationRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_inventory_stock_reservation"`
}

func NewStockReservationRest(params stockReservationRestParams) *StockReservationRest {
	rest := &StockReservationRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.stockReservationSvc = params.Engine.ApplicationService().(itStock.StockReservationApplicationService)
	return rest
}

// StockReservationRest serves the reads of the stock_reservation resource; the reservation
// operations are custom routes in custom_actions_rest.go.
type StockReservationRest struct {
	dynamicresource.CrudRestBase
	stockReservationSvc itStock.StockReservationApplicationService
}
