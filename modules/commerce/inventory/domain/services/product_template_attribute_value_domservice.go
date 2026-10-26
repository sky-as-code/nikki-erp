package services

import (
	"sort"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// NewProductTemplateAttributeValueDomainService is handed the composable default by the
// product_template_attribute_value onion.
func NewProductTemplateAttributeValueDomainService(base composable.CrudDomainService) itProduct.ProductTemplateAttributeValueDomainService {
	return &ProductTemplateAttributeValueDomainServiceImpl{CrudDomainService: base}
}

type ProductTemplateAttributeValueDomainServiceImpl struct {
	composable.CrudDomainService
}

// templateValueEditableFields are what a direct update may carry. Selection, sequence and the
// references belong to the template's configuration payload, which reconciles the variants; the
// price surcharge changes no variant identity, so it is edited in place.
var templateValueEditableFields = map[string]bool{
	basemodel.FieldId:    true,
	basemodel.FieldEtag:  true,
	basemodel.FieldOrgId: true,
	models.ProductTemplateAttributeValueFieldSalesPriceExtra: true,
}

// Update accepts only sales_price_extra. The configuration writer updates the other columns
// through the repository, so this rule never stands in its way.
func (this *ProductTemplateAttributeValueDomainServiceImpl) Update(
	ctx corectx.Context, cmd composable.UpdateCommand, options ...composable.UpdateOptions,
) (*composable.MutateResult, error) {
	if cErrs := assertOnlySalesPriceExtra(cmd); cErrs.Count() > 0 {
		return &composable.MutateResult{ClientErrors: cErrs}, nil
	}
	result, err := this.CrudDomainService.Update(ctx, cmd, options...)
	if err != nil || result == nil || result.ClientErrors.Count() > 0 {
		return result, err
	}
	if _, touched := cmd[models.ProductTemplateAttributeValueFieldSalesPriceExtra]; touched {
		if err := propagateSalesPriceExtra(ctx, derefString(cmd.GetString(basemodel.FieldId))); err != nil {
			return nil, err
		}
	}
	return result, nil
}

// propagateSalesPriceExtra re-copies the template value's surcharge onto every variant join row
// built from it: the variant's effective price sums those copies, so a stale copy is a wrong price.
func propagateSalesPriceExtra(ctx corectx.Context, templateValueId string) error {
	if templateValueId == "" {
		return nil
	}
	saved, err := searchConfigRows(ctx, models.ProductTemplateAttributeValueSchemaName,
		basemodel.FieldId, []string{templateValueId})
	if err != nil || len(saved) == 0 {
		return err
	}
	extra := saved[0].GetDecimal(models.ProductTemplateAttributeValueFieldSalesPriceExtra)
	joins, err := searchConfigRows(ctx, models.ProductVariantAttributeValueSchemaName,
		models.ProductVariantAttributeValueFieldTemplateAttributeValueId, []string{templateValueId})
	if err != nil {
		return err
	}
	joinRepo, err := repoFor(models.ProductVariantAttributeValueSchemaName)
	if err != nil {
		return err
	}
	for _, join := range joins {
		row := dmodel.DynamicFields{basemodel.FieldId: join[basemodel.FieldId]}
		row.SetDecimal(models.ProductVariantAttributeValueFieldSalesPriceExtra, extra)
		updated, err := joinRepo.Update(ctx, row)
		if err != nil {
			return errors.Wrap(err, "propagateSalesPriceExtra")
		}
		if updated.ClientErrors.Count() > 0 {
			return errors.Wrap(updated.ClientErrors.ToError(), "propagateSalesPriceExtra")
		}
	}
	return nil
}

func assertOnlySalesPriceExtra(cmd dmodel.DynamicFields) ft.ClientErrors {
	refused := make([]string, 0)
	for field := range cmd {
		if !templateValueEditableFields[field] {
			refused = append(refused, field)
		}
	}
	sort.Strings(refused)
	out := ft.ClientErrors{}
	for _, field := range refused {
		out.Append(*ft.NewValidationError(field,
			ft.ErrorKey("product_template_attribute_value.field_read_only", "inventory"),
			"only sales_price_extra can be edited here; change the selection on the product template"))
	}
	return out
}
