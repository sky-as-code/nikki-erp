package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
)

// A move of the delta may name its product by template and attribute choices instead of by
// variant: the editor groups moves that way. The choices resolve to the complete variant exactly as
// resolve_selection does with materialize_if_missing, inside the transfer's transaction, so a
// variant materialized for a refused delta is rolled back with it.

const (
	moveFieldTemplateId  = "product_template_id"
	moveFieldAttributes  = "attributes"
	moveFieldAttributeId = "product_attribute_id"
	moveFieldValueId     = "product_attribute_value_id"
)

// completeVariantResolverFor is a variable so a test can substitute its own resolver: the hub is
// filled during Init, which a unit test cannot run.
var completeVariantResolverFor = func() (itProduct.CompleteVariantResolver, error) {
	templateSvc, err := domainServiceFor(models.ProductTemplateSchemaName)
	if err != nil {
		return nil, err
	}
	resolver, ok := templateSvc.(itProduct.CompleteVariantResolver)
	if !ok {
		return nil, errors.New("the product template service cannot resolve complete variants")
	}
	return resolver, nil
}

// moveProductForm is how one delta item names its product, for the shape check.
type moveProductForm struct {
	variantId  string
	templateId *string
	attributes []itStock.TransferMoveAttribute
	required   bool
}

// validateMoveProductForm accepts exactly one of product_variant_id and product_template_id (with
// its attributes). required is set for an addition, which must name a product.
func validateMoveProductForm(form moveProductForm, field string) ft.ClientErrors {
	if form.variantId != "" && form.templateId != nil {
		return ft.ClientErrors{*ft.NewValidationError(field+"."+moveFieldTemplateId,
			ft.ErrorKey("stock_transfer.move_product_ambiguous", "inventory"),
			"name the product either by product_variant_id or by product_template_id with attributes, not both")}
	}
	if form.templateId == nil {
		if form.attributes != nil {
			return ft.ClientErrors{*ft.NewValidationError(field+"."+moveFieldAttributes,
				ft.ErrorKey("stock_transfer.move_attributes_need_template", "inventory"),
				"attributes are read only together with product_template_id")}
		}
		if form.required && form.variantId == "" {
			return ft.ClientErrors{*dmodel.NewMissingFieldErr(field + "." + models.StockMoveFieldProductVariantId)}
		}
		return nil
	}
	if *form.templateId == "" {
		return ft.ClientErrors{*dmodel.NewMissingFieldErr(field + "." + moveFieldTemplateId)}
	}
	return validateMoveAttributes(form.attributes, field+"."+moveFieldAttributes)
}

func validateMoveAttributes(attributes []itStock.TransferMoveAttribute, field string) ft.ClientErrors {
	var out ft.ClientErrors
	chosen := make(map[string]bool, len(attributes))
	for j, pair := range attributes {
		row := valuesDeltaRowPath(field, j)
		if pair.ProductAttributeId == "" {
			out = append(out, *dmodel.NewMissingFieldErr(row + "." + moveFieldAttributeId))
		}
		if pair.ProductAttributeValueId == "" {
			out = append(out, *dmodel.NewMissingFieldErr(row + "." + moveFieldValueId))
		}
		if pair.ProductAttributeId == "" {
			continue
		}
		if chosen[pair.ProductAttributeId] {
			out = append(out, *ft.NewBusinessViolation(row+"."+moveFieldAttributeId,
				ft.ErrorKey("product_template.selection_duplicate_attribute", "inventory"),
				"only one value may be chosen for attribute {{attribute_id}}",
				map[string]any{"attribute_id": pair.ProductAttributeId}))
		}
		chosen[pair.ProductAttributeId] = true
	}
	return out
}

// resolveTemplateForms replaces every template form of the delta by the variant it resolves to,
// so the writes that follow see plain variant ids.
func (this *transferMovesWriter) resolveTemplateForms(
	ctx corectx.Context, delta itStock.TransferMovesDelta,
) (ft.ClientErrors, error) {
	for i := range delta.Add {
		item := &delta.Add[i]
		if item.ProductTemplateId == nil {
			continue
		}
		variantId, cErrs, err := this.resolveMoveVariant(ctx, *item.ProductTemplateId, item.Attributes,
			valuesDeltaRowPath(movesFieldAdd, i))
		if err != nil || cErrs.Count() > 0 {
			return cErrs, err
		}
		item.ProductVariantId = variantId
	}
	for i := range delta.Update {
		item := &delta.Update[i]
		if item.ProductTemplateId == nil {
			continue
		}
		variantId, cErrs, err := this.resolveMoveVariant(ctx, *item.ProductTemplateId, item.Attributes,
			valuesDeltaRowPath(movesFieldUpdate, i))
		if err != nil || cErrs.Count() > 0 {
			return cErrs, err
		}
		item.ProductVariantId = &variantId
	}
	return nil, nil
}

