package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// StockProductConfigEngineName is the container name of the stock_product_config onion, declared once for the struct tag and
// checked against composable.EngineDependencyName by a test.
const StockProductConfigEngineName = "dynengine_inventory_stock_product_config"

type stockProductConfigRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_stock_product_config"`
}

func NewStockProductConfigRest(params stockProductConfigRestParams) *StockProductConfigRest {
	rest := &StockProductConfigRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.stockProductConfigSvc = params.Engine.ApplicationService().(itStock.StockProductConfigApplicationService)
	return rest
}

// StockProductConfigRest serves the built-in CRUD of the stock_product_config resource.
type StockProductConfigRest struct {
	composable.CrudRestBase
	stockProductConfigSvc itStock.StockProductConfigApplicationService
}
