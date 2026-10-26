package repository

import (
	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewStockQuantRepository is handed the composable default by the stock_quant onion. The type exists so a
// bespoke query has a home; it adds nothing today.
func NewStockQuantRepository(base composable.CrudRepository) itStock.StockQuantRepository {
	return &StockQuantRepositoryImpl{CrudRepository: base}
}

type StockQuantRepositoryImpl struct {
	composable.CrudRepository
}
