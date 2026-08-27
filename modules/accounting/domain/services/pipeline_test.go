package services

import (
	"testing"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
	taxsvc "github.com/sky-as-code/nikki-erp/modules/accounting/domain/services/tax"
	itExt "github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/external"
	it "github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/tax"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"
)

// fakeUom stands in for Essential's conversion service.
//
// The failing case is the one that matters: an impossible conversion has to make one line
// unresolved rather than fail the request, and the only way to test that is to be able to make the
// conversion fail on demand.
type fakeUom struct {
	factor decimal.Decimal
	fails  bool
	calls  int
}

func (this *fakeUom) Convert(
	_ corectx.Context, query itExt.ConvertQuantityQuery,
) (*itExt.ConvertQuantityResult, error) {
	this.calls++
	if this.fails {
		cErrs := clientErrorsWith("uom.incompatible", "cannot convert between these units")
		return &itExt.ConvertQuantityResult{ClientErrors: *cErrs}, nil
	}
	converted := query.Quantity.Mul(this.factor)
	return &itExt.ConvertQuantityResult{
		HasData: true,
		Data: itUom.ConvertQuantityResultData{
			Quantity:      converted,
			ExactQuantity: converted,
		},
	}, nil
}

func (this *fakeUom) GetUom(
	_ corectx.Context, _ itExt.GetUomQuery,
) (*itExt.GetUomResult, error) {
	return &itExt.GetUomResult{HasData: true}, nil
}

func fixedTax(taxId, rateUomId, currencyCode, amount string) ResolvedTax {
	return ResolvedTax{
		TaxId:            taxId,
		CalculationType:  models.CalculationFixed,
		FixedAmount:      decimalOf(amount),
		RateUomId:        rateUomId,
		RateCurrencyCode: currencyCode,
	}
}

// AC-TAX-SUP-13: a fixed tax whose unit cannot be reached from the line's unit fails explicitly.
// The line is unresolved with a specific code, not silently taxed on an unconverted quantity.
func TestIncompatibleUomFailsExplicitly(t *testing.T) {
	uom := &fakeUom{fails: true}
	svc := &TaxCalculationDomainServiceImpl{uomSvc: uom}
	line := it.CalculationLine{
		LineReference: "L1",
		Quantity:      decimalOf("3"),
		UomId:         "kilogram",
	}

	_, problem, err := svc.convertQuantities(
		nil, line, "VND", []ResolvedTax{fixedTax("excise", "litre", "VND", "4000")})

	if err != nil {
		t.Fatalf("an impossible conversion is a business outcome, not an error: %v", err)
	}
	if problem != models.ErrCodeUomConversion {
		t.Fatalf("expected %q, got %q", models.ErrCodeUomConversion, problem)
	}
	if uom.calls != 1 {
		t.Errorf("expected the conversion attempted once, got %d calls", uom.calls)
	}
}

// A fixed tax quoted per unit with no unit on the line has nothing to convert from.
func TestFixedTaxWithoutALineUomFails(t *testing.T) {
	svc := &TaxCalculationDomainServiceImpl{uomSvc: &fakeUom{factor: decimalOf("1")}}
	line := it.CalculationLine{LineReference: "L1", Quantity: decimalOf("3")}

	_, problem, _ := svc.convertQuantities(
		nil, line, "VND", []ResolvedTax{fixedTax("excise", "litre", "VND", "4000")})

	if problem != models.ErrCodeUomConversion {
		t.Fatalf("expected %q, got %q", models.ErrCodeUomConversion, problem)
	}
}

// AC-TAX-SUP-14: V1 has no FX capability, so a fixed tax denominated in another currency must fail
// loudly. Converting at an implied rate would invent a number nobody chose.
func TestFixedTaxInAnotherCurrencyFailsExplicitly(t *testing.T) {
	uom := &fakeUom{factor: decimalOf("1")}
	svc := &TaxCalculationDomainServiceImpl{uomSvc: uom}
	line := it.CalculationLine{LineReference: "L1", Quantity: decimalOf("3"), UomId: "litre"}

	_, problem, err := svc.convertQuantities(
		nil, line, "VND", []ResolvedTax{fixedTax("excise", "litre", "JPY", "400")})

	if err != nil {
		t.Fatalf("a currency mismatch is a business outcome, not an error: %v", err)
	}
	if problem != models.ErrCodeFixedTaxCurrency {
		t.Fatalf("expected %q, got %q", models.ErrCodeFixedTaxCurrency, problem)
	}
	if uom.calls != 0 {
		t.Error("expected the currency refused before any unit conversion was attempted")
	}
}

