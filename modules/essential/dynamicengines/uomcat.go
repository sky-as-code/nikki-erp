package dynamicengines

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
)


func uomCatEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.UomCatSchemaName,
		DefaultFields: []string{
			models.UomCatFieldName,
			models.UomCatFieldReferenceUomId,
		},
		DefineActions: defineUomCatActions,
	}
}

func defineUomCatActions(engine drif.DynamicResourceEngine) error {
	err := engine.ModifyAction(drif.DynamicActionDelta{
		ActionName:    drif.ActionCreate,
		ValidateExtra: validateUomCatReference(nil),
	})
	if err != nil {
		return errors.Wrap(err, "failed to attach uomcat create validation")
	}

	err = engine.ModifyAction(drif.DynamicActionDelta{
		ActionName:  drif.ActionUpdate,
		KeysToFetch: uomCatKeysToFetch,
		ValidateExtra: validateUomCatReference(func(
			ctx corectx.Context, params dmodel.DynamicFields, found *models.UomCat, vErrs *ft.ClientErrors,
		) {
			assertReferenceUomStableWhileInUse(ctx, params, found, vErrs)
		}),
	})
	return errors.Wrap(err, "failed to attach uomcat update validation")
}

func uomCatKeysToFetch(params dmodel.DynamicFields) dmodel.DynamicFields {
	return dmodel.DynamicFields{models.UomCatFieldId: params[models.UomCatFieldId]}
}

type uomCatUpdateCheckFn func(
	ctx corectx.Context, params dmodel.DynamicFields, found *models.UomCat, vErrs *ft.ClientErrors,
)

// validateUomCatReference enforces BR-UOM-ESS-004 and UOM-ESS-INV-03: a category's Reference
// UoM must be a UoM of that same category. onUpdate carries the checks that only make sense
// once a stored record exists.
func validateUomCatReference(onUpdate uomCatUpdateCheckFn) drif.ActionValidateExtraFn {
	return func(
		ctx corectx.Context, params dmodel.DynamicFields, foundModel *dmodel.DynamicFields, vErrs *ft.ClientErrors,
	) error {
		if onUpdate != nil && foundModel != nil {
			onUpdate(ctx, params, models.NewUomCatFrom(*foundModel), vErrs)
		}

		cat := models.NewUomCatFrom(params)
		referenceUomId := cat.GetReferenceUomId()
		if referenceUomId == nil {
			// A category may exist without a reference UoM until one is created for it.
			return nil
		}

		categoryId := cat.GetId()
		if categoryId == nil && foundModel != nil {
			categoryId = models.NewUomCatFrom(*foundModel).GetId()
		}
		if categoryId == nil {
			// Create: the category has no id yet, so no existing UoM can already belong to
			// it. The reference must still exist (BR-UOM-ESS-004 needs something to point
			// at), but any UoM that does exist necessarily belongs to another category.
			return assertReferenceUomBelongsToCategory(ctx, *referenceUomId, "", vErrs)
		}
		return assertReferenceUomBelongsToCategory(ctx, *referenceUomId, *categoryId, vErrs)
	}
}

// assertReferenceUomBelongsToCategory reports the reference UoM as not-found when it does
// not exist, and as foreign when it belongs to a category other than categoryId. An empty
// categoryId is the create case: the category has no id yet, so every existing UoM is foreign.
func assertReferenceUomBelongsToCategory(
	ctx corectx.Context, referenceUomId string, categoryId string, vErrs *ft.ClientErrors,
) error {
	uomEngine, ok := dynamicresource.Registry().GetEngine(models.UomSchemaName)
	if !ok {
		return errors.Errorf("assertReferenceUomBelongsToCategory: the '%s' engine is not registered",
			models.UomSchemaName)
	}

	found, err := uomEngine.ResourceRepository().GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{models.UomFieldId: referenceUomId},
	})
	if err != nil {
		return errors.Wrap(err, "assertReferenceUomBelongsToCategory")
	}
	if !found.HasData {
		vErrs.Append(*ft.NewBusinessViolation(models.UomCatFieldReferenceUomId, "uomcat.reference_uom_not_found",
			"the referenced UoM does not exist"))
		return nil
	}

	uom := models.NewUomFrom(found.Data)
	if uom.GetCategoryId() == nil || *uom.GetCategoryId() != categoryId {
		vErrs.Append(*ft.NewBusinessViolation(models.UomCatFieldReferenceUomId, "uomcat.reference_uom_foreign",
			"the Reference UoM must belong to this UoM Category"))
	}
	return nil
}

// assertReferenceUomStableWhileInUse enforces BR-UOM-ESS-021: once a category is in use,
// repointing its Reference UoM would reinterpret every factor in the category.
func assertReferenceUomStableWhileInUse(
	ctx corectx.Context, params dmodel.DynamicFields, found *models.UomCat, vErrs *ft.ClientErrors,
) {
	submitted, isSubmitted := params[models.UomCatFieldReferenceUomId]
	if !isSubmitted || !isUomCatInUse(ctx, found) {
		return
	}
	// Re-submitting the value unchanged is not a change.
	current := found.GetReferenceUomId()
	if current != nil && submitted == *current {
		return
	}
	vErrs.Append(*ft.NewBusinessViolation(models.UomCatFieldReferenceUomId, "uomcat.reference_uom_immutable_while_in_use",
		"this UoM Category is already in use; its Reference UoM can no longer be changed"))
}

// isUomCatInUse reports whether any product or transaction depends on this category.
//
// TODO: shares the fate of isUomInUse — no module consumes UoM yet. Wire a real probe when
// stock, purchase or sales land, or BR-UOM-ESS-021 stays structurally declared but inert.
func isUomCatInUse(_ corectx.Context, _ *models.UomCat) bool {
	return false
}
