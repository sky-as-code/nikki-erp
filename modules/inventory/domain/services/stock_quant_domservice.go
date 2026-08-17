package services

import (
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// NewStockQuantDomainService derives the quant service from the engine's default one.
//
// base is the Stock Quant engine's own resource service, which this type embeds: every built-in
// action keeps running through the default implementation. The count, adjustment and reservation
// behaviors in the sibling files are layered on top. The result is installed with
// Engine.SetResourceService.
//
// available_quantity needs no handling here anymore: it is declared as a computed field in
// stock_quant.json, and the engine's computed-field layer evaluates it on every read.
func NewStockQuantDomainService(base drif.DynamicResourceService) *StockQuantDomainServiceImpl {
	return &StockQuantDomainServiceImpl{DynamicResourceService: base}
}

// StockQuantDomainServiceImpl carries the quant's domain behaviors: counting, adjustment,
// reservation and location-usage reads (see stock_count.go, stock_reservation.go,
// location_usage_read_service.go).
type StockQuantDomainServiceImpl struct {
	drif.DynamicResourceService
}

var _ drif.DynamicResourceService = (*StockQuantDomainServiceImpl)(nil)
