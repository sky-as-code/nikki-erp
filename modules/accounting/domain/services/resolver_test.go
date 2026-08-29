package services

import (
	"testing"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
)

// Test configurations are built as raw rows rather than through the model builders, because most of
// these cases are configurations the builders would refuse: two published versions covering one
// date, a group with no components, a component cycle. The resolver must survive them anyway.

func taxRow(id, code string) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.TaxFieldId:   id,
		models.TaxFieldCode: code,
		models.TaxFieldName: map[string]any{"en-US": code},
	}
}

func definitionRow(id, taxId, from, to string, calculationType models.CalculationType) dmodel.DynamicFields {
	row := dmodel.DynamicFields{
		models.TaxDefinitionVersionFieldId:              id,
		models.TaxDefinitionVersionFieldTaxId:           taxId,
		models.TaxDefinitionVersionFieldVersionNo:       int32(1),
		models.TaxDefinitionVersionFieldCalculationType: string(calculationType),
		models.TaxDefinitionVersionFieldTaxTreatment:    string(models.TaxTreatmentTaxable),
		models.TaxDefinitionVersionFieldLifecycleStatus: string(models.LifecyclePublished),
		models.TaxDefinitionVersionFieldEffectiveFrom:   dateOf(from),
	}
	if to != "" {
		row[models.TaxDefinitionVersionFieldEffectiveTo] = dateOf(to)
	}
	return row
}

func rateRow(id, taxId, from, to, rate string) dmodel.DynamicFields {
	row := dmodel.DynamicFields{
		models.TaxRateVersionFieldId:              id,
		models.TaxRateVersionFieldTaxId:           taxId,
		models.TaxRateVersionFieldVersionNo:       int32(1),
		models.TaxRateVersionFieldRate:            decimalOf(rate),
		models.TaxRateVersionFieldLifecycleStatus: string(models.LifecyclePublished),
		models.TaxRateVersionFieldEffectiveFrom:   dateOf(from),
	}
	if to != "" {
		row[models.TaxRateVersionFieldEffectiveTo] = dateOf(to)
	}
	return row
}

// reposFor wires a set of fakes into the repository bundle the resolver takes.
func reposFor(tax, definition, rate, component *fakeRepo) *TaxRepos {
	return &TaxRepos{
		Tax:               tax,
		DefinitionVersion: definition,
		RateVersion:       rate,
		Component:         component,
	}
}

func TestResolveTaxReadsTheEffectiveVersions(t *testing.T) {
	repos := reposFor(
		newFakeRepo(taxRow("t1", "VAT10")),
		newFakeRepo(definitionRow("dv1", "t1", "2025-01-01", "", models.CalculationPercentage)),
		newFakeRepo(rateRow("rv1", "t1", "2025-01-01", "", "10")),
		newFakeRepo(),
	)

	resolved, problem, err := ResolveTax(nil, repos, "t1", "2026-08-25")

	if err != nil || problem != nil {
		t.Fatalf("expected a clean resolution, got problem=%+v err=%v", problem, err)
	}
	if resolved.TaxCode != "VAT10" || resolved.TaxName != "VAT10" {
		t.Errorf("expected the tax identity resolved, got code=%q name=%q", resolved.TaxCode, resolved.TaxName)
	}
	if resolved.DefinitionVersionId != "dv1" || resolved.RateVersionId != "rv1" {
		t.Errorf("expected both version ids, got dv=%q rv=%q",
			resolved.DefinitionVersionId, resolved.RateVersionId)
	}
	if !resolved.Rate.Equal(decimalOf("10")) {
		t.Errorf("expected rate 10, got %s", resolved.Rate)
	}
}

// A tax with no definition in force on the date is unresolved and says so specifically: a missing
// definition is a different fault from a missing rate.
func TestResolveTaxOutsideEffectivePeriodIsUnresolved(t *testing.T) {
	repos := reposFor(
		newFakeRepo(taxRow("t1", "VAT10")),
		newFakeRepo(definitionRow("dv1", "t1", "2025-01-01", "2025-12-31", models.CalculationPercentage)),
		newFakeRepo(rateRow("rv1", "t1", "2025-01-01", "", "10")),
		newFakeRepo(),
	)

	_, problem, err := ResolveTax(nil, repos, "t1", "2026-08-25")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if problem == nil || problem.ErrorCode != models.ErrCodeTaxDefinitionMissing {
		t.Fatalf("expected tax_definition_missing, got %+v", problem)
	}
}