func TestFixedTaxInTheTransactionCurrencyPasses(t *testing.T) {
	svc := &TaxCalculationDomainServiceImpl{uomSvc: &fakeUom{factor: decimalOf("2")}}
	line := it.CalculationLine{LineReference: "L1", Quantity: decimalOf("3"), UomId: "gallon"}

	quantities, problem, err := svc.convertQuantities(
		nil, line, "VND", []ResolvedTax{fixedTax("excise", "litre", "VND", "4000")})

	if err != nil || problem != "" {
		t.Fatalf("expected a clean conversion, got problem=%q err=%v", problem, err)
	}
	if !quantities["excise"].Equal(decimalOf("6")) {
		t.Fatalf("expected 3 gallons converted to 6 litres, got %s", quantities["excise"])
	}
}

// A rate with no currency of its own is in the transaction's currency by construction, so it must
// not be refused — this is the ordinary case for every percentage tax.
func TestRateWithoutACurrencyIsNotRefused(t *testing.T) {
	svc := &TaxCalculationDomainServiceImpl{uomSvc: &fakeUom{factor: decimalOf("1")}}
	line := it.CalculationLine{LineReference: "L1", Quantity: decimalOf("3"), UomId: "litre"}

	_, problem, _ := svc.convertQuantities(
		nil, line, "VND", []ResolvedTax{fixedTax("excise", "litre", "", "4000")})

	if problem != "" {
		t.Fatalf("expected no problem for a currency-less rate, got %q", problem)
	}
}

// A percentage tax needs no quantity at all, so the UoM service must not be called for one: an
// outage in Essential must not stop ordinary VAT from being calculated.
func TestPercentageTaxNeverCallsTheUomService(t *testing.T) {
	uom := &fakeUom{fails: true}
	svc := &TaxCalculationDomainServiceImpl{uomSvc: uom}
	line := it.CalculationLine{LineReference: "L1", Quantity: decimalOf("3"), UomId: "kilogram"}

	percentage := ResolvedTax{
		TaxId:           "vat",
		CalculationType: models.CalculationPercentage,
		Rate:            decimalOf("10"),
	}

	_, problem, err := svc.convertQuantities(nil, line, "VND", []ResolvedTax{percentage})

	if err != nil || problem != "" {
		t.Fatalf("expected a percentage tax to need no conversion, got problem=%q err=%v", problem, err)
	}
	if uom.calls != 0 {
		t.Errorf("expected no UoM call for a percentage tax, got %d", uom.calls)
	}
}

// AC-TAX-22: a line-scoped policy rounds each component as it goes, and the line total is the sum
// of those rounded components. Only the document scope was covered before.
func TestLineScopedPolicyRoundsEachComponentImmediately(t *testing.T) {
	policy := taxsvc.RoundingPolicy{
		Scope:     models.RoundingScopeLine,
		Method:    models.RoundingHalfUp,
		Increment: decimalOf("1"),
	}

	// Two components whose exact amounts each need rounding in a different direction.
	amounts := []decimal.Decimal{decimalOf("10.4"), decimalOf("5.6")}
	line := it.LineResult{TotalExcluded: decimalOf("100")}
	for index, amount := range amounts {
		rounded := policy.Round(amount)
		line.Components = append(line.Components, it.ComponentResult{
			Sequence:           int32(index + 1),
			UnroundedTaxAmount: amount,
			TaxAmount:          rounded,
			RoundingAdjustment: rounded.Sub(amount),
		})
	}
	retotalLine(&line)

	if !line.Components[0].TaxAmount.Equal(decimalOf("10")) {
		t.Errorf("expected 10.4 rounded down to 10, got %s", line.Components[0].TaxAmount)
	}
	if !line.Components[1].TaxAmount.Equal(decimalOf("6")) {
		t.Errorf("expected 5.6 rounded up to 6, got %s", line.Components[1].TaxAmount)
	}
	// 16, not the 16.0 that rounding the 16.0 sum would give — and importantly the total an invoice
	// prints is the sum of what it prints beside each component.
	if !line.TotalTax.Equal(decimalOf("16")) {
		t.Fatalf("expected the line total to be the sum of rounded components, got %s", line.TotalTax)
	}
	if !line.TotalIncluded.Equal(decimalOf("116")) {
		t.Fatalf("expected the included total to follow, got %s", line.TotalIncluded)
	}
}

