package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// NewStockTransferRepository is handed the composable default by the stock_transfer onion. The type exists so a
// bespoke query has a home; it adds nothing today.
func NewStockTransferRepository(base composable.CrudRepository) itStock.StockTransferRepository {
	return &StockTransferRepositoryImpl{CrudRepository: base}
}

type StockTransferRepositoryImpl struct {
	composable.CrudRepository
}
