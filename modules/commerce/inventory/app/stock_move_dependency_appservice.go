package app

import (
	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewStockMoveDependencyApplicationService is handed the composable default by the stock_move_dependency onion.
func NewStockMoveDependencyApplicationService(base composable.CrudApplicationService) itStock.StockMoveDependencyApplicationService {
	return &StockMoveDependencyApplicationServiceImpl{CrudApplicationService: base}
}

// StockMoveDependencyApplicationServiceImpl is the authorized CRUD of the resource.
type StockMoveDependencyApplicationServiceImpl struct {
	composable.CrudApplicationService
}
