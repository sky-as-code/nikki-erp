package services

import (
	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewStockOperationTypeDomainService is handed the composable default by the stock_operation_type onion. The resource is plain
// CRUD; the type exists so a rule has a home when one is needed.
func NewStockOperationTypeDomainService(base composable.CrudDomainService) itStock.StockOperationTypeDomainService {
	return &StockOperationTypeDomainServiceImpl{CrudDomainService: base}
}

type StockOperationTypeDomainServiceImpl struct {
	composable.CrudDomainService
}
