package repository

import (
	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewStockMoveLineRepository is handed the composable default by the stock_move_line onion. The type exists so a
// bespoke query has a home; it adds nothing today.
func NewStockMoveLineRepository(base dynamicresource.CrudRepository) itStock.StockMoveLineRepository {
	return &StockMoveLineRepositoryImpl{CrudRepository: base}
}

type StockMoveLineRepositoryImpl struct {
	dynamicresource.CrudRepository
}
