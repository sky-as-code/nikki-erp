package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// checkCreateCombination runs before schema validation on purpose: the schema's composite unique
// on (product_template_id, variant_kind, combination_key) is checked right after it, and would report the clash
// first as a generic error naming no field. This reports it on combination_key with its own key.
// A value of the wrong type is left to schema validation, which reports it properly.
func checkCreateCombination(repo models.ProductSearcher) composable.BeforeValidationFn {
	return func(
		ctx corectx.Context, entity *composable.DynamicEntity, vErrs *ft.ClientErrors,
	) (*composable.DynamicEntity, error) {
		data := entity.GetFieldData()
		templateId, templateOk := data[models.ProductVariantFieldProductTemplateId].(string)
		combinationKey, keyOk := data[models.ProductVariantFieldCombinationKey].(string)
		if !templateOk || !keyOk {
			return entity, nil
		}
		variant := models.NewProductVariant()
		variant.SetProductTemplateId(ptrOf(model.Id(templateId)))
		variant.SetCombinationKey(&combinationKey)
		if kind, ok := data[models.ProductVariantFieldVariantKind].(string); ok {
			variant.SetVariantKind(models.WrapVariantKind(kind))
		}
		return entity, checkUniqueCombination(ctx, repo, variant, "", vErrs)
	}
}

// captureSubmittedTemplateId records the product_template_id a client sent before the schema
// validator drops it: it is declared no_update, so by ValidateExtra the field is gone and a
// re-parenting attempt would otherwise be answered 200 as though it had applied.
func captureSubmittedTemplateId(submitted *string) composable.BeforeValidationFn {
	return func(
		_ corectx.Context, entity *composable.DynamicEntity, _ *ft.ClientErrors,
	) (*composable.DynamicEntity, error) {
		if value, ok := entity.GetFieldData()[models.ProductVariantFieldProductTemplateId].(string); ok {
			*submitted = value
		}
		return entity, nil
	}
}

// validateVariantUpdate re-checks uniqueness when the combination changes. An update is partial,
// so the submitted fields are overlaid onto the stored record and the result validated, rather
// than only what was sent. submittedTemplateId is filled by captureSubmittedTemplateId.
func validateVariantUpdate(repo models.ProductSearcher, submittedTemplateId *string) composable.ValidateExtraFn {
	return func(
		ctx corectx.Context, inputModel *composable.DynamicEntity, foundModel *composable.DynamicEntity, vErrs *ft.ClientErrors,
	) error {
		if foundModel == nil {
			return nil
		}
		submitted := models.NewProductVariantFrom(inputModel.GetFieldData())
		stored := models.NewProductVariantFrom(foundModel.GetFieldData())
		assertNotReparented(*submittedTemplateId, stored, vErrs)
		if vErrs.Count() > 0 {
			return nil
		}
		merged := mergeVariantForValidation(submitted, stored)
		return checkUniqueCombination(ctx, repo, merged, derefString(stored.GetId()), vErrs)
	}
}

// assertNotReparented refuses an update that moves a variant to another template. Re-parenting
// would silently reinterpret every transaction already referencing the variant, so it is reported
// rather than ignored.
func assertNotReparented(newTemplateId string, stored *models.ProductVariant, vErrs *ft.ClientErrors) {
	if newTemplateId == "" || newTemplateId == derefString(stored.GetProductTemplateId()) {
		return
	}
	vErrs.Append(*ft.NewBusinessViolation(models.ProductVariantFieldProductTemplateId,
		ft.ErrorKey("product_variant.immutable_template", "inventory"),
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

	// The database unique is (template, variant_kind, key): a base and a complete may share a key,
	// as when the last dynamic line goes and the complete takes its old base's key (ADD27). Size 3:
	// one record of each kind plus the record being updated, which is no conflict with itself.
	kind := variantKindOrComplete(variant)
	existing, err := models.FindVariantsByCombination(ctx, repo, templateId, *combinationKey, 3)
	if err != nil {
		return errors.Wrap(err, "checkUniqueCombination")
	}

	for _, item := range existing {
		other := models.NewProductVariantFrom(item)
		if selfId != "" && derefString(other.GetId()) == selfId {
			continue
		}
		if variantKindOrComplete(other) != kind {
			continue
		}
		vErrs.Append(*ft.NewBusinessViolation(models.ProductVariantFieldCombinationKey,
			ft.ErrorKey("product_variant.duplicate_combination", "inventory"),
			"this template already has a variant with the same attribute combination"))
		return nil
	}
	return nil
}

// variantKindOrComplete reads the kind, defaulting to complete as the schema does.
func variantKindOrComplete(variant *models.ProductVariant) models.VariantKind {
	if kind := variant.GetVariantKind(); kind != nil && *kind != "" {
		return *kind
	}
	return models.VariantKindComplete
}
