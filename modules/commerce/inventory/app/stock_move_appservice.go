package app

import (
	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewStockMoveApplicationService is handed the composable default by the stock_move onion.
func NewStockMoveApplicationService(base dynamicresource.CrudApplicationService) itStock.StockMoveApplicationService {
	return &StockMoveApplicationServiceImpl{CrudApplicationService: base}
}

// StockMoveApplicationServiceImpl is the authorized CRUD of the resource.
type StockMoveApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}
