package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// NewStockQuantDomainService derives the quant service from the engine's default one.
//
// base is the Stock Quant engine's own resource service, which this type embeds: every built-in
// action keeps running through the default implementation, and the available_quantity fill below
// is layered on top. The result is installed with Engine.SetResourceService.
func NewStockQuantDomainService(base drif.DynamicResourceService) *StockQuantDomainServiceImpl {
	return &StockQuantDomainServiceImpl{DynamicResourceService: base}
}

// StockQuantDomainServiceImpl fills a quant's available_quantity virtual field.
//
// The field has no database column: it is on-hand minus reserved, computed per row. Unlike the
// variant's template_* fields this needs no second query, because both operands are already on
// the row — so the fill is a local pass rather than a batched read.
type StockQuantDomainServiceImpl struct {
	drif.DynamicResourceService
}

var _ drif.DynamicResourceService = (*StockQuantDomainServiceImpl)(nil)

// Search fills available_quantity across the page.
//
// When the caller names fields it must also receive the two operands, or the arithmetic has
// nothing to work from. They are added to the projection rather than the result being left blank,
// because a caller asking for available_quantity wants the number, not an explanation.
func (this *StockQuantDomainServiceImpl) Search(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]], error) {
	requested := readRequestedFields(params)
	wantsAvailable := wantsAvailableQuantity(requested)

	if wantsAvailable && len(requested) > 0 {
		params[paramFieldNames] = appendMissing(requested, availableQuantityOperands()...)
	}

	result, err := this.DynamicResourceService.Search(ctx, params)
	if err != nil || result == nil || !result.HasData || !wantsAvailable {
		return result, err
	}

	for _, row := range result.Data.Items {
		FillAvailableQuantity(row)
	}
	return result, nil
}

// GetById fills available_quantity on a single quant.
func (this *StockQuantDomainServiceImpl) GetById(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.SingleResultData[dmodel.DynamicFields]], error) {
	return this.getOneWithAvailable(ctx, params, this.DynamicResourceService.GetById)
}

// GetOne fills available_quantity on a quant fetched by any unique key.
func (this *StockQuantDomainServiceImpl) GetOne(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.SingleResultData[dmodel.DynamicFields]], error) {
	return this.getOneWithAvailable(ctx, params, this.DynamicResourceService.GetOne)
}

func (this *StockQuantDomainServiceImpl) getOneWithAvailable(
	ctx corectx.Context, params dmodel.DynamicFields, delegate getOneFn,
) (*dyn.OpResult[dyn.SingleResultData[dmodel.DynamicFields]], error) {
	requested := readRequestedFields(params)
	wantsAvailable := wantsAvailableQuantity(requested)

	if wantsAvailable && len(requested) > 0 {
		params[paramFieldNames] = appendMissing(requested, availableQuantityOperands()...)
	}

	result, err := delegate(ctx, params)
	if err != nil || result == nil || !result.HasData || !wantsAvailable {
		return result, err
	}

	FillAvailableQuantity(result.Data.Item)
	return result, nil
}
