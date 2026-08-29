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

// Turning an attribute selection into a concrete variant, and bringing a template's variants in
// step with its attribute configuration. Methods on the same type as the service core.

func (this *ProductTemplateDomainServiceImpl) ResolveProductSelection(
	ctx corectx.Context, query itProduct.ResolveProductSelectionQuery,
) (*itProduct.ResolveProductSelectionResult, error) {
	if query.TemplateId == "" {
		return &itProduct.ResolveProductSelectionResult{}, nil
	}

	// The template is read first, so resolving against one that does not exist is a missing record
	// rather than a 200 carrying the combination key of nothing.
	template, err := this.fetchTemplate(ctx, query.TemplateId)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return &itProduct.ResolveProductSelectionResult{}, nil
	}

	// HasData means the template resolved, not that a variant exists for the combination: "no variant
	// yet" is a real answer carrying the combination key, and reporting it as missing would be a 404.
	combinationKey := ResolveProductSelection(query.Selections)
	result := &itProduct.ResolveProductSelectionResult{
		Data:    itProduct.ResolveProductSelectionResultData{CombinationKey: combinationKey},
		HasData: true,
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
		return result, nil
	}

	if !query.MaterializeIfMissing {
		return result, nil
	}

	// The composite unique on (product_template_id, combination_key) makes this safe under
	// concurrency: the loser of a race sees the constraint rather than creating a duplicate.
	variantId, err := this.materializeVariant(ctx, variantEngine, query.TemplateId, combinationKey)
	if err != nil {
		return nil, err
	}
	result.Data.VariantId = variantId
	result.Data.Materialized = variantId != ""
	return result, nil
}

// GenerateVariants brings a template's variants in step with its INSTANT attribute configuration.
func (this *ProductTemplateDomainServiceImpl) GenerateVariants(
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

	plan := PlanVariantSync(wanted, existingKeys)
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
func (this *ProductTemplateDomainServiceImpl) wantedCombinations(
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
			// The variant references the template-scoped value, so that is the id the combination key
			// must carry.
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

	return BuildInstantCombinations(options), nil
}

func (this *ProductTemplateDomainServiceImpl) attributeMode(
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
func (this *ProductTemplateDomainServiceImpl) materializeVariant(
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

	// Created through the engine's resource service, not its repository: the create pipeline applies
	// the schema defaults (id, created_at, etag, is_archived) and the audit fields, and a raw Insert
	// would reach the database with a null primary key.
	created, err := variantEngine.ResourceService().Create(ctx, variant.GetFieldData())
	if err != nil {
		return "", errors.Wrap(err, "materializeVariant")
	}
	if created.ClientErrors.Count() > 0 {
		return "", errors.Wrap(created.ClientErrors.ToError(), "materializeVariant")
	}
	if !created.HasData {
		return "", nil
	}
	return derefString(models.NewProductVariantFrom(created.Data).GetId()), nil
}

// flatten builds the effective product of a variant, reading its template and the labels of its
// attribute values.
