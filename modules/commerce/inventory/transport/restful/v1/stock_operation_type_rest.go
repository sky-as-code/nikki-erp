package v1

import (
	"go.uber.org/dig"

	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// StockOperationTypeEngineName is the container name of the stock_operation_type onion, declared once for the struct tag and
// checked against dynamicresource.EngineDependencyName by a test.
const StockOperationTypeEngineName = "dynengine_inventory_stock_operation_type"

type stockOperationTypeRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_inventory_stock_operation_type"`
}

func NewStockOperationTypeRest(params stockOperationTypeRestParams) *StockOperationTypeRest {
	rest := &StockOperationTypeRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.stockOperationTypeSvc = params.Engine.ApplicationService().(itStock.StockOperationTypeApplicationService)
	return rest
}

// StockOperationTypeRest serves the built-in CRUD of the stock_operation_type resource.
type StockOperationTypeRest struct {
	dynamicresource.CrudRestBase
	stockOperationTypeSvc itStock.StockOperationTypeApplicationService
}
