package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// validateVariantCreate enforces the invariants that apply to a brand-new variant.
func validateVariantCreate(repo models.ProductSearcher) composable.ValidateExtraFn {
	return func(
		ctx corectx.Context, inputModel *composable.DynamicEntity, _ *composable.DynamicEntity, vErrs *ft.ClientErrors,
	) error {
		variant := models.NewProductVariantFrom(inputModel.GetFieldData())
		return checkUniqueCombination(ctx, repo, variant, "", vErrs)
	}
}

// validateVariantUpdate re-checks uniqueness when the combination changes. An update is partial,
// so the submitted fields are overlaid onto the stored record and the result validated, rather
// than only what was sent.
func validateVariantUpdate(repo models.ProductSearcher) composable.ValidateExtraFn {
	return func(
		ctx corectx.Context, inputModel *composable.DynamicEntity, foundModel *composable.DynamicEntity, vErrs *ft.ClientErrors,
	) error {
		if foundModel == nil {
			return nil
		}
		submitted := models.NewProductVariantFrom(inputModel.GetFieldData())
		stored := models.NewProductVariantFrom(foundModel.GetFieldData())
		assertNotReparented(submitted, stored, vErrs)
		if vErrs.Count() > 0 {
			return nil
		}
		merged := mergeVariantForValidation(submitted, stored)
		return checkUniqueCombination(ctx, repo, merged, derefString(stored.GetId()), vErrs)
	}
}

// assertNotReparented refuses an update that moves a variant to another template.
// product_template_id is declared no_update and the schema validator honours that by dropping the
// field, answering 200 as though the change had applied. Re-parenting would silently reinterpret
// every transaction already referencing the variant, so it is reported rather than ignored.
func assertNotReparented(submitted *models.ProductVariant, stored *models.ProductVariant, vErrs *ft.ClientErrors) {
	newTemplateId := derefString(submitted.GetProductTemplateId())
	if newTemplateId == "" || newTemplateId == derefString(stored.GetProductTemplateId()) {
		return
	}
	vErrs.Append(*ft.NewBusinessViolation(models.ProductVariantFieldProductTemplateId,
		"product_variant.immutable_template",
		"product_template_id cannot be changed; create a variant under the other template instead"))
}

// mergeVariantForValidation overlays the submitted fields onto the stored record, so that a
// partial update is checked against the record it will produce.
func mergeVariantForValidation(submitted *models.ProductVariant, stored *models.ProductVariant) *models.ProductVariant {
	merged := dmodel.DynamicFields{}
	for key, val := range stored.GetFieldData() {
		merged[key] = val
	}
	for key, val := range submitted.GetFieldData() {
		merged[key] = val
	}
	return models.NewProductVariantFrom(merged)
}

// checkUniqueCombination enforces that one template never holds two variants with the same
// attribute combination; selfId excludes the record being updated. The database carries the same
// composite unique, so this exists to turn a constraint violation into a field-level business
// error the UI can show rather than a 500.
func checkUniqueCombination(
	ctx corectx.Context,
	repo models.ProductSearcher,
	variant *models.ProductVariant,
	selfId string,
	vErrs *ft.ClientErrors,
) error {
	templateId := derefString(variant.GetProductTemplateId())
	combinationKey := variant.GetCombinationKey()
	if templateId == "" || combinationKey == nil {
		// Absence is the schema's business, not ours.
		return nil
	}

	// Size 2: the record being updated may itself hold this combination, and it must not be
	// mistaken for a conflicting one.
	existing, err := models.FindVariantsByCombination(ctx, repo, templateId, *combinationKey, 2)
	if err != nil {
		return errors.Wrap(err, "checkUniqueCombination")
	}

	for _, item := range existing {
		other := models.NewProductVariantFrom(item)
		if selfId != "" && derefString(other.GetId()) == selfId {
			continue
		}
		vErrs.Append(*ft.NewBusinessViolation(models.ProductVariantFieldCombinationKey,
			"product_variant.duplicate_combination",
			"this template already has a variant with the same attribute combination"))
		return nil
	}
	return nil
}
