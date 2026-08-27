package dynamicengines

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

func taxDefinitionVersionEngineSpec() engineSpec {
	return engineSpec{
		SchemaName:    models.TaxDefinitionVersionSchemaName,
		DefineActions: defineDefinitionVersionActions,
	}
}

func defineDefinitionVersionActions(engine drif.DynamicResourceEngine) error {
	err := engine.ModifyAction(drif.DynamicActionDelta{
		ActionName:    drif.ActionCreate,
		ValidateExtra: validateDefinitionVersionCreate(engine),
	})
	if err != nil {
		return errors.Wrap(err, "failed to attach tax definition version create validation")
	}

	err = engine.ModifyAction(drif.DynamicActionDelta{
		ActionName:    drif.ActionUpdate,
		KeysToFetch:   definitionVersionKeysToFetch,
		ValidateExtra: validateDefinitionVersionUpdate(engine),
	})
	if err != nil {
		return errors.Wrap(err, "failed to attach tax definition version update validation")
	}

	err = engine.ModifyAction(drif.DynamicActionDelta{
		ActionName:    drif.ActionDelete,
		KeysToFetch:   definitionVersionKeysToFetch,
		ValidateExtra: validateDefinitionVersionDelete,
	})
	return errors.Wrap(err, "failed to attach tax definition version delete validation")
}

func definitionVersionKeysToFetch(params dmodel.DynamicFields) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.TaxDefinitionVersionFieldId: params[models.TaxDefinitionVersionFieldId],
	}
}

func validateDefinitionVersionCreate(engine drif.DynamicResourceEngine) drif.ActionValidateExtraFn {
	return func(
		ctx corectx.Context, inputModel *drif.DynamicEntity, _ *drif.DynamicEntity, vErrs *ft.ClientErrors,
	) error {
		params := inputModel.GetFieldData()
		version := models.NewTaxDefinitionVersionFrom(params)
		assertWellFormedPeriod(
			version.GetEffectiveFrom(), version.GetEffectiveTo(),
			models.TaxDefinitionVersionFieldEffectiveTo, vErrs)
		assertTreatmentMatchesCalculationType(version, vErrs)
		return assertNoPublishedOverlap(ctx, engine, version, nil, vErrs)
	}
}

func validateDefinitionVersionUpdate(engine drif.DynamicResourceEngine) drif.ActionValidateExtraFn {
	return func(
		ctx corectx.Context, inputModel *drif.DynamicEntity, foundModel *drif.DynamicEntity, vErrs *ft.ClientErrors,
	) error {
		params := inputModel.GetFieldData()
		found := entityFields(foundModel)
		assertLifecycleTransition(
			models.TaxDefinitionVersionFieldLifecycleStatus, params, found, vErrs)
		assertMaterialFieldsImmutable(
			models.TaxDefinitionVersionSchemaName,
			models.TaxDefinitionVersionFieldLifecycleStatus, params, found, vErrs)
		if vErrs.Count() > 0 {
			return nil
		}

		merged := mergeFields(found, params)
		version := models.NewTaxDefinitionVersionFrom(merged)
		assertWellFormedPeriod(
			version.GetEffectiveFrom(), version.GetEffectiveTo(),
			models.TaxDefinitionVersionFieldEffectiveTo, vErrs)
		assertTreatmentMatchesCalculationType(version, vErrs)

		var selfId *string
		if found != nil {
			selfId = models.NewTaxDefinitionVersionFrom(*found).GetId()
		}
		return assertNoPublishedOverlap(ctx, engine, version, selfId, vErrs)
	}
}

func validateDefinitionVersionDelete(
	_ corectx.Context, _ *drif.DynamicEntity, foundModel *drif.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	found := entityFields(foundModel)
	assertDeletableLifecycle(models.TaxDefinitionVersionFieldLifecycleStatus, found, vErrs)
	return nil
}

// assertTreatmentMatchesCalculationType enforces the two pairings BR-TAX-ESS-SUP-015 fixes.
//
// The "none" calculation type exists so that a tax can carry legal meaning without producing an
// amount — an exemption still needs a code on the invoice. It is therefore valid only where the
// treatment says no tax is due for a legal reason.
//
// Zero-rated is deliberately excluded from that set even though it also yields zero. A zero-rated
// supply is taxed, at 0%, and the distinction is what lets the supplier reclaim input tax; modelling
// it as "no calculation" would erase the fact that a rate was applied at all.
func assertTreatmentMatchesCalculationType(
	version *models.TaxDefinitionVersion, vErrs *ft.ClientErrors,
) {
	rawCalc, rawTreat := version.GetCalculationType(), version.GetTaxTreatment()
	if rawCalc == nil || rawTreat == nil {
		return
	}
	calc, treat := models.WrapCalculationType(*rawCalc), models.WrapTaxTreatment(*rawTreat)
	if calc == nil || treat == nil {
		return
	}

	if *calc == models.CalculationNone {
		switch *treat {
		case models.TaxTreatmentExempt, models.TaxTreatmentNonTaxable, models.TaxTreatmentOutOfScope:
		default:
			vErrs.Append(*ft.NewBusinessViolation(
				models.TaxDefinitionVersionFieldCalculationType,
				"tax.none_calculation_requires_no_charge_treatment",
				"a calculation type of 'none' is only valid with exempt, non_taxable or out_of_scope"))
		}
	}

	if *treat == models.TaxTreatmentZeroRated && *calc == models.CalculationNone {
		vErrs.Append(*ft.NewBusinessViolation(
			models.TaxDefinitionVersionFieldCalculationType,
			"tax.zero_rated_requires_percentage",
			"a zero-rated tax is charged at 0% and must use calculation type 'percentage', not 'none'"))
	}
}

// assertNoPublishedOverlap rejects a published version whose period overlaps another published
// version of the same tax.
//
// The check runs only for published rows, because two overlapping drafts are a work in progress
// rather than an error — they become one the moment someone tries to publish the second.
func assertNoPublishedOverlap(
	ctx corectx.Context,
	engine drif.DynamicResourceEngine,
	version *models.TaxDefinitionVersion,
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

	existing, err := models.FindPublishedVersionsOfTax(ctx, engine.ResourceRepository(), *taxId, 200)
	if err != nil {
		return errors.Wrap(err, "assertNoPublishedOverlap")
	}

	for _, item := range existing {
		other := models.NewTaxDefinitionVersionFrom(item)
		if selfId != nil && other.GetId() != nil && *other.GetId() == *selfId {
			continue
		}
		if models.PeriodsOverlap(
			version.GetEffectiveFrom(), version.GetEffectiveTo(),
			other.GetEffectiveFrom(), other.GetEffectiveTo(),
		) {
			vErrs.Append(*ft.NewBusinessViolation(
				models.TaxDefinitionVersionFieldEffectiveFrom,
				"tax.published_version_overlap",
				"another published version of this tax already covers part of this period; "+
					"a tax date must resolve to exactly one definition"))
			return nil
		}
	}
	return nil
}