// resolveMoveVariant resolves one template form. The template must belong to the transfer's org,
// as resolve_selection confines it to the caller's.
func (this *transferMovesWriter) resolveMoveVariant(
	ctx corectx.Context, templateId string, attributes []itStock.TransferMoveAttribute, field string,
) (string, ft.ClientErrors, error) {
	templateField := field + "." + moveFieldTemplateId
	inOrg, err := templateInOrg(ctx, templateId, derefString(this.transfer.GetOrgId()))
	if err != nil || !inOrg {
		return "", ft.ClientErrors{*ft.NewNotFoundError(templateField)}, err
	}
	if this.resolver == nil {
		if this.resolver, err = completeVariantResolverFor(); err != nil {
			return "", nil, err
		}
	}
	result, err := this.resolver.ResolveCompleteVariant(ctx, itProduct.ResolveCompleteVariantQuery{
		TemplateId: templateId, Selections: moveSelections(attributes), MaterializeIfMissing: true,
	})
	if err != nil {
		return "", nil, errors.Wrap(err, "resolve the move's variant")
	}
	if result.ClientErrors.Count() > 0 {
		return "", pointResolveErrors(result.ClientErrors, field), nil
	}
	if !result.HasData {
		return "", ft.ClientErrors{*ft.NewNotFoundError(templateField)}, nil
	}
	return acceptResolvedVariant(result.Data, attributes, field)
}

// acceptResolvedVariant refuses a choice on a never line, which no variant carries and a move has
// nowhere to keep, and an answer that names no variant.
func acceptResolvedVariant(
	data itProduct.ResolveCompleteVariantResultData, attributes []itStock.TransferMoveAttribute, field string,
) (string, ft.ClientErrors, error) {
	var out ft.ClientErrors
	for _, never := range data.NeverSelections {
		out = append(out, *ft.NewBusinessViolation(attributeRowPath(attributes, never.AttributeId, field),
			ft.ErrorKey("stock_transfer.move_attribute_not_variant", "inventory"),
			"attribute {{attribute_id}} does not create variants, so a move cannot carry it",
			map[string]any{"attribute_id": never.AttributeId}))
	}
	if out.Count() > 0 {
		return "", out, nil
	}
	if data.VariantId == "" || data.Outcome == itProduct.ResolveOutcomeMissing {
		return "", ft.ClientErrors{*ft.NewBusinessViolation(field+"."+moveFieldAttributes,
			ft.ErrorKey("stock_transfer.move_variant_unresolved", "inventory"),
			"no variant could be found or created for these attribute values")}, nil
	}
	return data.VariantId, nil, nil
}

func moveSelections(attributes []itStock.TransferMoveAttribute) map[string]string {
	selections := make(map[string]string, len(attributes))
	for _, pair := range attributes {
		selections[pair.ProductAttributeId] = pair.ProductAttributeValueId
	}
	return selections
}

// pointResolveErrors moves the resolver's errors from its own `selections` field onto the delta
// row: a refusal of the template onto product_template_id, a refusal of the choices onto attributes.
func pointResolveErrors(cErrs ft.ClientErrors, field string) ft.ClientErrors {
	out := make(ft.ClientErrors, 0, len(cErrs))
	templateArchived := ft.ErrorKey(ReasonResolveTemplateArchived, "inventory")
	for _, item := range cErrs {
		switch {
		case item.Key == templateArchived:
			item.Field = field + "." + moveFieldTemplateId
		case item.Field == PlanFieldSelection || item.Field == "":
			item.Field = field + "." + moveFieldAttributes
		default:
			item.Field = field + "." + item.Field
		}
		out = append(out, item)
	}
	return out
}

// attributeRowPath names the attributes row choosing an attribute, or the list when none does.
func attributeRowPath(attributes []itStock.TransferMoveAttribute, attributeId string, field string) string {
	list := field + "." + moveFieldAttributes
	for j, pair := range attributes {
		if pair.ProductAttributeId == attributeId {
			return valuesDeltaRowPath(list, j)
		}
	}
	return list
}

func templateInOrg(ctx corectx.Context, templateId string, orgId string) (bool, error) {
	repo, err := repoFor(models.ProductTemplateSchemaName)
	if err != nil {
		return false, err
	}
	filter := dmodel.DynamicFields{basemodel.FieldId: templateId}
	if orgId != "" {
		filter[basemodel.FieldOrgId] = orgId
	}
	found, err := repo.GetOne(ctx, dyn.RepoGetOneParam{Filter: filter})
	if err != nil {
		return false, errors.Wrap(err, "templateInOrg")
	}
	return found.HasData, nil
}
