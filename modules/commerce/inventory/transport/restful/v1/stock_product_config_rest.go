package v1

import (
	"go.uber.org/dig"

	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// StockProductConfigEngineName is the container name of the stock_product_config onion, declared once for the struct tag and
// checked against dynamicresource.EngineDependencyName by a test.
const StockProductConfigEngineName = "dynengine_inventory_stock_product_config"

type stockProductConfigRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_inventory_stock_product_config"`
}

func NewStockProductConfigRest(params stockProductConfigRestParams) *StockProductConfigRest {
	rest := &StockProductConfigRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.stockProductConfigSvc = params.Engine.ApplicationService().(itStock.StockProductConfigApplicationService)
	return rest
}

// StockProductConfigRest serves the built-in CRUD of the stock_product_config resource.
type StockProductConfigRest struct {
	dynamicresource.CrudRestBase
	stockProductConfigSvc itStock.StockProductConfigApplicationService
}
