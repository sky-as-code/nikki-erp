package services

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// NewStockOperationTypeDomainService is handed the composable default by the stock_operation_type onion. The resource is plain
// CRUD; the type exists so a rule has a home when one is needed.
func NewStockOperationTypeDomainService(base composable.CrudDomainService) itStock.StockOperationTypeDomainService {
	return &StockOperationTypeDomainServiceImpl{CrudDomainService: base}
}

type StockOperationTypeDomainServiceImpl struct {
	composable.CrudDomainService
}
