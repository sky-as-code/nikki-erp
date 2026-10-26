package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// The variant rules that depend on the template's configuration: who may come back from the
// archive, and which variant may not be deleted. A template with a dynamic or never line resolves
// its products on demand, so having no active variant is a normal state for it and none of the
// "last variant" rules apply.

// assertVariantRestorable refuses a manual unarchive the configuration would not produce: the
// template is archived, or the combination is no longer one of the template's enabled
// combinations. Unarchiving must not bring back a combination the configuration removed.
func assertVariantRestorable(ctx corectx.Context, variantId string, vErrs *ft.ClientErrors) error {
	variant, template, err := loadVariantWithTemplate(ctx, variantId)
	if err != nil || variant == nil || template == nil {
		return err
	}
	if derefBool(template.IsArchived()) {
		vErrs.Append(*ft.NewBusinessViolation(basemodel.FieldIsArchived,
			ft.ErrorKey("product_variant.template_archived", "inventory"),
			"the template is archived; unarchive the template first"))
		return nil
	}
	config, err := LoadTemplateConfiguration(ctx, derefString(variant.GetProductTemplateId()))
	if err != nil {
		return err
	}
	if !config.IsCombinationValid(derefString(variant.GetCombinationKey()), variantKindOf(variant)) {
		vErrs.Append(*ft.NewBusinessViolation(basemodel.FieldIsArchived,
			ft.ErrorKey("product_variant.combination_not_valid", "inventory"),
			"this combination is no longer part of the template's attribute configuration"))
	}
	return nil
}

// assertVariantDeleteKeepsTemplateUsable refuses deleting a base that still has completes under
// it, and deleting the last active complete of a template with no on-demand line: such a template
// would have nothing left to sell, and the way to stop selling it is to archive it.
func assertVariantDeleteKeepsTemplateUsable(ctx corectx.Context, variantId string, vErrs *ft.ClientErrors) error {
	variant, err := fetchVariantRow(ctx, variantId)
	if err != nil || variant == nil {
		return err
	}
	if variantKindOf(variant) == models.VariantKindBase {
		return assertBaseHasNoCompletes(ctx, variantId, vErrs)
	}
	if derefBool(variant.IsArchived()) {
		return nil
	}
	templateId := derefString(variant.GetProductTemplateId())
	config, err := LoadTemplateConfiguration(ctx, templateId)
	if err != nil || config.HasOnDemandLine() {
		return err
	}
	remaining, err := findActiveCompleteVariants(ctx, templateId, 2)
	if err != nil {
		return err
	}
	for _, row := range remaining {
		if derefString(models.NewProductVariantFrom(row).GetId()) != variantId {
			return nil
		}
	}
	vErrs.Append(*ft.NewBusinessViolation(models.ProductVariantFieldId,
		ft.ErrorKey("product_variant.last_complete_variant", "inventory"),
		"this is the template's last active variant; archive the template instead of deleting it"))
	return nil
}

// assertBaseHasNoCompletes reports the base_variant_id foreign key as a business error before the
// database refuses the delete with a bare constraint name.
func assertBaseHasNoCompletes(ctx corectx.Context, baseId string, vErrs *ft.ClientErrors) error {
	children, err := searchVariants(ctx, 1,
		*dmodel.NewSearchNode().NewCondition(models.ProductVariantFieldBaseVariantId, dmodel.Equals, baseId))
	if err != nil {
		return err
	}
	if len(children) > 0 {
		vErrs.Append(*ft.NewBusinessViolation(models.ProductVariantFieldId,
			ft.ErrorKey("product_variant.base_has_complete_variants", "inventory"),
			"this base variant still has complete variants built on it"))
	}
	return nil
}

// findActiveCompleteVariants returns up to limit unarchived complete variants of a template. A
// base is never selectable, so it never keeps a template available.
func findActiveCompleteVariants(ctx corectx.Context, templateId string, limit int) ([]dmodel.DynamicFields, error) {
	return searchVariants(ctx, limit,
		*dmodel.NewSearchNode().NewCondition(models.ProductVariantFieldProductTemplateId, dmodel.Equals, templateId),
		*dmodel.NewSearchNode().NewCondition(basemodel.FieldIsArchived, dmodel.Equals, false),
		*dmodel.NewSearchNode().NewCondition(
			models.ProductVariantFieldVariantKind, dmodel.Equals, models.VariantKindComplete.String()),
	)
}

func searchVariants(ctx corectx.Context, limit int, conditions ...dmodel.SearchNode) ([]dmodel.DynamicFields, error) {
	repo, err := repoFor(models.ProductVariantSchemaName)
	if err != nil {
		return nil, err
	}
	graph := &dmodel.SearchGraph{}
	graph.And(conditions...)
	found, err := repo.Search(ctx, dyn.RepoSearchParam{Graph: graph, Page: 0, Size: limit})
	if err != nil {
		return nil, errors.Wrap(err, "searchVariants")
	}
	if found.ClientErrors.Count() > 0 {
		return nil, errors.Wrap(found.ClientErrors.ToError(), "searchVariants")
	}
	return found.Data.Items, nil
}

func fetchVariantRow(ctx corectx.Context, variantId string) (*models.ProductVariant, error) {
	row, err := findRecord(ctx, models.ProductVariantSchemaName, models.ProductVariantFieldId, variantId)
	if err != nil || row == nil {
		return nil, err
	}
	return models.NewProductVariantFrom(row), nil
}

// loadVariantWithTemplate reads a variant and its template; either is nil when it does not exist.
func loadVariantWithTemplate(
	ctx corectx.Context, variantId string,
) (*models.ProductVariant, *models.ProductTemplate, error) {
	variant, err := fetchVariantRow(ctx, variantId)
	if err != nil || variant == nil {
		return nil, nil, err
	}
	row, err := findRecord(ctx, models.ProductTemplateSchemaName,
		models.ProductTemplateFieldId, derefString(variant.GetProductTemplateId()))
	if err != nil || row == nil {
		return variant, nil, err
	}
	return variant, models.NewProductTemplateFrom(row), nil
}

// variantKindOf defaults to complete: the column is required with that default, so only a row
// read without the field lacks it.
func variantKindOf(variant *models.ProductVariant) models.VariantKind {
	if kind := variant.GetVariantKind(); kind != nil && *kind != "" {
		return *kind
	}
	return models.VariantKindComplete
}
