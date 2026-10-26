package services

import (
	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewStockMoveDependencyDomainService is handed the composable default by the stock_move_dependency onion. The resource is plain
// CRUD; the type exists so a rule has a home when one is needed.
func NewStockMoveDependencyDomainService(base composable.CrudDomainService) itStock.StockMoveDependencyDomainService {
	return &StockMoveDependencyDomainServiceImpl{CrudDomainService: base}
}

type StockMoveDependencyDomainServiceImpl struct {
	composable.CrudDomainService
}
