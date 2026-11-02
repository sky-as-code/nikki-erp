package repository

import (
	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewStockTransferRepository is handed the composable default by the stock_transfer onion. The type exists so a
// bespoke query has a home; it adds nothing today.
func NewStockTransferRepository(base dynamicresource.CrudRepository) itStock.StockTransferRepository {
	return &StockTransferRepositoryImpl{CrudRepository: base}
}

type StockTransferRepositoryImpl struct {
	dynamicresource.CrudRepository
}
