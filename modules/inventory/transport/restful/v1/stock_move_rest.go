package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// StockMoveEngineName is the container name of the stock_move onion, declared once for the struct tag and
// checked against composable.EngineDependencyName by a test.
const StockMoveEngineName = "dynengine_inventory_stock_move"

type stockMoveRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_stock_move"`
}

func NewStockMoveRest(params stockMoveRestParams) *StockMoveRest {
	rest := &StockMoveRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.stockMoveSvc = params.Engine.ApplicationService().(itStock.StockMoveApplicationService)
	return rest
}

// StockMoveRest serves the built-in CRUD of the stock_move resource.
type StockMoveRest struct {
	composable.CrudRestBase
	stockMoveSvc itStock.StockMoveApplicationService
}
