package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// maxTemplateVariants bounds the variant set one call will read or generate, past what a
// synchronous request should carry.
const maxTemplateVariants = 1000

// NewProductTemplateDomainService derives the Products service from the engine's default one, which
// it embeds so built-in CRUD keeps running unchanged. Installed with Engine.SetResourceService.
func NewProductTemplateDomainService(base drif.DynamicResourceService) itProduct.ProductService {
	return &ProductTemplateDomainServiceImpl{DynamicResourceService: base}
}

// ProductTemplateDomainServiceImpl serves the Products capabilities the resource engine cannot
// express: flattening a template and variant into one effective product, and turning an attribute
// selection into a concrete variant.
//
// It resolves other resources' engines at call time rather than holding them, because engine
// creation and this service's construction both happen during Init.
type ProductTemplateDomainServiceImpl struct {
	drif.DynamicResourceService
}

// The derived service must satisfy both contracts — the engine installs it as its resource service,
// and custom actions type-assert it back to the Products capability — since losing either would
// only show up as a failed assertion at request time.
var (
	_ drif.DynamicResourceService = (*ProductTemplateDomainServiceImpl)(nil)
	_ itProduct.ProductService    = (*ProductTemplateDomainServiceImpl)(nil)
)

func (this *ProductTemplateDomainServiceImpl) GetEffectiveProduct(
	ctx corectx.Context, query itProduct.GetEffectiveProductQuery,
) (*itProduct.GetEffectiveProductResult, error) {
	if query.VariantId == "" {
		return &itProduct.GetEffectiveProductResult{}, nil
	}

	variant, err := this.fetchVariant(ctx, query.VariantId)
	if err != nil || variant == nil {
		return &itProduct.GetEffectiveProductResult{}, err
	}

	effective, err := this.flatten(ctx, variant)
	if err != nil || effective == nil {
		return &itProduct.GetEffectiveProductResult{}, err
	}
	return &itProduct.GetEffectiveProductResult{
		Data:    itProduct.GetEffectiveProductResultData{Product: *effective},
		HasData: true,
	}, nil
}

func (this *ProductTemplateDomainServiceImpl) GetEffectiveProducts(
	ctx corectx.Context, query itProduct.GetEffectiveProductsQuery,
) (*itProduct.GetEffectiveProductsResult, error) {
	products := map[string]itProduct.EffectiveProduct{}

	for _, variantId := range query.VariantIds {
		one, err := this.GetEffectiveProduct(ctx, itProduct.GetEffectiveProductQuery{VariantId: variantId})
		if err != nil {
			return nil, err
		}
		if one.HasData {
			products[variantId] = one.Data.Product
		}
	}
	return &itProduct.GetEffectiveProductsResult{
		Data:    itProduct.GetEffectiveProductsResultData{Products: products},
		HasData: len(products) > 0,
	}, nil
}

// ResolveProductSelection turns the values a configurator picked into the variant a transaction
// line must reference.
func (this *ProductTemplateDomainServiceImpl) flatten(
	ctx corectx.Context, variant *models.ProductVariant,
) (*itProduct.EffectiveProduct, error) {
	template, err := this.fetchTemplate(ctx, derefString(variant.GetProductTemplateId()))
	if err != nil || template == nil {
		return nil, err
	}

	labels, err := this.valueLabels(ctx, variant)
	if err != nil {
		return nil, err
	}

	effective := BuildEffectiveProduct(template, variant, labels)
	return &effective, nil
}

// valueLabels resolves the display labels of a variant's attribute values, which turns the template
// name into "Classic T-Shirt / Black / M".
func (this *ProductTemplateDomainServiceImpl) valueLabels(
	ctx corectx.Context, variant *models.ProductVariant,
) ([]string, error) {
	templateValueIds := VariantValueIds(variant)
	if len(templateValueIds) == 0 {
		return nil, nil
	}

	templateValueEngine, err := engineFor(models.ProductTemplateAttributeValueSchemaName)
	if err != nil {
		return nil, err
	}
	valueEngine, err := engineFor(models.ProductAttributeValueSchemaName)
	if err != nil {
		return nil, err
	}

	labels := make([]string, 0, len(templateValueIds))
	for _, templateValueId := range templateValueIds {
		templateValue, err := templateValueEngine.ResourceRepository().GetOne(ctx, dyn.RepoGetOneParam{
			Filter: dmodel.DynamicFields{models.ProductTemplateAttributeValueFieldId: templateValueId},
			Fields: []string{
				models.ProductTemplateAttributeValueFieldId,
				models.ProductTemplateAttributeValueFieldAttributeValueId,
			},
		})
		if err != nil {
			return nil, errors.Wrap(err, "valueLabels")
		}
		if !templateValue.HasData {
			continue
		}

		globalValueId := derefString(
			models.NewProductTemplateAttributeValueFrom(templateValue.Data).GetAttributeValueId())
		if globalValueId == "" {
			continue
		}

		value, err := valueEngine.ResourceRepository().GetOne(ctx, dyn.RepoGetOneParam{
			Filter: dmodel.DynamicFields{models.ProductAttributeValueFieldId: globalValueId},
			Fields: []string{models.ProductAttributeValueFieldId, models.ProductAttributeValueFieldName},
		})
		if err != nil {
			return nil, errors.Wrap(err, "valueLabels")
		}
		if !value.HasData {
			continue
		}

		name := models.NewProductAttributeValueFrom(value.Data).GetName()
		labels = append(labels, BuildDisplayName(name, nil))
	}
	return labels, nil
}

func (this *ProductTemplateDomainServiceImpl) fetchVariant(
	ctx corectx.Context, variantId string,
) (*models.ProductVariant, error) {
	variantEngine, err := engineFor(models.ProductVariantSchemaName)
	if err != nil {
		return nil, err
	}

	found, err := variantEngine.ResourceRepository().GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{models.ProductVariantFieldId: variantId},
	})
	if err != nil {
		return nil, errors.Wrap(err, "fetchVariant")
	}
	if !found.HasData {
		return nil, nil
	}
	return models.NewProductVariantFrom(found.Data), nil
}

func (this *ProductTemplateDomainServiceImpl) fetchTemplate(
	ctx corectx.Context, templateId string,
) (*models.ProductTemplate, error) {
	if templateId == "" {
		return nil, nil
	}

	templateEngine, err := engineFor(models.ProductTemplateSchemaName)
	if err != nil {
		return nil, err
	}

	found, err := templateEngine.ResourceRepository().GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{models.ProductTemplateFieldId: templateId},
	})
	if err != nil {
		return nil, errors.Wrap(err, "fetchTemplate")
	}
	if !found.HasData {
		return nil, nil
	}
	return models.NewProductTemplateFrom(found.Data), nil
}

// engineFor resolves another resource's engine from the registry. It is a variable so a test can
// substitute its own engines: the registry is a package singleton populated during Init, which a
// unit test cannot build.

func ptrOf[T any](v T) *T {
	return &v
}
