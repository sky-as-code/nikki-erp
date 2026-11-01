package v1

import (
	"go.uber.org/dig"

	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// StockMoveEngineName is the container name of the stock_move onion, declared once for the struct tag and
// checked against dynamicresource.EngineDependencyName by a test.
const StockMoveEngineName = "dynengine_inventory_stock_move"

type stockMoveRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_inventory_stock_move"`
}

func NewStockMoveRest(params stockMoveRestParams) *StockMoveRest {
	rest := &StockMoveRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.stockMoveSvc = params.Engine.ApplicationService().(itStock.StockMoveApplicationService)
	return rest
}

// StockMoveRest serves the built-in CRUD of the stock_move resource.
type StockMoveRest struct {
	dynamicresource.CrudRestBase
	stockMoveSvc itStock.StockMoveApplicationService
}
