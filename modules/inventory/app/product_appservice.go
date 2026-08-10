package app

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/services"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// maxTemplateVariants bounds the variant set one call will read or generate. A template beyond
// this is past what a synchronous request should carry.
const maxTemplateVariants = 1000

// NewProductAppService derives the Products service from the engine's default one.
//
// base is the Product Template engine's own resource service, which this type embeds: every
// built-in CRUD action keeps running through the default implementation, and the capabilities
// below are added on top. The result is installed with Engine.SetResourceService.
func NewProductAppService(base drif.DynamicResourceService) itProduct.ProductService {
	return &ProductAppServiceImpl{DynamicResourceService: base}
}

// ProductAppServiceImpl serves the Products capabilities the dynamic resource engine cannot
// express: flattening a template and variant into one effective product, and turning an
// attribute selection into a concrete variant.
//
// It resolves other resources' engines at call time rather than holding them, because a
// capability spanning several schemas needs engines the embedded service does not carry, and
// engine creation and this service's construction both happen during Init.
type ProductAppServiceImpl struct {
	drif.DynamicResourceService
}

// The derived service must satisfy both contracts: the engine installs it as its resource
// service, and custom actions type-assert it back to the Products capability. Losing either
// would only show up as a failed assertion at request time.
var (
	_ drif.DynamicResourceService = (*ProductAppServiceImpl)(nil)
	_ itProduct.ProductService    = (*ProductAppServiceImpl)(nil)
)

func (this *ProductAppServiceImpl) GetEffectiveProduct(
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

func (this *ProductAppServiceImpl) GetEffectiveProducts(
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
// line must reference. See BR §8.3 and §14.4.
func (this *ProductAppServiceImpl) ResolveProductSelection(
	ctx corectx.Context, query itProduct.ResolveProductSelectionQuery,
) (*itProduct.ResolveProductSelectionResult, error) {
	if query.TemplateId == "" {
		return &itProduct.ResolveProductSelectionResult{}, nil
	}

	combinationKey := services.ResolveProductSelection(query.Selections)
	result := &itProduct.ResolveProductSelectionResult{
		Data: itProduct.ResolveProductSelectionResultData{CombinationKey: combinationKey},
	}

	variantEngine, err := engineFor(models.ProductVariantSchemaName)
	if err != nil {
		return nil, err
	}

	existing, err := models.FindVariantsByCombination(
		ctx, variantEngine.ResourceRepository(), query.TemplateId, combinationKey, 1)
	if err != nil {
		return nil, errors.Wrap(err, "ResolveProductSelection")
	}
	if len(existing) > 0 {
		variant := models.NewProductVariantFrom(existing[0])
		result.Data.VariantId = derefString(variant.GetId())
		result.HasData = true
		return result, nil
	}

	if !query.MaterializeIfMissing {
		return result, nil
	}

	// The composite unique on (product_template_id, combination_key) is what makes this safe
	// under concurrency: two requests racing to materialize the same combination cannot both
	// win, and the loser sees the constraint rather than creating a duplicate.
	variantId, err := this.materializeVariant(ctx, variantEngine, query.TemplateId, combinationKey)
	if err != nil {
		return nil, err
	}
	result.Data.VariantId = variantId
	result.Data.Materialized = true
	result.HasData = variantId != ""
	return result, nil
}

// GenerateVariants brings a template's variants in step with its INSTANT attribute
// configuration. See BR §8.2 and §8.5.
func (this *ProductAppServiceImpl) GenerateVariants(
	ctx corectx.Context, query itProduct.GenerateVariantsQuery,
) (*itProduct.GenerateVariantsResult, error) {
	if query.TemplateId == "" {
		return &itProduct.GenerateVariantsResult{}, nil
	}

	wanted, err := this.wantedCombinations(ctx, query.TemplateId)
	if err != nil {
		return nil, err
	}

	variantEngine, err := engineFor(models.ProductVariantSchemaName)
	if err != nil {
		return nil, err
	}
	existingRows, err := models.FindActiveTemplateVariants(
		ctx, variantEngine.ResourceRepository(), query.TemplateId, maxTemplateVariants)
	if err != nil {
		return nil, errors.Wrap(err, "GenerateVariants")
	}

	existingKeys := make([]string, 0, len(existingRows))
	keyToId := map[string]string{}
	for _, row := range existingRows {
		variant := models.NewProductVariantFrom(row)
		key := ""
		if stored := variant.GetCombinationKey(); stored != nil {
			key = *stored
		}
		existingKeys = append(existingKeys, key)
		keyToId[key] = derefString(variant.GetId())
	}

	plan := services.PlanVariantSync(wanted, existingKeys)
	result := &itProduct.GenerateVariantsResult{
		Data:    itProduct.GenerateVariantsResultData{UnchangedCount: len(plan.Unchanged)},
		HasData: true,
	}

	for _, key := range plan.ToCreate {
		variantId, err := this.materializeVariant(ctx, variantEngine, query.TemplateId, key)
		if err != nil {
			return nil, err
		}
		if variantId != "" {
			result.Data.CreatedVariantIds = append(result.Data.CreatedVariantIds, variantId)
		}
	}
	for _, key := range plan.Obsolete {
		if variantId := keyToId[key]; variantId != "" {
			result.Data.ObsoleteVariantIds = append(result.Data.ObsoleteVariantIds, variantId)
		}
	}
	return result, nil
}

// wantedCombinations reads a template's attribute configuration and expands it into the
// combinations it implies.
func (this *ProductAppServiceImpl) wantedCombinations(
	ctx corectx.Context, templateId string,
) ([]string, error) {
	templateAttrEngine, err := engineFor(models.ProductTemplateAttributeSchemaName)
	if err != nil {
		return nil, err
	}
	valueEngine, err := engineFor(models.ProductTemplateAttributeValueSchemaName)
	if err != nil {
		return nil, err
	}
	attributeEngine, err := engineFor(models.ProductAttributeSchemaName)
	if err != nil {
		return nil, err
	}

	templateAttrs, err := models.FindTemplateAttributes(
		ctx, templateAttrEngine.ResourceRepository(), templateId, maxTemplateVariants)
	if err != nil {
		return nil, errors.Wrap(err, "wantedCombinations")
	}

	options := make([]itProduct.AttributeOptions, 0, len(templateAttrs))
	for _, row := range templateAttrs {
		templateAttr := models.NewProductTemplateAttributeFrom(row)
		attributeId := derefString(templateAttr.GetAttributeId())
		if attributeId == "" {
			continue
		}

		mode, err := this.attributeMode(ctx, attributeEngine, attributeId)
		if err != nil {
			return nil, err
		}

		valueRows, err := models.FindTemplateAttributeValues(
			ctx, valueEngine.ResourceRepository(), derefString(templateAttr.GetId()), maxTemplateVariants)
		if err != nil {
			return nil, errors.Wrap(err, "wantedCombinations")
		}

		valueIds := make([]string, 0, len(valueRows))
		for _, valueRow := range valueRows {
			value := models.NewProductTemplateAttributeValueFrom(valueRow)
			// The variant references the template-scoped value, so that is the id the
			// combination key must carry. See BR §6.7.
			if id := derefString(value.GetId()); id != "" {
				valueIds = append(valueIds, id)
			}
		}

		options = append(options, itProduct.AttributeOptions{
			AttributeId: attributeId,
			Mode:        mode,
			ValueIds:    valueIds,
		})
	}

	return services.BuildInstantCombinations(options), nil
}

func (this *ProductAppServiceImpl) attributeMode(
	ctx corectx.Context, attributeEngine drif.DynamicResourceEngine, attributeId string,
) (models.VariantCreationMode, error) {
	found, err := attributeEngine.ResourceRepository().GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{models.ProductAttributeFieldId: attributeId},
		Fields: []string{models.ProductAttributeFieldId, models.ProductAttributeFieldVariantCreationMode},
	})
	if err != nil {
		return "", errors.Wrap(err, "attributeMode")
	}
	if !found.HasData {
		return "", nil
	}
	if mode := models.NewProductAttributeFrom(found.Data).GetVariantCreationMode(); mode != nil {
		return *mode, nil
	}
	return "", nil
}

