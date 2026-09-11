package app

import (
	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// NewProductVariantApplicationService is handed the composable default by the variant onion, plus
// the Products capability of the template onion: the effective product is a template read.
func NewProductVariantApplicationService(
	base composable.CrudApplicationService, productSvc itProduct.ProductService,
) itProduct.ProductVariantApplicationService {
	return &ProductVariantApplicationServiceImpl{CrudApplicationService: base, productSvc: productSvc}
}

type ProductVariantApplicationServiceImpl struct {
	composable.CrudApplicationService
	productSvc itProduct.ProductService
}

// GetEffective flattens a variant together with its template.
func (this *ProductVariantApplicationServiceImpl) GetEffective(
	ctx corectx.Context, query itProduct.GetEffectiveVariantQuery,
) (*itProduct.GetEffectiveVariantResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, composable.PermissionRead, query); err != nil || cErrs != nil {
		return anyFailure(cErrs, err)
	}

	result, err := this.productSvc.GetEffectiveProduct(ctx, itProduct.GetEffectiveProductQuery{
		VariantId: readStringField(query, paramRecordId),
	})
	if err != nil {
		return nil, errors.Wrap(err, "GetEffective")
	}
	if result.ClientErrors.Count() > 0 {
		return anyFailure(&result.ClientErrors, nil)
	}
	if !result.HasData {
		return &itProduct.GetEffectiveVariantResult{}, nil
	}
	return anyResult(itProduct.NewEffectiveProductView(result.Data.Product)), nil
}