// A tie is never broken by taking the newest: two published versions covering one date means the
// configuration is corrupt, and picking one would charge under a configuration nobody chose.
func TestTwoEffectiveDefinitionsAreAmbiguousNotResolved(t *testing.T) {
	repos := reposFor(
		newFakeRepo(taxRow("t1", "VAT10")),
		newFakeRepo(
			definitionRow("dv1", "t1", "2025-01-01", "", models.CalculationPercentage),
			definitionRow("dv2", "t1", "2026-01-01", "", models.CalculationPercentage),
		),
		newFakeRepo(rateRow("rv1", "t1", "2025-01-01", "", "10")),
		newFakeRepo(),
	)

	_, problem, err := ResolveTax(nil, repos, "t1", "2026-08-25")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if problem == nil || problem.ErrorCode != models.ErrCodeTaxDefinitionAmbiguous {
		t.Fatalf("expected tax_definition_ambiguous, got %+v", problem)
	}
}

func TestTwoEffectiveRatesAreAmbiguous(t *testing.T) {
	repos := reposFor(
		newFakeRepo(taxRow("t1", "VAT10")),
		newFakeRepo(definitionRow("dv1", "t1", "2025-01-01", "", models.CalculationPercentage)),
		newFakeRepo(
			rateRow("rv1", "t1", "2025-01-01", "", "10"),
			rateRow("rv2", "t1", "2026-01-01", "", "8"),
		),
		newFakeRepo(),
	)

	_, problem, _ := ResolveTax(nil, repos, "t1", "2026-08-25")

	if problem == nil || problem.ErrorCode != models.ErrCodeTaxRateAmbiguous {
		t.Fatalf("expected tax_rate_ambiguous, got %+v", problem)
	}
}

func TestMissingRateIsDistinctFromMissingDefinition(t *testing.T) {
	repos := reposFor(
		newFakeRepo(taxRow("t1", "VAT10")),
		newFakeRepo(definitionRow("dv1", "t1", "2025-01-01", "", models.CalculationPercentage)),
		newFakeRepo(), // no rate at all
		newFakeRepo(),
	)

	_, problem, _ := ResolveTax(nil, repos, "t1", "2026-08-25")

	if problem == nil || problem.ErrorCode != models.ErrCodeTaxRateMissing {
		t.Fatalf("expected tax_rate_missing, got %+v", problem)
	}
}

func TestUnknownTaxIsReportedNotPanicked(t *testing.T) {
	repos := reposFor(newFakeRepo(), newFakeRepo(), newFakeRepo(), newFakeRepo())

	_, problem, err := ResolveTax(nil, repos, "nope", "2026-08-25")

	if err != nil {
		t.Fatalf("a missing tax is a business outcome, not an error: %v", err)
	}
	if problem == nil || problem.ErrorCode != models.ErrCodeTaxNotFound {
		t.Fatalf("expected tax_not_found, got %+v", problem)
	}
}

// A "none" calculation carries a legal treatment and no rate, so an exempt tax with no rate version
// is correct rather than broken.
func TestNoneCalculationNeedsNoRate(t *testing.T) {
	definition := definitionRow("dv1", "t1", "2025-01-01", "", models.CalculationNone)
	definition[models.TaxDefinitionVersionFieldTaxTreatment] = string(models.TaxTreatmentExempt)

	rates := newFakeRepo()
	repos := reposFor(
		newFakeRepo(taxRow("t1", "VAT_EXEMPT")),
		newFakeRepo(definition),
		rates,
		newFakeRepo(),
	)

	resolved, problem, err := ResolveTax(nil, repos, "t1", "2026-08-25")

	if err != nil || problem != nil {
		t.Fatalf("expected an exempt tax to resolve, got problem=%+v err=%v", problem, err)
	}
	if resolved.Treatment != models.TaxTreatmentExempt {
		t.Errorf("expected the exempt treatment carried, got %q", resolved.Treatment)
	}
	if rates.calls != 0 {
		t.Error("expected no rate lookup for a none-typed tax")
	}
}

