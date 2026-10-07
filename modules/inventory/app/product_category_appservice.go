package app

import (
	"github.com/samber/lo"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itEvent "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/event"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// NewProductCategoryApplicationService is handed the composable default by the product_category onion.
func NewProductCategoryApplicationService(
	base composable.CrudApplicationService, catalogChangedPub itEvent.CatalogChangedEventPublisher,
) itProduct.ProductCategoryApplicationService {
	return &ProductCategoryApplicationServiceImpl{CrudApplicationService: base, catalogChangedPub: catalogChangedPub}
}

// ProductCategoryApplicationServiceImpl is the authorized CRUD of the resource.
type ProductCategoryApplicationServiceImpl struct {
	composable.CrudApplicationService

	catalogChangedPub itEvent.CatalogChangedEventPublisher
}

func (this *ProductCategoryApplicationServiceImpl) Create(
	ctx corectx.Context, cmd composable.CreateCommand,
) (*composable.CreateResult, error) {
	res, err := this.CrudApplicationService.Create(ctx, cmd)
	if err == nil && res.HasData && res.ClientErrors.Count() == 0 {
		this.catalogChangedPub.PublishAsync(ctx, itEvent.CatalogChangedCategory, lo.FromPtr(res.Data.GetString(basemodel.FieldOrgId)), lo.FromPtr(res.Data.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *ProductCategoryApplicationServiceImpl) Update(
	ctx corectx.Context, cmd composable.UpdateCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.Update(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.catalogChangedPub.PublishAsync(ctx, itEvent.CatalogChangedCategory, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *ProductCategoryApplicationServiceImpl) Delete(
	ctx corectx.Context, cmd composable.DeleteCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.Delete(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.catalogChangedPub.PublishAsync(ctx, itEvent.CatalogChangedCategory, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}

func (this *ProductCategoryApplicationServiceImpl) SetArchived(
	ctx corectx.Context, cmd composable.SetArchivedCommand,
) (*composable.MutateResult, error) {
	res, err := this.CrudApplicationService.SetArchived(ctx, cmd)
	if err == nil && res.ClientErrors.Count() == 0 {
		this.catalogChangedPub.PublishAsync(ctx, itEvent.CatalogChangedCategory, lo.FromPtr(cmd.GetString(basemodel.FieldOrgId)), lo.FromPtr(cmd.GetString(basemodel.FieldId)))
	}

	return res, err
}
