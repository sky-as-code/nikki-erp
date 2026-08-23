package dynamicengines

import (
	"go.bryk.io/pkg/errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

func productVariantEngineSpec() engineSpec {
	return engineSpec{
		SchemaName:    models.ProductVariantSchemaName,
		DefineActions: defineProductVariantActions,
	}
}

func defineProductVariantActions(engine drif.DynamicResourceEngine) error {
	err := engine.ModifyAction(drif.DynamicActionDelta{
		ActionName:    drif.ActionCreate,
		ValidateExtra: validateVariantCreate(engine),
	})
	if err != nil {
		return errors.Wrap(err, "failed to attach product variant create validation")
	}

	err = engine.ModifyAction(drif.DynamicActionDelta{
		ActionName:    drif.ActionUpdate,
		KeysToFetch:   variantKeysToFetch,
		ValidateExtra: validateVariantUpdate(engine),
	})
	if err != nil {
		return errors.Wrap(err, "failed to attach product variant update validation")
	}

	// BR-PROD-VAR-006/007: the template follows its variants' availability. The cascade lives in
	// ProductVariantDomainServiceImpl.SetArchived rather than an action hook, because it must run
	// after the variant row is written — AfterValidationSuccess runs before MainProcess.

	// AC-PROD-032: consumers read the flattened product rather than re-deriving which fields
	// come from the template and which from the variant.
	err = engine.DefineAction(drif.DynamicActionDefinition{
		ActionName:  "get_effective",
		ActionType:  drif.ActionTypeRead,
		RestPath:    ":id/effective",
		Permission:  drif.PermissionRead,
		KeysToFetch: variantKeysToFetch,
		MainProcess: processGetEffectiveProduct,
	})
	return errors.Wrap(err, "failed to define get_effective")
}

// processGetEffectiveProduct flattens a variant together with its template.
//
// Unlike the template engine's actions it cannot read the capability off input.ResourceService:
// the derived service is installed on the *template* engine, so this engine's service is the
// plain default. It resolves the capability from the container instead. This is the one place
// where "one derived service per engine" does not fit a capability spanning two resources.
func processGetEffectiveProduct(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	// The pipeline already fetched the variant; its absence is a missing record.
	if input.FoundModel == nil {
		return &drif.ActionResult{HasData: false}, nil
	}
	variantId := derefId(models.NewProductVariantFrom(*input.FoundModel).GetId())

	var productSvc itProduct.ProductService
	if err := deps.Invoke(func(svc itProduct.ProductService) { productSvc = svc }); err != nil {
		return nil, errors.Wrap(err, "processGetEffectiveProduct")
	}

	result, err := productSvc.GetEffectiveProduct(ctx, itProduct.GetEffectiveProductQuery{
		VariantId: variantId,
	})
	if err != nil {
		return nil, errors.Wrap(err, "processGetEffectiveProduct")
	}
	if result.ClientErrors.Count() > 0 {
		return &drif.ActionResult{ClientErrors: result.ClientErrors}, nil
	}
	if !result.HasData {
		return &drif.ActionResult{HasData: false}, nil
	}

	return &drif.ActionResult{
		Data:    itProduct.NewEffectiveProductView(result.Data.Product),
		HasData: true,
	}, nil
}

func variantKeysToFetch(params dmodel.DynamicFields) dmodel.DynamicFields {
	return dmodel.DynamicFields{models.ProductVariantFieldId: params[models.ProductVariantFieldId]}
}

// validateVariantCreate enforces the invariants that apply to a brand-new variant.
func validateVariantCreate(engine drif.DynamicResourceEngine) drif.ActionValidateExtraFn {
	return func(
		ctx corectx.Context, params dmodel.DynamicFields, _ *dmodel.DynamicFields, vErrs *ft.ClientErrors,
	) error {
		variant := models.NewProductVariantFrom(params)
		return assertUniqueCombination(ctx, engine, variant, "", vErrs)
	}
}

// validateVariantUpdate re-checks uniqueness when the combination changes.
//
// An update is partial, so the submitted fields are overlaid onto the stored record and the
// result is validated, rather than only what was sent.
func validateVariantUpdate(engine drif.DynamicResourceEngine) drif.ActionValidateExtraFn {
	return func(
		ctx corectx.Context, params dmodel.DynamicFields, foundModel *dmodel.DynamicFields, vErrs *ft.ClientErrors,
	) error {
		if foundModel == nil {
			return nil
		}
		submitted, stored := models.NewProductVariantFrom(params), models.NewProductVariantFrom(*foundModel)
		assertNotReparented(submitted, stored, vErrs)
		if vErrs.Count() > 0 {
			return nil
		}
		merged := mergeVariantForValidation(submitted, stored)
		return assertUniqueCombination(ctx, engine, merged, derefId(stored.GetId()), vErrs)
	}
}

// assertNotReparented refuses an update that moves a variant to another template.
//
// product_template_id is declared no_update, and the schema validator honours that by dropping the
// field from the write — which answers 200 as though the change had been applied. Re-parenting
// would silently reinterpret every transaction already referencing the variant, so it is reported
// rather than ignored. See BR §6.1.
func assertNotReparented(submitted *models.ProductVariant, stored *models.ProductVariant, vErrs *ft.ClientErrors) {
	newTemplateId := derefId(submitted.GetProductTemplateId())
	if newTemplateId == "" || newTemplateId == derefId(stored.GetProductTemplateId()) {
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

// assertUniqueCombination enforces BR-PROD-VAR-002 and AC-PROD-012: one template never holds two
// variants with the same attribute combination. selfId excludes the record being updated.
//
// The database carries the same composite unique, so this exists to turn a constraint violation
// into a field-level business error the UI can show, rather than a 500.
func assertUniqueCombination(
	ctx corectx.Context,
	engine drif.DynamicResourceEngine,
	variant *models.ProductVariant,
	selfId string,
	vErrs *ft.ClientErrors,
) error {
	return checkUniqueCombination(ctx, engine.ResourceRepository(), variant, selfId, vErrs)
}

// checkUniqueCombination is assertUniqueCombination over the repository slice it actually needs,
// so the rule can be exercised without building an engine.
func checkUniqueCombination(
	ctx corectx.Context,
	repo models.ProductSearcher,
	variant *models.ProductVariant,
	selfId string,
	vErrs *ft.ClientErrors,
) error {
	templateId := derefId(variant.GetProductTemplateId())
	combinationKey := variant.GetCombinationKey()
	if templateId == "" || combinationKey == nil {
		// Absence is the schema's business, not ours.
		return nil
	}

	// Size 2: the record being updated may itself hold this combination, and it must not be
	// mistaken for a conflicting one.
	existing, err := models.FindVariantsByCombination(
		ctx, repo, templateId, *combinationKey, 2)
	if err != nil {
		return errors.Wrap(err, "assertUniqueCombination")
	}

	for _, item := range existing {
		other := models.NewProductVariantFrom(item)
		if selfId != "" && derefId(other.GetId()) == selfId {
			continue
		}
		vErrs.Append(*ft.NewBusinessViolation(models.ProductVariantFieldCombinationKey,
			"product_variant.duplicate_combination",
			"this template already has a variant with the same attribute combination"))
		return nil
	}
	return nil
}