// A group tax delegates to its components and has no rate of its own. The components' own sequence
// orders them, overriding whatever sequence their standalone definitions carry.
func TestGroupTaxResolvesItsComponentsInSequence(t *testing.T) {
	repos := &TaxRepos{
		Tax: newFakeRepo(taxRow("g1", "GROUP"), taxRow("c1", "CHILD1"), taxRow("c2", "CHILD2")),
		DefinitionVersion: newFakeRepo(
			definitionRow("dvg", "g1", "2025-01-01", "", models.CalculationGroup),
			definitionRow("dv1", "c1", "2025-01-01", "", models.CalculationPercentage),
			definitionRow("dv2", "c2", "2025-01-01", "", models.CalculationPercentage),
		),
		RateVersion: newFakeRepo(
			rateRow("rv1", "c1", "2025-01-01", "", "10"),
			rateRow("rv2", "c2", "2025-01-01", "", "5"),
		),
		Component: newFakeRepo(
			dmodel.DynamicFields{
				models.TaxComponentFieldId:                           "cmp2",
				models.TaxComponentFieldParentTaxDefinitionVersionId: "dvg",
				models.TaxComponentFieldComponentTaxId:               "c2",
				models.TaxComponentFieldSequence:                     int32(2),
			},
			dmodel.DynamicFields{
				models.TaxComponentFieldId:                           "cmp1",
				models.TaxComponentFieldParentTaxDefinitionVersionId: "dvg",
				models.TaxComponentFieldComponentTaxId:               "c1",
				models.TaxComponentFieldSequence:                     int32(1),
			},
		),
	}

	resolved, problem, err := ResolveTax(nil, repos, "g1", "2026-08-25")

	if err != nil || problem != nil {
		t.Fatalf("expected the group to resolve, got problem=%+v err=%v", problem, err)
	}
	if len(resolved.Components) != 2 {
		t.Fatalf("expected 2 components, got %d", len(resolved.Components))
	}
	// Supplied out of order on purpose: the database guarantees no ordering and a compound chain is
	// order-dependent, so the resolver has to sort.
	if resolved.Components[0].TaxId != "c1" || resolved.Components[1].TaxId != "c2" {
		t.Fatalf("expected components sorted by sequence, got %s then %s",
			resolved.Components[0].TaxId, resolved.Components[1].TaxId)
	}

	specs := ToComponentSpecs(*resolved, nil)
	if len(specs) != 2 {
		t.Fatalf("expected the group to flatten to its 2 children, got %d", len(specs))
	}
	for _, spec := range specs {
		if spec.TaxId == "g1" {
			t.Fatal("the group itself must contribute no component, or the line is taxed twice")
		}
	}
}

// A group with no components is reported rather than returned as a group that would silently
// compute no tax.
func TestGroupWithNoComponentsIsInvalid(t *testing.T) {
	repos := reposFor(
		newFakeRepo(taxRow("g1", "GROUP")),
		newFakeRepo(definitionRow("dvg", "g1", "2025-01-01", "", models.CalculationGroup)),
		newFakeRepo(),
		newFakeRepo(),
	)

	_, problem, err := ResolveTax(nil, repos, "g1", "2026-08-25")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if problem == nil || problem.ErrorCode != models.ErrCodeTaxConfigurationInvalid {
		t.Fatalf("expected tax_configuration_invalid, got %+v", problem)
	}
}

// A component cycle would recurse forever. Reaching the depth limit means a row was written around
// the validation guard, so the resolver refuses rather than taking the process down.
func TestComponentCycleIsRefusedRatherThanLooping(t *testing.T) {
	repos := &TaxRepos{
		Tax: newFakeRepo(taxRow("g1", "GROUP")),
		DefinitionVersion: newFakeRepo(
			definitionRow("dvg", "g1", "2025-01-01", "", models.CalculationGroup),
		),
		RateVersion: newFakeRepo(),
		// The group's only component is the group itself.
		Component: newFakeRepo(dmodel.DynamicFields{
			models.TaxComponentFieldId:                           "cmp1",
			models.TaxComponentFieldParentTaxDefinitionVersionId: "dvg",
			models.TaxComponentFieldComponentTaxId:               "g1",
			models.TaxComponentFieldSequence:                     int32(1),
		}),
	}

	_, problem, err := ResolveTax(nil, repos, "g1", "2026-08-25")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if problem == nil || problem.ErrorCode != models.ErrCodeTaxConfigurationInvalid {
		t.Fatalf("expected the cycle refused as invalid configuration, got %+v", problem)
	}
}