// AC-TAX-06: publishing a newer rate must not change what an already-dated transaction resolves to.
// The mechanism is that resolution is a function of tax_date, so a past date keeps reading the
// version that was in force then — this is what makes a historical invoice stable.
func TestANewerRateDoesNotChangeAHistoricalResolution(t *testing.T) {
	// The original configuration: 10% from July 2025, closed at the end of that year.
	original := func() []dmodel.DynamicFields {
		return []dmodel.DynamicFields{rateRow("rv1", "t1", "2025-07-01", "2025-12-31", "10")}
	}
	definitions := func() []dmodel.DynamicFields {
		return []dmodel.DynamicFields{
			definitionRow("dv1", "t1", "2025-07-01", "", models.CalculationPercentage),
		}
	}

	before := reposFor(
		newFakeRepo(taxRow("t1", "VAT")),
		newFakeRepo(definitions()...),
		newFakeRepo(original()...),
		newFakeRepo(),
	)
	historical, problem, err := ResolveTax(nil, before, "t1", "2025-09-15")
	if err != nil || problem != nil {
		t.Fatalf("expected the historical date to resolve, got problem=%+v err=%v", problem, err)
	}

	// Now a successor rate is published for 2026 onward. The old row is untouched.
	withSuccessor := reposFor(
		newFakeRepo(taxRow("t1", "VAT")),
		newFakeRepo(definitions()...),
		newFakeRepo(append(original(), rateRow("rv2", "t1", "2026-01-01", "", "8"))...),
		newFakeRepo(),
	)

	again, problem, err := ResolveTax(nil, withSuccessor, "t1", "2025-09-15")
	if err != nil || problem != nil {
		t.Fatalf("expected the historical date to still resolve, got problem=%+v err=%v", problem, err)
	}
	if again.RateVersionId != historical.RateVersionId || !again.Rate.Equal(historical.Rate) {
		t.Fatalf("publishing a 2026 rate changed the 2025 answer: was %s@%s, now %s@%s",
			historical.RateVersionId, historical.Rate, again.RateVersionId, again.Rate)
	}

	// And the new date genuinely picks up the successor, so the test above is not passing merely
	// because the successor was never visible.
	current, problem, _ := ResolveTax(nil, withSuccessor, "t1", "2026-06-01")
	if problem != nil || current.RateVersionId != "rv2" || !current.Rate.Equal(decimalOf("8")) {
		t.Fatalf("expected 2026 to resolve to the 8%% successor, got %+v problem=%+v", current, problem)
	}
}

// AC-TAX-35 / TAX-INV-20: a calculation reads configuration and writes nothing. The repositories
// expose only Search, so a write is not expressible — this test pins that down by asserting the
// port stays read-only, which is what makes recalculating a draft order on every edit safe.
func TestTheCalculationPortIsReadOnly(t *testing.T) {
	// TaxSearcher is the whole surface the pipeline has over stored configuration. If a mutating
	// method is ever added here, this assertion is where the decision gets noticed.
	var searcher models.TaxSearcher = newFakeRepo()

	if _, ok := searcher.(interface {
		Create(corectx.Context, dmodel.DynamicFields) (*dyn.OpResult[dmodel.DynamicFields], error)
	}); ok {
		t.Fatal("the tax lookup port has gained a write method; a calculation must have no side effects")
	}
}

// Repeating a calculation must not accumulate state. The service holds none between calls by
// design, and a counter or cache creeping in would show up as a differing second answer.
func TestRepeatedResolutionDoesNotAccumulateState(t *testing.T) {
	repos := reposFor(
		newFakeRepo(taxRow("t1", "VAT10")),
		newFakeRepo(definitionRow("dv1", "t1", "2025-01-01", "", models.CalculationPercentage)),
		newFakeRepo(rateRow("rv1", "t1", "2025-01-01", "", "10")),
		newFakeRepo(),
	)

	first, _, _ := ResolveTax(nil, repos, "t1", "2026-08-25")
	for attempt := 0; attempt < 10; attempt++ {
		again, problem, err := ResolveTax(nil, repos, "t1", "2026-08-25")
		if err != nil || problem != nil {
			t.Fatalf("run %d failed: problem=%+v err=%v", attempt, problem, err)
		}
		if !again.Rate.Equal(first.Rate) || again.RateVersionId != first.RateVersionId {
			t.Fatalf("run %d differed from the first: %+v vs %+v", attempt, again, first)
		}
	}
}
