package repository

import (
	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewStockProductConfigRepository is handed the composable default by the stock_product_config onion. The type exists so a
// bespoke query has a home; it adds nothing today.
func NewStockProductConfigRepository(base composable.CrudRepository) itStock.StockProductConfigRepository {
	return &StockProductConfigRepositoryImpl{CrudRepository: base}
}

type StockProductConfigRepositoryImpl struct {
	composable.CrudRepository
}
