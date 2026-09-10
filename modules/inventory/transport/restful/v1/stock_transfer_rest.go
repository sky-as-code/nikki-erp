package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// StockTransferEngineName is the container name of the stock_transfer onion, declared once for the struct tag and
// checked against composable.EngineDependencyName by a test.
const StockTransferEngineName = "dynengine_inventory_stock_transfer"

type stockTransferRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_stock_transfer"`
}

func NewStockTransferRest(params stockTransferRestParams) *StockTransferRest {
	rest := &StockTransferRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.stockTransferSvc = params.Engine.ApplicationService().(itStock.StockTransferApplicationService)
	return rest
}

// StockTransferRest serves the built-in CRUD of the stock_transfer resource.
type StockTransferRest struct {
	composable.CrudRestBase
	stockTransferSvc itStock.StockTransferApplicationService
}
