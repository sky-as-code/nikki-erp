package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// StockScrapEngineName is the container name of the stock_scrap onion, declared once for the struct tag and
// checked against composable.EngineDependencyName by a test.
const StockScrapEngineName = "dynengine_inventory_stock_scrap"

type stockScrapRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_stock_scrap"`
}

func NewStockScrapRest(params stockScrapRestParams) *StockScrapRest {
	rest := &StockScrapRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.stockScrapSvc = params.Engine.ApplicationService().(itStock.StockScrapApplicationService)
	return rest
}

// StockScrapRest serves the built-in CRUD of the stock_scrap resource.
type StockScrapRest struct {
	composable.CrudRestBase
	stockScrapSvc itStock.StockScrapApplicationService
}
