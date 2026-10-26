package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// NewStockMoveDependencyRepository is handed the composable default by the stock_move_dependency onion. The type exists so a
// bespoke query has a home; it adds nothing today.
func NewStockMoveDependencyRepository(base composable.CrudRepository) itStock.StockMoveDependencyRepository {
	return &StockMoveDependencyRepositoryImpl{CrudRepository: base}
}

type StockMoveDependencyRepositoryImpl struct {
	composable.CrudRepository
}
