package services

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// NewStockMoveDomainService is handed the composable default by the stock_move onion. The resource is plain
// CRUD; the type exists so a rule has a home when one is needed.
func NewStockMoveDomainService(base composable.CrudDomainService) itStock.StockMoveDomainService {
	return &StockMoveDomainServiceImpl{CrudDomainService: base}
}

type StockMoveDomainServiceImpl struct {
	composable.CrudDomainService
}
