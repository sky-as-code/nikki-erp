package v1

import (
	"go.uber.org/dig"

	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// StockScrapEngineName is the container name of the stock_scrap onion, declared once for the struct tag and
// checked against dynamicresource.EngineDependencyName by a test.
const StockScrapEngineName = "dynengine_inventory_stock_scrap"

type stockScrapRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_inventory_stock_scrap"`
}

func NewStockScrapRest(params stockScrapRestParams) *StockScrapRest {
	rest := &StockScrapRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.stockScrapSvc = params.Engine.ApplicationService().(itStock.StockScrapApplicationService)
	return rest
}

// StockScrapRest serves the built-in CRUD of the stock_scrap resource.
type StockScrapRest struct {
	dynamicresource.CrudRestBase
	stockScrapSvc itStock.StockScrapApplicationService
}
