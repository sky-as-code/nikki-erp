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

// NewProductAttributeDomainService is handed the composable default by the product_attribute onion. The resource is plain
// CRUD; the type exists so a rule has a home when one is needed.
func NewProductAttributeDomainService(base composable.CrudDomainService) itProduct.ProductAttributeDomainService {
	return &ProductAttributeDomainServiceImpl{CrudDomainService: base}
}

type ProductAttributeDomainServiceImpl struct {
	composable.CrudDomainService
}

// Delete refuses to remove an attribute that still owns values. Without the guard the value
// foreign key refuses the statement and the failure reaches the client as a 500.
func (this *ProductAttributeDomainServiceImpl) Delete(
	ctx corectx.Context, params dmodel.DynamicFields, options ...composable.DeleteOptions,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	valueRepo, err := repoFor(models.ProductAttributeValueSchemaName)
	if err != nil {
		return nil, err
	}

	vErrs := ft.NewClientErrors()
	attributeId := readStringParam(params, models.ProductAttributeFieldId)
	if err := AssertProductAttributeDeletable(ctx, valueRepo, attributeId, vErrs); err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}
	return this.CrudDomainService.Delete(ctx, params, options...)
}

// AssertProductAttributeDeletable blocks the delete once the attribute owns any value.
func AssertProductAttributeDeletable(
	ctx corectx.Context, repo models.ProductSearcher, attributeId string, vErrs *ft.ClientErrors,
) error {
	if attributeId == "" {
		return nil
	}

	values, err := models.FindAttributeValues(ctx, repo, attributeId, 1)
	if err != nil {
		return errors.Wrap(err, "AssertProductAttributeDeletable")
	}
	if len(values) > 0 {
		vErrs.Append(*ft.NewBusinessViolation(models.ProductAttributeFieldId,
			"product_attribute.has_values",
			"this attribute still has values; delete them first"))
	}
	return nil
}
