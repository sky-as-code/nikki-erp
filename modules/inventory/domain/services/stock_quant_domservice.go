package services

import (
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// NewStockQuantDomainService derives the quant service from the engine's default one, which it
// embeds so built-in actions keep running unchanged. Installed with Engine.SetResourceService.
//
// available_quantity is a computed field declared in stock_quant.json, evaluated by the engine on
// every read.
func NewStockQuantDomainService(base drif.DynamicResourceService) *StockQuantDomainServiceImpl {
	return &StockQuantDomainServiceImpl{DynamicResourceService: base}
}

// StockQuantDomainServiceImpl carries the quant's domain behaviors: counting, adjustment,
// reservation and location-usage reads.
type StockQuantDomainServiceImpl struct {
	drif.DynamicResourceService
}

var _ drif.DynamicResourceService = (*StockQuantDomainServiceImpl)(nil)
