package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// StockMoveLineEngineName is the container name of the stock_move_line onion, declared once for the struct tag and
// checked against composable.EngineDependencyName by a test.
const StockMoveLineEngineName = "dynengine_inventory_stock_move_line"

type stockMoveLineRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_stock_move_line"`
}

func NewStockMoveLineRest(params stockMoveLineRestParams) *StockMoveLineRest {
	rest := &StockMoveLineRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.stockMoveLineSvc = params.Engine.ApplicationService().(itStock.StockMoveLineApplicationService)
	return rest
}

// StockMoveLineRest serves the built-in CRUD of the stock_move_line resource.
type StockMoveLineRest struct {
	composable.CrudRestBase
	stockMoveLineSvc itStock.StockMoveLineApplicationService
}
