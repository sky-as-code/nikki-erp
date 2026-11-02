package app

import (
	"go.bryk.io/pkg/errors"

	"github.com/samber/lo"
	itEvent "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/event"
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewProductVariantApplicationService is handed the composable default by the variant onion, plus
// the Products capability of the template onion: the effective product is a template read.
func NewProductVariantApplicationService(
	base dynamicresource.CrudApplicationService, catalogChangedPub itEvent.CatalogChangedEventPublisher,
	productSvc itProduct.ProductService,
) itProduct.ProductVariantApplicationService {
	return &ProductVariantApplicationServiceImpl{
		CrudApplicationService: base,
		productSvc:             productSvc,
		catalogChangedPub:      catalogChangedPub,
	}
}

type ProductVariantApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
	productSvc        itProduct.ProductService
	catalogChangedPub itEvent.CatalogChangedEventPublisher
}

// GetEffective flattens a variant together with its template.
func (this *ProductVariantApplicationServiceImpl) GetEffective(
	ctx corectx.Context, query itProduct.GetEffectiveVariantQuery,
) (*itProduct.GetEffectiveVariantResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, dynamicresource.PermissionRead, query); err != nil || cErrs != nil {
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

func (this *ProductVariantApplicationServiceImpl) Create(
	ctx corectx.Context, cmd dynamicresource.CreateCommand,
) (*dynamicresource.CreateResult, error) {
	res, err := this.CrudApplicationService.Create(ctx, cmd)
	if err == nil && res.HasData && res.ClientErrors.Count() == 0 {
		this.catalogChangedPub.PublishAsync(ctx, itEvent.CatalogChangedProductVariant, lo.FromPtr(res.Data.GetString(basemodel.FieldOrgId)), lo.FromPtr(res.Data.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *ProductVariantApplicationServiceImpl) Update(
	ctx corectx.Context, cmd dynamicresource.UpdateCommand,
) (*dynamicresource.MutateResult, error) {
	res, err := this.CrudApplicationService.Update(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.catalogChangedPub.PublishAsync(ctx, itEvent.CatalogChangedProductVariant, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *ProductVariantApplicationServiceImpl) Delete(
	ctx corectx.Context, cmd dynamicresource.DeleteCommand,
) (*dynamicresource.MutateResult, error) {
	res, err := this.CrudApplicationService.Delete(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.catalogChangedPub.PublishAsync(ctx, itEvent.CatalogChangedProductVariant, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *ProductVariantApplicationServiceImpl) SetArchived(
	ctx corectx.Context, cmd dynamicresource.SetArchivedCommand,
) (*dynamicresource.MutateResult, error) {
	res, err := this.CrudApplicationService.SetArchived(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.catalogChangedPub.PublishAsync(ctx, itEvent.CatalogChangedProductVariant, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}
