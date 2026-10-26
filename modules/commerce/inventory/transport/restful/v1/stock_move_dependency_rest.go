package v1

import (
	"go.uber.org/dig"

	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// StockMoveDependencyEngineName is the container name of the stock_move_dependency onion, declared once for the struct tag and
// checked against composable.EngineDependencyName by a test.
const StockMoveDependencyEngineName = "dynengine_inventory_stock_move_dependency"

type stockMoveDependencyRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_stock_move_dependency"`
}

func NewStockMoveDependencyRest(params stockMoveDependencyRestParams) *StockMoveDependencyRest {
	rest := &StockMoveDependencyRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.stockMoveDependencySvc = params.Engine.ApplicationService().(itStock.StockMoveDependencyApplicationService)
	return rest
}

// StockMoveDependencyRest serves the built-in CRUD of the stock_move_dependency resource.
type StockMoveDependencyRest struct {
	composable.CrudRestBase
	stockMoveDependencySvc itStock.StockMoveDependencyApplicationService
}