// materializeVariant creates the variant holding one combination of a template.
func (this *ProductAppServiceImpl) materializeVariant(
	ctx corectx.Context, variantEngine drif.DynamicResourceEngine, templateId string, combinationKey string,
) (string, error) {
	template, err := this.fetchTemplate(ctx, templateId)
	if err != nil {
		return "", err
	}
	if template == nil {
		return "", nil
	}

	variant := models.NewProductVariant()
	variant.SetProductTemplateId(&templateId)
	variant.SetCombinationKey(&combinationKey)
	variant.SetStatus(ptrOf(models.ProductVariantStatusActive))
	variant.SetIsMaterialized(ptrOf(true))
	if orgId := template.GetOrgId(); orgId != nil {
		variant.SetOrgId(orgId)
	}

	created, err := variantEngine.ResourceRepository().Insert(ctx, variant.GetFieldData())
	if err != nil {
		return "", errors.Wrap(err, "materializeVariant")
	}
	if created.ClientErrors.Count() > 0 {
		return "", errors.Wrap(created.ClientErrors.ToError(), "materializeVariant")
	}
	return derefString(variant.GetId()), nil
}

// flatten builds the effective product of a variant, reading its template and the labels of its
// attribute values.
func (this *ProductAppServiceImpl) flatten(
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

	effective := services.BuildEffectiveProduct(template, variant, labels)
	return &effective, nil
}

// valueLabels resolves the display labels of a variant's attribute values, which is what turns
// the template name into "Classic T-Shirt / Black / M".
func (this *ProductAppServiceImpl) valueLabels(
	ctx corectx.Context, variant *models.ProductVariant,
) ([]string, error) {
	templateValueIds := services.VariantValueIds(variant)
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
		labels = append(labels, services.BuildDisplayName(name, nil))
	}
	return labels, nil
}

func (this *ProductAppServiceImpl) fetchVariant(
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

func (this *ProductAppServiceImpl) fetchTemplate(
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

// engineFor resolves another resource's engine from the registry.
//
// It is a variable rather than a plain function so that a test can supply its own engines: the
// registry is a package singleton populated during Init, which a unit test has no way to build.
var engineFor = func(schemaName string) (drif.DynamicResourceEngine, error) {
	engine, ok := dynamicresource.Registry().GetEngine(schemaName)
	if !ok {
		return nil, errors.Errorf("no resource engine for '%s'", schemaName)
	}
	return engine, nil
}

func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func ptrOf[T any](v T) *T {
	return &v
}
