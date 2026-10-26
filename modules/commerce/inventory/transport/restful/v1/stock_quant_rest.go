package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// StockQuantEngineName is the container name of the stock_quant onion, declared once for the struct tag and
// checked against composable.EngineDependencyName by a test.
const StockQuantEngineName = "dynengine_inventory_stock_quant"

type stockQuantRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_stock_quant"`
}

func NewStockQuantRest(params stockQuantRestParams) *StockQuantRest {
	rest := &StockQuantRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.stockQuantSvc = params.Engine.ApplicationService().(itStock.StockQuantApplicationService)
	return rest
}

// StockQuantRest serves the built-in CRUD of the stock_quant resource.
type StockQuantRest struct {
	composable.CrudRestBase
	stockQuantSvc itStock.StockQuantApplicationService
}
