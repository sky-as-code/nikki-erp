package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

func NewProductVariantReadService(domain itProduct.ProductVariantDomainService) *ProductVariantReadServiceImpl {
	return &ProductVariantReadServiceImpl{domain: domain}
}

type ProductVariantReadServiceImpl struct {
	domain itProduct.ProductVariantDomainService
}

var _ itProduct.ProductVariantService = (*ProductVariantReadServiceImpl)(nil)

func (this *ProductVariantReadServiceImpl) SearchProductVariants(
	ctx corectx.Context, params itProduct.SearchProductVariantsParams,
) (*itProduct.SearchProductVariantsResult, error) {
	scoped := params.SearchQuery
	scoped.Graph = dmodel.MergeAndCondition(
		scoped.Graph, models.ProductVariantFieldOrgId, dmodel.Equals, string(params.OrgId),
	)

	return this.domain.SearchProductVariants(ctx, scoped)
}
