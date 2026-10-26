package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
	itQuotation "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/quotation"
)

type SalesQuotationApplicationServiceImpl struct {
	composable.CrudApplicationService

	effectiveSettings itExt.EffectiveSettingsExtService
	taxCalculation    itExt.TaxCalculationExtService
	productVariants   itExt.ProductVariantExtService
	pricingBasis      itExt.ProductPricingBasisExtService
}

func NewSalesQuotationApplicationService(
	base composable.CrudApplicationService,
	settings itExt.EffectiveSettingsExtService,
	tax itExt.TaxCalculationExtService,
	products itExt.ProductVariantExtService,
	basis itExt.ProductPricingBasisExtService,
) itQuotation.SalesQuotationApplicationService {
	return &SalesQuotationApplicationServiceImpl{
		CrudApplicationService: base,
		effectiveSettings:      settings,
		taxCalculation:         tax,
		productVariants:        products,
		pricingBasis:           basis,
	}
}

func (this *SalesQuotationApplicationServiceImpl) Convert(
	ctx corectx.Context, cmd itQuotation.QuotationActionCommand,
) (*dyn.OpResult[any], error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionConvertQuotation, cmd); cErrs != nil || err != nil {
		return anyFailure(cErrs, err)
	}
	return this.runConvertQuotation(ctx, cmd)
}

func (this *SalesQuotationApplicationServiceImpl) Send(
	ctx corectx.Context, cmd itQuotation.QuotationActionCommand,
) (*dyn.OpResult[any], error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionTransitionQuotation, cmd); cErrs != nil || err != nil {
		return anyFailure(cErrs, err)
	}
	return this.runSendQuotation(ctx, cmd)
}

func (this *SalesQuotationApplicationServiceImpl) Cancel(
	ctx corectx.Context, cmd itQuotation.QuotationActionCommand,
) (*dyn.OpResult[any], error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionTransitionQuotation, cmd); cErrs != nil || err != nil {
		return anyFailure(cErrs, err)
	}
	return this.runCancelQuotation(ctx, cmd)
}

func NewSalesQuotationLineApplicationService(base composable.CrudApplicationService) itQuotation.SalesQuotationLineApplicationService {
	return &SalesQuotationLineApplicationServiceImpl{CrudApplicationService: base}
}

type SalesQuotationLineApplicationServiceImpl struct {
	composable.CrudApplicationService
}

// The moved action bodies follow.

func (this *SalesQuotationApplicationServiceImpl) runConvertQuotation(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[any], error) {
	policy := services.ResolveSalesPolicy(ctx, this.effectiveSettings)

	result, vErrs, err := services.ConvertQuotation(ctx, services.ConvertQuotationParams{
		SalesQuotationId: readStringParam(params, paramRecordId),
		SalesPointId:     readStringParam(params, "sales_point_id"),
		IdempotencyKey:   readStringParam(params, "idempotency_key"),
	}, this.taxCalculation, this.productVariants, this.pricingBasis, policy)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &dyn.OpResult[any]{ClientErrors: *vErrs}, nil
	}

	return &dyn.OpResult[any]{
		HasData: true,
		Data: map[string]any{
			"sales_quotation_id": result.SalesQuotationId,
			"sales_order_id":     result.SalesOrderId,
			"order_number":       result.OrderNumber,
			"already_converted":  result.AlreadyConverted,

			// Both totals, so the caller can see whether repricing moved the number.
			"quoted_total": result.QuotedTotal,
			"order_total":  result.OrderTotal,
		},
	}, nil
}

func (this *SalesQuotationApplicationServiceImpl) runSendQuotation(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[any], error) {
	return transitionQuotationResult(ctx,
		readStringParam(params, paramRecordId), string(models.SalesQuotationStatusSent))
}

func (this *SalesQuotationApplicationServiceImpl) runCancelQuotation(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[any], error) {
	return transitionQuotationResult(ctx,
		readStringParam(params, paramRecordId), string(models.SalesQuotationStatusCancelled))
}

func transitionQuotationResult(
	ctx corectx.Context, quotationId, toStatus string,
) (*dyn.OpResult[any], error) {
	vErrs, err := services.TransitionQuotation(ctx, quotationId, toStatus)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &dyn.OpResult[any]{ClientErrors: *vErrs}, nil
	}
	return &dyn.OpResult[any]{
		HasData: true,
		Data: map[string]any{
			"sales_quotation_id": quotationId,
			"status":             toStatus,
		},
	}, nil
}
