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

// NewProductTypeDomainService is handed the composable default by the product_type onion. The resource is plain
// CRUD; the type exists so a rule has a home when one is needed.
func NewProductTypeDomainService(base composable.CrudDomainService) itProduct.ProductTypeDomainService {
	return &ProductTypeDomainServiceImpl{CrudDomainService: base}
}

type ProductTypeDomainServiceImpl struct {
	composable.CrudDomainService
}

// Delete refuses to remove a product type still carried by a product template. Without the guard
// the template foreign key refuses the statement and the failure reaches the client as a 500.
func (this *ProductTypeDomainServiceImpl) Delete(
	ctx corectx.Context, params dmodel.DynamicFields, options ...composable.DeleteOptions,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	templateRepo, err := repoFor(models.ProductTemplateSchemaName)
	if err != nil {
		return nil, err
	}

	vErrs := ft.NewClientErrors()
	productTypeId := readStringParam(params, models.ProductTypeFieldId)
	if err := AssertProductTypeDeletable(ctx, templateRepo, productTypeId, vErrs); err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}
	return this.CrudDomainService.Delete(ctx, params, options...)
}

// AssertProductTypeDeletable blocks the delete once any template references the product type.
func AssertProductTypeDeletable(
	ctx corectx.Context, repo models.ProductSearcher, productTypeId string, vErrs *ft.ClientErrors,
) error {
	if productTypeId == "" {
		return nil
	}

	templates, err := models.FindTemplatesByProductType(ctx, repo, productTypeId, 1)
	if err != nil {
		return errors.Wrap(err, "AssertProductTypeDeletable")
	}
	if len(templates) > 0 {
		vErrs.Append(*ft.NewBusinessViolation(models.ProductTypeFieldId,
			"product_type.has_templates",
			"this product type is still used by product templates"))
	}
	return nil
}
