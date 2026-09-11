package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// StockOperationTypeEngineName is the container name of the stock_operation_type onion, declared once for the struct tag and
// checked against composable.EngineDependencyName by a test.
const StockOperationTypeEngineName = "dynengine_inventory_stock_operation_type"

type stockOperationTypeRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_stock_operation_type"`
}

func NewStockOperationTypeRest(params stockOperationTypeRestParams) *StockOperationTypeRest {
	rest := &StockOperationTypeRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.stockOperationTypeSvc = params.Engine.ApplicationService().(itStock.StockOperationTypeApplicationService)
	return rest
}

// StockOperationTypeRest serves the built-in CRUD of the stock_operation_type resource.
type StockOperationTypeRest struct {
	composable.CrudRestBase
	stockOperationTypeSvc itStock.StockOperationTypeApplicationService
}
