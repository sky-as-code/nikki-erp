package v1

import (
	"go.uber.org/dig"

	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// StockTransferEngineName is the container name of the stock_transfer onion, declared once for the struct tag and
// checked against dynamicresource.EngineDependencyName by a test.
const StockTransferEngineName = "dynengine_inventory_stock_transfer"

type stockTransferRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_inventory_stock_transfer"`
}

func NewStockTransferRest(params stockTransferRestParams) *StockTransferRest {
	rest := &StockTransferRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.stockTransferSvc = params.Engine.ApplicationService().(itStock.StockTransferApplicationService)
	return rest
}

// StockTransferRest serves the built-in CRUD of the stock_transfer resource.
type StockTransferRest struct {
	dynamicresource.CrudRestBase
	stockTransferSvc itStock.StockTransferApplicationService
}
