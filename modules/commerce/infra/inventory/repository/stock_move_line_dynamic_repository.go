package repository

import (
	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewStockMoveLineRepository is handed the composable default by the stock_move_line onion. The type exists so a
// bespoke query has a home; it adds nothing today.
func NewStockMoveLineRepository(base composable.CrudRepository) itStock.StockMoveLineRepository {
	return &StockMoveLineRepositoryImpl{CrudRepository: base}
}

type StockMoveLineRepositoryImpl struct {
	composable.CrudRepository
}
