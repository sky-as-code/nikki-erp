package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// NewStockOperationTypeRepository is handed the composable default by the stock_operation_type onion. The type exists so a
// bespoke query has a home; it adds nothing today.
func NewStockOperationTypeRepository(base composable.CrudRepository) itStock.StockOperationTypeRepository {
	return &StockOperationTypeRepositoryImpl{CrudRepository: base}
}

type StockOperationTypeRepositoryImpl struct {
	composable.CrudRepository
}
