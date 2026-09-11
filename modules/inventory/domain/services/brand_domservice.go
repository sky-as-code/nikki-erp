package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// NewBrandDomainService is handed the composable default by the brand onion. The resource is plain
// CRUD; the type exists so a rule has a home when one is needed.
func NewBrandDomainService(base composable.CrudDomainService) itProduct.BrandDomainService {
	return &BrandDomainServiceImpl{CrudDomainService: base}
}

type BrandDomainServiceImpl struct {
	composable.CrudDomainService
}

// Delete refuses to remove a brand still carried by a product template. Without the guard the
// template foreign key refuses the statement and the failure reaches the client as a 500.
func (this *BrandDomainServiceImpl) Delete(
	ctx corectx.Context, params dmodel.DynamicFields, options ...composable.DeleteOptions,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	templateRepo, err := repoFor(models.ProductTemplateSchemaName)
	if err != nil {
		return nil, err
	}

	vErrs := ft.NewClientErrors()
	brandId := readStringParam(params, models.BrandFieldId)
	if err := AssertBrandDeletable(ctx, templateRepo, brandId, vErrs); err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}
	return this.CrudDomainService.Delete(ctx, params, options...)
}

// AssertBrandDeletable blocks the delete once any template references the brand.
func AssertBrandDeletable(
	ctx corectx.Context, repo models.ProductSearcher, brandId string, vErrs *ft.ClientErrors,
) error {
	if brandId == "" {
		return nil
	}

	templates, err := models.FindTemplatesByBrand(ctx, repo, brandId, 1)
	if err != nil {
		return errors.Wrap(err, "AssertBrandDeletable")
	}
	if len(templates) > 0 {
		vErrs.Append(*ft.NewBusinessViolation(models.BrandFieldId,
			"brand.has_templates",
			"this brand is still used by product templates"))
	}
	return nil
}
