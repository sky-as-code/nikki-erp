package services

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewStockQuantDomainService derives the quant service from the engine's default one, which it
// embeds so built-in actions keep running unchanged. Installed with Engine.SetResourceService.
//
// available_quantity is a computed field declared in stock_quant.json, evaluated by the engine on
// every read.
func NewStockQuantDomainService(base composable.CrudDomainService) *StockQuantDomainServiceImpl {
	return &StockQuantDomainServiceImpl{CrudDomainService: base}
}

// StockQuantDomainServiceImpl carries the quant's domain behaviors: counting, adjustment,
// reservation and location-usage reads.
type StockQuantDomainServiceImpl struct {
	composable.CrudDomainService
}

var _ composable.CrudDomainService = (*StockQuantDomainServiceImpl)(nil)
