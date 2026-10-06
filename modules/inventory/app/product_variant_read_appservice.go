package app

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/filefield"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/filestorage"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/services"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

func NewProductVariantReadService(
	domain itProduct.ProductVariantDomainService, storage filestorage.FileStorageAdapter,
) *ProductVariantReadServiceImpl {
	return &ProductVariantReadServiceImpl{domain: domain, storage: storage}
}

type ProductVariantReadServiceImpl struct {
	domain  itProduct.ProductVariantDomainService
	storage filestorage.FileStorageAdapter
}

var _ itProduct.ProductVariantService = (*ProductVariantReadServiceImpl)(nil)

func (this *ProductVariantReadServiceImpl) SearchProductVariants(
	ctx corectx.Context, params itProduct.SearchProductVariantsParams,
) (*itProduct.SearchProductVariantsResult, error) {
	scoped := params.SearchQuery
	scoped.Graph = dmodel.MergeAndCondition(
		scoped.Graph, models.ProductVariantFieldOrgId, dmodel.Equals, string(params.OrgId),
	)

	result, err := this.domain.SearchProductVariants(ctx, scoped)
	if err != nil || result == nil || !result.HasData {
		return result, err
	}

	fields := services.ProductVariantFileFields()
	for _, variant := range result.Data.Items {
		if err := filefield.PresignFileFields(ctx, this.storage, variant.GetFieldData(), fields); err != nil {
			return nil, errors.Wrap(err, "SearchProductVariants")
		}
	}

	return result, nil
}
