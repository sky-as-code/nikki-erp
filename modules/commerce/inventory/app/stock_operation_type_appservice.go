package app

import (
	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewStockOperationTypeApplicationService is handed the composable default by the stock_operation_type onion.
func NewStockOperationTypeApplicationService(base composable.CrudApplicationService) itStock.StockOperationTypeApplicationService {
	return &StockOperationTypeApplicationServiceImpl{CrudApplicationService: base}
}

// StockOperationTypeApplicationServiceImpl is the authorized CRUD of the resource.
type StockOperationTypeApplicationServiceImpl struct {
	composable.CrudApplicationService
}
