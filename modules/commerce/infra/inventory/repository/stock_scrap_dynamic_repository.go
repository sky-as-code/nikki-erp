package repository

import (
	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewStockScrapRepository is handed the composable default by the stock_scrap onion. The type exists so a
// bespoke query has a home; it adds nothing today.
func NewStockScrapRepository(base dynamicresource.CrudRepository) itStock.StockScrapRepository {
	return &StockScrapRepositoryImpl{CrudRepository: base}
}

type StockScrapRepositoryImpl struct {
	dynamicresource.CrudRepository
}
