package app

import (
	"go.bryk.io/pkg/errors"

	"github.com/samber/lo"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/filefield"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/filestorage"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/services"
	itEvent "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/event"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// NewProductVariantApplicationService is handed the composable default by the variant onion, plus
// the Products capability of the template onion: the effective product is a template read.
func NewProductVariantApplicationService(
	base composable.CrudApplicationService, catalogChangedPub itEvent.CatalogChangedEventPublisher,
	productSvc itProduct.ProductService, storage filestorage.FileStorageAdapter,
) itProduct.ProductVariantApplicationService {
	return &ProductVariantApplicationServiceImpl{
		fileBackedApplicationService: newFileBackedApplicationService(
			base, storage, services.ProductVariantFileFields()),
		productSvc:        productSvc,
		catalogChangedPub: catalogChangedPub,
	}
}

type ProductVariantApplicationServiceImpl struct {
	fileBackedApplicationService
	productSvc        itProduct.ProductService
	catalogChangedPub itEvent.CatalogChangedEventPublisher
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
		view.ImageUrl = url
	}
	return anyResult(view), nil
}

func (this *ProductVariantApplicationServiceImpl) Create(
	ctx corectx.Context, cmd composable.CreateCommand,
) (*composable.CreateResult, error) {
	res, err := this.fileBackedApplicationService.Create(ctx, cmd)
	if err == nil && res.HasData && res.ClientErrors.Count() == 0 {
		this.catalogChangedPub.PublishAsync(ctx, itEvent.CatalogChangedProductVariant, lo.FromPtr(res.Data.GetString(basemodel.FieldOrgId)), lo.FromPtr(res.Data.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *ProductVariantApplicationServiceImpl) Update(
	ctx corectx.Context, cmd composable.UpdateCommand,
) (*composable.MutateResult, error) {
	res, err := this.fileBackedApplicationService.Update(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.catalogChangedPub.PublishAsync(ctx, itEvent.CatalogChangedProductVariant, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *ProductVariantApplicationServiceImpl) Delete(
	ctx corectx.Context, cmd composable.DeleteCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.Delete(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.catalogChangedPub.PublishAsync(ctx, itEvent.CatalogChangedProductVariant, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *ProductVariantApplicationServiceImpl) SetArchived(
	ctx corectx.Context, cmd composable.SetArchivedCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.SetArchived(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.catalogChangedPub.PublishAsync(ctx, itEvent.CatalogChangedProductVariant, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *ProductVariantApplicationServiceImpl) UploadFile(
	ctx corectx.Context, cmd itProduct.UploadFileCommand,
) (*itProduct.FileFieldResult, error) {
	res, err := this.fileBackedApplicationService.UploadFile(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.catalogChangedPub.PublishAsync(ctx, itEvent.CatalogChangedProductVariant, lo.FromPtr(cmd.Params.GetString(basemodel.FieldOrgId)), string(cmd.RecordId))
	}

	return res, err
}

func (this *ProductVariantApplicationServiceImpl) ReplaceFile(
	ctx corectx.Context, cmd itProduct.UploadFileCommand,
) (*itProduct.FileFieldResult, error) {
	res, err := this.fileBackedApplicationService.ReplaceFile(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.catalogChangedPub.PublishAsync(ctx, itEvent.CatalogChangedProductVariant, lo.FromPtr(cmd.Params.GetString(basemodel.FieldOrgId)), string(cmd.RecordId))
	}

	return res, err
}

func (this *ProductVariantApplicationServiceImpl) DeleteFile(
	ctx corectx.Context, cmd itProduct.DeleteFileCommand,
) (*itProduct.FileFieldResult, error) {
	res, err := this.fileBackedApplicationService.DeleteFile(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.catalogChangedPub.PublishAsync(ctx, itEvent.CatalogChangedProductVariant, lo.FromPtr(cmd.Params.GetString(basemodel.FieldOrgId)), string(cmd.RecordId))
	}

	return res, err
}
