package app

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// NewStockMoveApplicationService is handed the composable default by the stock_move onion.
func NewStockMoveApplicationService(base composable.CrudApplicationService) itStock.StockMoveApplicationService {
	return &StockMoveApplicationServiceImpl{CrudApplicationService: base}
}

// StockMoveApplicationServiceImpl is the authorized CRUD of the resource.
type StockMoveApplicationServiceImpl struct {
	composable.CrudApplicationService
}
