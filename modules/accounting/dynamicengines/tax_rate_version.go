package dynamicengines

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

func taxRateVersionEngineSpec() engineSpec {
	return engineSpec{
		SchemaName:    models.TaxRateVersionSchemaName,
		DefineActions: defineRateVersionActions,
	}
}

func defineRateVersionActions(engine drif.DynamicResourceEngine) error {
	err := engine.ModifyAction(drif.DynamicActionDelta{
		ActionName:    drif.ActionCreate,
		ValidateExtra: validateRateVersionCreate(engine),
	})
	if err != nil {
		return errors.Wrap(err, "failed to attach tax rate version create validation")
	}

	err = engine.ModifyAction(drif.DynamicActionDelta{
		ActionName:    drif.ActionUpdate,
		KeysToFetch:   rateVersionKeysToFetch,
		ValidateExtra: validateRateVersionUpdate(engine),
	})
	if err != nil {
		return errors.Wrap(err, "failed to attach tax rate version update validation")
	}

	err = engine.ModifyAction(drif.DynamicActionDelta{
		ActionName:    drif.ActionDelete,
		KeysToFetch:   rateVersionKeysToFetch,
		ValidateExtra: validateRateVersionDelete,
	})
	return errors.Wrap(err, "failed to attach tax rate version delete validation")
}

func rateVersionKeysToFetch(params dmodel.DynamicFields) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.TaxRateVersionFieldId: params[models.TaxRateVersionFieldId],
	}
}

func validateRateVersionCreate(engine drif.DynamicResourceEngine) drif.ActionValidateExtraFn {
	return func(
		ctx corectx.Context, inputModel *drif.DynamicEntity, _ *drif.DynamicEntity, vErrs *ft.ClientErrors,
	) error {
		params := inputModel.GetFieldData()
		version := models.NewTaxRateVersionFrom(params)
		assertWellFormedPeriod(
			version.GetEffectiveFrom(), version.GetEffectiveTo(),
			models.TaxRateVersionFieldEffectiveTo, vErrs)
		assertRateFieldsExclusive(version, vErrs)
		return assertNoPublishedRateOverlap(ctx, engine, version, nil, vErrs)
	}
}

func validateRateVersionUpdate(engine drif.DynamicResourceEngine) drif.ActionValidateExtraFn {
	return func(
		ctx corectx.Context, inputModel *drif.DynamicEntity, foundModel *drif.DynamicEntity, vErrs *ft.ClientErrors,
	) error {
		params := inputModel.GetFieldData()
		found := entityFields(foundModel)
		assertLifecycleTransition(models.TaxRateVersionFieldLifecycleStatus, params, found, vErrs)
		assertMaterialFieldsImmutable(
			models.TaxRateVersionSchemaName,
			models.TaxRateVersionFieldLifecycleStatus, params, found, vErrs)
		if vErrs.Count() > 0 {
			return nil
		}

		version := models.NewTaxRateVersionFrom(mergeFields(found, params))
		assertWellFormedPeriod(
			version.GetEffectiveFrom(), version.GetEffectiveTo(),
			models.TaxRateVersionFieldEffectiveTo, vErrs)
		assertRateFieldsExclusive(version, vErrs)

		var selfId *string
		if found != nil {
			selfId = models.NewTaxRateVersionFrom(*found).GetId()
		}
		return assertNoPublishedRateOverlap(ctx, engine, version, selfId, vErrs)
	}
}

func validateRateVersionDelete(
	_ corectx.Context, _ *drif.DynamicEntity, foundModel *drif.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	found := entityFields(foundModel)
	assertDeletableLifecycle(models.TaxRateVersionFieldLifecycleStatus, found, vErrs)
	return nil
}

// assertRateFieldsExclusive rejects a rate version that is both a percentage and a fixed amount,
// or neither.
//
// The two describe incompatible arithmetic — one multiplies a base, the other multiplies a
// quantity — so a row carrying both leaves the engine no principled way to choose, and a row
// carrying neither resolves to no charge at all while looking like a configured rate.
//
// Which of the two is required depends on the tax's calculation type, which lives on the
// definition version rather than here. Checking that pairing needs the effective definition at
// this rate's start date and belongs to the publish-time validation; what this function can settle
// on its own is that exactly one of the two shapes is present.
func assertRateFieldsExclusive(version *models.TaxRateVersion, vErrs *ft.ClientErrors) {
	hasRate := version.GetRate() != nil
	hasFixed := version.GetFixedAmount() != nil

	if hasRate && hasFixed {
		vErrs.Append(*ft.NewBusinessViolation(models.TaxRateVersionFieldFixedAmount,
			"tax.rate_and_fixed_amount_both_set",
			"a tax rate version carries either a percentage rate or a fixed amount, never both"))
		return
	}
	if !hasRate && !hasFixed {
		vErrs.Append(*ft.NewBusinessViolation(models.TaxRateVersionFieldRate,
			"tax.rate_or_fixed_amount_required",
			"a tax rate version must carry either a percentage rate or a fixed amount"))
		return
	}

	if hasFixed {
		if version.GetCurrencyCode() == nil {
			vErrs.Append(*ft.NewBusinessViolation(models.TaxRateVersionFieldCurrencyCode,
				"tax.fixed_amount_requires_currency",
				"a fixed tax amount must state the currency it is denominated in"))
		}
		if version.GetRateUomId() == nil {
			vErrs.Append(*ft.NewBusinessViolation(models.TaxRateVersionFieldRateUomId,
				"tax.fixed_amount_requires_uom",
				"a fixed tax amount must state the unit it is charged per"))
		}
	}
}

// assertNoPublishedRateOverlap is the rate-version counterpart of assertNoPublishedOverlap.
//
// This is TAX-SUP-INV-06 made enforceable: a tax_date must resolve to exactly one published rate.
// Without it the engine would face two candidates and have to pick, and BR-TAX-ESS-SUP-006
// explicitly forbids resolving that by taking the newest row — an arbitrary tie-break is how a
// customer ends up charged 10% on a document that says 8%.
func assertNoPublishedRateOverlap(
	ctx corectx.Context,
	engine drif.DynamicResourceEngine,
	version *models.TaxRateVersion,
	selfId *string,
	vErrs *ft.ClientErrors,
) error {
	status := version.GetLifecycleStatus()
	if status == nil || *status != string(models.LifecyclePublished) {
		return nil
	}
	taxId := version.GetTaxId()
	if taxId == nil {
		return nil
	}

	existing, err := models.FindPublishedRateVersionsOfTax(ctx, engine.ResourceRepository(), *taxId, 200)
	if err != nil {
		return errors.Wrap(err, "assertNoPublishedRateOverlap")
	}

	for _, item := range existing {
		other := models.NewTaxRateVersionFrom(item)
		if selfId != nil && other.GetId() != nil && *other.GetId() == *selfId {
			continue
		}
		if models.PeriodsOverlap(
			version.GetEffectiveFrom(), version.GetEffectiveTo(),
			other.GetEffectiveFrom(), other.GetEffectiveTo(),
		) {
			vErrs.Append(*ft.NewBusinessViolation(models.TaxRateVersionFieldEffectiveFrom,
				"tax.published_rate_overlap",
				"another published rate version of this tax already covers part of this period; "+
					"a tax date must resolve to exactly one rate"))
			return nil
		}
	}
	return nil
}
