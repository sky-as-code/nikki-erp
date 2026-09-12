package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// NewProductVariantService adapts the domain reader for a caller with no user permission of its
// own — a device certificate proves which machine it is, and nothing about who is operating it —
// so it must supply the org itself rather than have it asserted from the caller's membership.
//
// A separate type rather than another method on ProductVariantDomainServiceImpl: Go allows only
// one method named SearchProductVariants per receiver, and that name is already the permission-
// checked read's.
func NewProductVariantService(domain itProduct.ProductVariantDomainService) *ProductVariantServiceImpl {
	return &ProductVariantServiceImpl{domain: domain}
}

type ProductVariantServiceImpl struct {
	domain itProduct.ProductVariantDomainService
}

var _ itProduct.ProductVariantService = (*ProductVariantServiceImpl)(nil)

func (this *ProductVariantServiceImpl) SearchProductVariants(
	ctx corectx.Context, params itProduct.SearchProductVariantsParams,
) (*itProduct.SearchProductVariantsResult, error) {
	scoped := params.SearchQuery
	scoped.Graph = dmodel.MergeAndCondition(
		scoped.Graph, models.ProductVariantFieldOrgId, dmodel.Equals, string(params.OrgId),
	)
	return this.domain.SearchProductVariants(ctx, scoped)
}