// A draft version is invisible to the engine: only published configuration prices a transaction.
func TestDraftVersionsAreNotEffective(t *testing.T) {
	draft := definitionRow("dv1", "t1", "2025-01-01", "", models.CalculationPercentage)
	draft[models.TaxDefinitionVersionFieldLifecycleStatus] = string(models.LifecycleDraft)

	repos := reposFor(
		newFakeRepo(taxRow("t1", "VAT10")),
		newFakeRepo(draft),
		newFakeRepo(rateRow("rv1", "t1", "2025-01-01", "", "10")),
		newFakeRepo(),
	)

	_, problem, _ := ResolveTax(nil, repos, "t1", "2026-08-25")

	if problem == nil || problem.ErrorCode != models.ErrCodeTaxDefinitionMissing {
		t.Fatalf("expected a draft to be invisible to the engine, got %+v", problem)
	}
}

// A withdrawn version stays readable for audit but must not price a new transaction.
func TestWithdrawnVersionsAreNotEffective(t *testing.T) {
	withdrawn := definitionRow("dv1", "t1", "2025-01-01", "", models.CalculationPercentage)
	withdrawn[models.TaxDefinitionVersionFieldLifecycleStatus] = string(models.LifecycleWithdrawn)

	repos := reposFor(
		newFakeRepo(taxRow("t1", "VAT10")),
		newFakeRepo(withdrawn),
		newFakeRepo(rateRow("rv1", "t1", "2025-01-01", "", "10")),
		newFakeRepo(),
	)

	_, problem, _ := ResolveTax(nil, repos, "t1", "2026-08-25")

	if problem == nil || problem.ErrorCode != models.ErrCodeTaxDefinitionMissing {
		t.Fatalf("expected a withdrawn version to be invisible to the engine, got %+v", problem)
	}
}

// Effective bounds are inclusive at both ends: a rate change takes effect at midnight, so the last
// covered day must still resolve.
func TestEffectivePeriodBoundsAreInclusive(t *testing.T) {
	build := func() *TaxRepos {
		return reposFor(
			newFakeRepo(taxRow("t1", "VAT10")),
			newFakeRepo(definitionRow("dv1", "t1", "2025-07-01", "2025-12-31", models.CalculationPercentage)),
			newFakeRepo(rateRow("rv1", "t1", "2025-07-01", "2025-12-31", "10")),
			newFakeRepo(),
		)
	}

	for _, date := range []string{"2025-07-01", "2025-12-31"} {
		if _, problem, _ := ResolveTax(nil, build(), "t1", date); problem != nil {
			t.Errorf("expected %s to be inside the period, got %+v", date, problem)
		}
	}
	for _, date := range []string{"2025-06-30", "2026-01-01"} {
		if _, problem, _ := ResolveTax(nil, build(), "t1", date); problem == nil {
			t.Errorf("expected %s to fall outside the period", date)
		}
	}
}

// The same request against the same configuration must produce the same answer; a map iteration
// leaking into the result would show up as an occasional difference here.
func TestResolutionIsDeterministic(t *testing.T) {
	build := func() *TaxRepos {
		return reposFor(
			newFakeRepo(taxRow("t1", "VAT10")),
			newFakeRepo(definitionRow("dv1", "t1", "2025-01-01", "", models.CalculationPercentage)),
			newFakeRepo(rateRow("rv1", "t1", "2025-01-01", "", "10")),
			newFakeRepo(),
		)
	}

	first, _, _ := ResolveTax(nil, build(), "t1", "2026-08-25")
	for attempt := 0; attempt < 20; attempt++ {
		again, _, _ := ResolveTax(nil, build(), "t1", "2026-08-25")
		if again.DefinitionVersionId != first.DefinitionVersionId ||
			again.RateVersionId != first.RateVersionId ||
			!again.Rate.Equal(first.Rate) {
			t.Fatalf("resolution differed between runs: %+v vs %+v", first, again)
		}
	}
}
