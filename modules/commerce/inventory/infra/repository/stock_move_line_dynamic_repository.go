package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// NewStockMoveLineRepository is handed the composable default by the stock_move_line onion. The type exists so a
// bespoke query has a home; it adds nothing today.
func NewStockMoveLineRepository(base composable.CrudRepository) itStock.StockMoveLineRepository {
	return &StockMoveLineRepositoryImpl{CrudRepository: base}
}

type StockMoveLineRepositoryImpl struct {
	composable.CrudRepository
}
