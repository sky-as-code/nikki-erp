package app

import (
	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewStockMoveLineApplicationService is handed the composable default by the stock_move_line onion.
func NewStockMoveLineApplicationService(base composable.CrudApplicationService) itStock.StockMoveLineApplicationService {
	return &StockMoveLineApplicationServiceImpl{CrudApplicationService: base}
}

// StockMoveLineApplicationServiceImpl is the authorized CRUD of the resource.
type StockMoveLineApplicationServiceImpl struct {
	composable.CrudApplicationService
}
