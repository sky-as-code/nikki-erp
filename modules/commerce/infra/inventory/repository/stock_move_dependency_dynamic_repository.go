package repository

import (
	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewStockMoveDependencyRepository is handed the composable default by the stock_move_dependency onion. The type exists so a
// bespoke query has a home; it adds nothing today.
func NewStockMoveDependencyRepository(base dynamicresource.CrudRepository) itStock.StockMoveDependencyRepository {
	return &StockMoveDependencyRepositoryImpl{CrudRepository: base}
}

type StockMoveDependencyRepositoryImpl struct {
	dynamicresource.CrudRepository
}
