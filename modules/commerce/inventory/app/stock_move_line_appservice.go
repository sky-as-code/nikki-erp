package app

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// NewStockMoveLineApplicationService is handed the composable default by the stock_move_line onion.
func NewStockMoveLineApplicationService(base composable.CrudApplicationService) itStock.StockMoveLineApplicationService {
	return &StockMoveLineApplicationServiceImpl{CrudApplicationService: base}
}

// StockMoveLineApplicationServiceImpl is the authorized CRUD of the resource.
type StockMoveLineApplicationServiceImpl struct {
	composable.CrudApplicationService
}
