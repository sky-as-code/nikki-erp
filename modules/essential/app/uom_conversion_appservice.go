package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"
)

func NewUomConversionApplicationServiceImpl(
	conversionSvc itUom.UomConversionDomainService,
) itUom.UomConversionAppService {
	return &UomConversionApplicationServiceImpl{conversionSvc: conversionSvc}
}

// UomConversionApplicationServiceImpl is the capability boundary other modules bind to.
// It stays a thin delegation on purpose: when Essential is split into its own service, this
// is the type a REST client replaces, and any logic living here would have to be duplicated.
type UomConversionApplicationServiceImpl struct {
	conversionSvc itUom.UomConversionDomainService
}

func (this *UomConversionApplicationServiceImpl) Convert(
	ctx corectx.Context, query itUom.ConvertQuantityQuery,
) (*itUom.ConvertQuantityResult, error) {
	return this.conversionSvc.Convert(ctx, query)
}

func (this *UomConversionApplicationServiceImpl) GetUom(
	ctx corectx.Context, query itUom.GetUomQuery,
) (*itUom.GetUomResult, error) {
	return this.conversionSvc.GetUom(ctx, query)
}

func (this *UomConversionApplicationServiceImpl) ToReference(
	ctx corectx.Context, query itUom.ToReferenceQuery,
) (*itUom.ToReferenceResult, error) {
	return this.conversionSvc.ToReference(ctx, query)
}
