package services

import (
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

func validateDefinitionVersionCreate(repo models.TaxSearcher) composable.ValidateExtraFn {
	return func(
		ctx corectx.Context, inputModel *composable.DynamicEntity, _ *composable.DynamicEntity, vErrs *ft.ClientErrors,
	) error {
		params := inputModel.GetFieldData()
		version := models.NewTaxDefinitionVersionFrom(params)
		assertWellFormedPeriod(
			version.GetEffectiveFrom(), version.GetEffectiveTo(),
			models.TaxDefinitionVersionFieldEffectiveTo, vErrs)
		assertTreatmentMatchesCalculationType(version, vErrs)
		return assertNoPublishedOverlap(ctx, repo, version, nil, vErrs)
	}
}

func validateDefinitionVersionUpdate(repo models.TaxSearcher) composable.ValidateExtraFn {
	return func(
		ctx corectx.Context, inputModel *composable.DynamicEntity, foundModel *composable.DynamicEntity,
		vErrs *ft.ClientErrors,
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
		return assertNoPublishedOverlap(ctx, repo, version, selfId, vErrs)
	}
}

func validateDefinitionVersionDelete(
	_ corectx.Context, _ *composable.DynamicEntity, foundModel *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	found := entityFields(foundModel)
	assertDeletableLifecycle(models.TaxDefinitionVersionFieldLifecycleStatus, found, vErrs)
	return nil
}

// assertTreatmentMatchesCalculationType pairs the treatment with the calculation type.
//
// The "none" calculation type lets a tax carry legal meaning without producing an amount, so it is
// valid only where the treatment says no tax is due for a legal reason. Zero-rated is excluded even
// though it also yields zero: a zero-rated supply is taxed at 0%, and that distinction is what lets
// the supplier reclaim input tax.
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
				ft.ErrorKey("tax.none_calculation_requires_no_charge_treatment", "accounting"),
				"a calculation type of 'none' is only valid with exempt, non_taxable or out_of_scope"))
		}
	}

	if *treat == models.TaxTreatmentZeroRated && *calc == models.CalculationNone {
		vErrs.Append(*ft.NewBusinessViolation(
			models.TaxDefinitionVersionFieldCalculationType,
			ft.ErrorKey("tax.zero_rated_requires_percentage", "accounting"),
			"a zero-rated tax is charged at 0% and must use calculation type 'percentage', not 'none'"))
	}
}

// assertNoPublishedOverlap rejects a published version whose period overlaps another published
// version of the same tax. It runs only for published rows; two overlapping drafts are a work in
// progress until the second is published.
func assertNoPublishedOverlap(
	ctx corectx.Context,
	repo models.TaxSearcher,
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

	existing, err := models.FindPublishedVersionsOfTax(ctx, repo, *taxId, 200)
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
				ft.ErrorKey("tax.published_version_overlap", "accounting"),
				"another published version of this tax already covers part of this period; "+
					"a tax date must resolve to exactly one definition"))
			return nil
		}
	}
	return nil
}
