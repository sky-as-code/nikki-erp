package app

import (
	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/filefield"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/filestorage"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/services"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// NewProductVariantApplicationService is handed the composable default by the variant onion, plus
// the Products capability of the template onion: the effective product is a template read.
func NewProductVariantApplicationService(
	base composable.CrudApplicationService, productSvc itProduct.ProductService, storage filestorage.FileStorageAdapter,
) itProduct.ProductVariantApplicationService {
	return &ProductVariantApplicationServiceImpl{
		fileBackedApplicationService: newFileBackedApplicationService(
			base, storage, services.ProductVariantFileFields()),
		productSvc: productSvc,
	}
}

type ProductVariantApplicationServiceImpl struct {
	fileBackedApplicationService
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

	view := itProduct.NewEffectiveProductView(result.Data.Product)
	if view.ImageKey != "" {
		url, err := this.storage.GeneratePresignedUrl(ctx.InnerContext(), view.ImageKey, filefield.PresignedUrlTtl)
		if err != nil {
			return nil, errors.Wrap(err, "GetEffective: failed to presign the image")
		}
		view.Image = url
	}
	return anyResult(view), nil
}
