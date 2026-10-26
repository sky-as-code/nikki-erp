package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// NewStockMoveRepository is handed the composable default by the stock_move onion. The type exists so a
// bespoke query has a home; it adds nothing today.
func NewStockMoveRepository(base composable.CrudRepository) itStock.StockMoveRepository {
	return &StockMoveRepositoryImpl{CrudRepository: base}
}

type StockMoveRepositoryImpl struct {
	composable.CrudRepository
}
