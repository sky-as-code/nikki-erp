package services

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services/pricing"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// stubTaxService returns a canned answer, so these tests exercise the mapping and the refusal rules
// rather than Accounting's arithmetic, which has its own tests in its own module.
type stubTaxService struct {
	result  *itExt.CalculateResult
	err     error
	request itExt.CalculationRequest
	calls   int
}

func (this *stubTaxService) Calculate(
	_ corectx.Context, request itExt.CalculationRequest,
) (*itExt.CalculateResult, error) {
	this.calls++
	this.request = request
	return this.result, this.err
}

func (this *stubTaxService) ReverseFull(
	_ corectx.Context, _ itExt.FullReversalRequest,
) (*itExt.ReverseResult, error) {
	return nil, nil
}

func (this *stubTaxService) ReversePartial(
	_ corectx.Context, _ itExt.PartialReversalRequest,
) (*itExt.ReverseResult, error) {
	return nil, nil
}

var _ itExt.TaxCalculationExtService = (*stubTaxService)(nil)

func resolvedResult(lines ...itExt.TaxLineResult) *itExt.CalculateResult {
	total := decimal.Zero
	for _, line := range lines {
		total = total.Add(line.TotalTax)
	}
	return &itExt.CalculateResult{
		HasData: true,
		Data: itExt.CalculationResult{
			Status:   itExt.DeterminationResolved,
			TotalTax: total,
			Lines:    lines,
		},
	}
}

func pricedLine(key string, net string) pricing.LineResult {
	return pricing.LineResult{
		Key:              key,
		ProductVariantId: "variant-" + key,
		UomId:            "uom",
		Quantity:         dec("1"),
		NetAmount:        dec(net),
		GrossAmount:      dec(net),
	}
}

func testContext() TaxRequestContext {
	return TaxRequestContext{
		OrgId:        "org",
		TaxDate:      "2026-08-26",
		CurrencyCode: "VND",
		TaxCode:      "01M4ACCTVN0000000000000020",
		PriceMode:    itExt.PriceInclusionIncluded,
	}
}

// THE test of this file. Accounting stores a rate as a PERCENTAGE and Sales stores
// tax_rate_snapshot as a FRACTION. Getting the conversion wrong is a hundredfold error that no
// total flags as impossible: a 10% rate stored as 10 reads as 1000%.
func TestRateIsConvertedFromPercentageToFraction(t *testing.T) {
	taxSvc := &stubTaxService{result: resolvedResult(itExt.TaxLineResult{
		LineReference: "a",
		Status:        itExt.DeterminationResolved,
		TotalTax:      dec("10000"),
		Components: []itExt.TaxComponentResult{
			{TaxCode: "VN_VAT_10", Rate: dec("10"), TaxAmount: dec("10000")},
		},
	})}

	tax, vErrs, err := ResolveBasketTax(nil, taxSvc, testContext(),
		[]pricing.LineResult{pricedLine("a", "110000")})

	require.NoError(t, err)
	require.Nil(t, vErrs)
	require.NotNil(t, tax)

	assert.True(t, tax.ByLineKey["a"].RateSnapshot.Equal(dec("0.1")),
		"a rate of 10 percent must be stored as the fraction 0.1, got %s",
		tax.ByLineKey["a"].RateSnapshot)
	assert.True(t, tax.ByLineKey["a"].Amount.Equal(dec("10000")))
	assert.True(t, tax.Total.Equal(dec("10000")))
}

// An 8% rate converts to 0.08, not 0.8; the second is what a division by ten rather than a hundred
// would produce.
func TestEightPercentConvertsToEightHundredths(t *testing.T) {
	taxSvc := &stubTaxService{result: resolvedResult(itExt.TaxLineResult{
		LineReference: "a",
		Status:        itExt.DeterminationResolved,
		TotalTax:      dec("7407"),
		Components: []itExt.TaxComponentResult{
			{TaxCode: "VN_VAT_8", Rate: dec("8"), TaxAmount: dec("7407")},
		},
	})}

	tax, _, err := ResolveBasketTax(nil, taxSvc, testContext(),
		[]pricing.LineResult{pricedLine("a", "100000")})

	require.NoError(t, err)
	assert.True(t, tax.ByLineKey["a"].RateSnapshot.Equal(dec("0.08")),
		"8 percent must be 0.08, got %s", tax.ByLineKey["a"].RateSnapshot)
}

// The whole point of failing closed. An undetermined tax must never read as zero: that
// under-charges VAT silently and surfaces at a tax audit rather than at the till.
func TestUnresolvedTaxIsRefusedRatherThanTreatedAsZero(t *testing.T) {
	taxSvc := &stubTaxService{result: &itExt.CalculateResult{
		HasData: true,
		Data: itExt.CalculationResult{
			Status: itExt.DeterminationUnresolved,
			Lines: []itExt.TaxLineResult{{
				LineReference: "a",
				Status:        itExt.DeterminationUnresolved,
				ErrorCode:     "tax_rate_missing",
			}},
		},
	}}

	tax, vErrs, err := ResolveBasketTax(nil, taxSvc, testContext(),
		[]pricing.LineResult{pricedLine("a", "110000")})

	require.NoError(t, err, "an undetermined tax is a configuration fault, not a server fault")
	require.Nil(t, tax, "no tax may be returned when none could be determined")
	require.NotNil(t, vErrs)
	assert.Positive(t, vErrs.Count())
	assert.Contains(t, vErrs.ToError().Error(), "tax_rate_missing",
		"the error must name WHY, so an administrator knows which configuration to fix")
}

// A missing port is not a failure: a deployment without accounting sells untaxed.
func TestNoTaxServiceMeansEveryLineIsTaxedAtZero(t *testing.T) {
	lines := []pricing.LineResult{pricedLine("a", "110000"), pricedLine("b", "50000")}

	tax, vErrs, err := ResolveBasketTax(nil, nil, testContext(), lines)

	require.NoError(t, err)
	require.Nil(t, vErrs)
	require.NotNil(t, tax)
	assert.Len(t, tax.ByLineKey, 2, "every line must be present, taxed at zero")
	assert.True(t, tax.Total.IsZero())
	for key, lineTax := range tax.ByLineKey {
		assert.True(t, lineTax.Amount.IsZero(), key)
		assert.True(t, lineTax.RateSnapshot.IsZero(), key)
	}
}

// An organization with no tax code configured is untaxed, and Accounting is not called at all.
func TestEmptyTaxCodeSkipsTheCall(t *testing.T) {
	taxSvc := &stubTaxService{}
	context := testContext()
	context.TaxCode = ""

	tax, vErrs, err := ResolveBasketTax(nil, taxSvc, context,
		[]pricing.LineResult{pricedLine("a", "110000")})

	require.NoError(t, err)
	require.Nil(t, vErrs)
	require.NotNil(t, tax)
	assert.Zero(t, taxSvc.calls, "with no tax to apply there is nothing to ask Accounting")
	assert.True(t, tax.Total.IsZero())
}

// The taxable base is the line's NET, after discounts. Passing gross would tax the customer on
// money they never paid.
func TestTheTaxableBaseIsNetOfDiscount(t *testing.T) {
	line := pricedLine("a", "190000")
	line.GrossAmount = dec("200000")
	line.DiscountAmount = dec("10000")

	taxSvc := &stubTaxService{result: resolvedResult(itExt.TaxLineResult{
		LineReference: "a", Status: itExt.DeterminationResolved, TotalTax: dec("15200"),
	})}

	_, _, err := ResolveBasketTax(nil, taxSvc, testContext(), []pricing.LineResult{line})
	require.NoError(t, err)

	require.Len(t, taxSvc.request.Lines, 1)
	sent := taxSvc.request.Lines[0]
	assert.True(t, sent.CommercialBaseAmount.Equal(dec("190000")),
		"the base must be the net 190000, not the gross 200000, got %s", sent.CommercialBaseAmount)
	assert.True(t, sent.DiscountAmount.Equal(dec("10000")),
		"the discount travels for audit even though it is not part of the base")
}

// One request for the whole basket, never one per line: a document-scoped rounding policy rounds
// the total once, and per-line calls summed afterwards produce a number no policy asked for.
func TestTheWholeBasketIsOneRequest(t *testing.T) {
	lines := []pricing.LineResult{
		pricedLine("a", "110000"), pricedLine("b", "50000"), pricedLine("c", "20000"),
	}
	taxSvc := &stubTaxService{result: resolvedResult(
		itExt.TaxLineResult{LineReference: "a", Status: itExt.DeterminationResolved},
		itExt.TaxLineResult{LineReference: "b", Status: itExt.DeterminationResolved},
		itExt.TaxLineResult{LineReference: "c", Status: itExt.DeterminationResolved},
	)}

	_, _, err := ResolveBasketTax(nil, taxSvc, testContext(), lines)

	require.NoError(t, err)
	assert.Equal(t, 1, taxSvc.calls, "three lines must be one document-level call, not three")
	assert.Len(t, taxSvc.request.Lines, 3)
}

// The tax date is the caller's and is never defaulted from the clock: a sale must be taxed under
// the configuration in force when it legally happened.
func TestTheTaxDateTravelsUnchanged(t *testing.T) {
	taxSvc := &stubTaxService{result: resolvedResult(
		itExt.TaxLineResult{LineReference: "a", Status: itExt.DeterminationResolved})}
	context := testContext()
	context.TaxDate = "2025-01-15"

	_, _, err := ResolveBasketTax(nil, taxSvc, context,
		[]pricing.LineResult{pricedLine("a", "110000")})

	require.NoError(t, err)
	assert.Equal(t, "2025-01-15", taxSvc.request.TaxDate)
	assert.Equal(t, itExt.OperationSale, taxSvc.request.OperationType)
}

// A transport failure inside Accounting is a 500, not a validation error: it is not something a
// caller can fix by correcting a form.
func TestATransportFailurePropagatesAsAnError(t *testing.T) {
	taxSvc := &stubTaxService{err: assert.AnError}

	tax, vErrs, err := ResolveBasketTax(nil, taxSvc, testContext(),
		[]pricing.LineResult{pricedLine("a", "110000")})

	require.Error(t, err)
	assert.Nil(t, tax)
	assert.Nil(t, vErrs, "a server fault must not be reported as a client error")
}

// A component-less line reads as a zero rate rather than dividing by nothing.
func TestALineWithNoComponentsHasAZeroRate(t *testing.T) {
	taxSvc := &stubTaxService{result: resolvedResult(itExt.TaxLineResult{
		LineReference: "a", Status: itExt.DeterminationResolved, TotalTax: decimal.Zero,
	})}

	tax, _, err := ResolveBasketTax(nil, taxSvc, testContext(),
		[]pricing.LineResult{pricedLine("a", "110000")})

	require.NoError(t, err)
	assert.True(t, tax.ByLineKey["a"].RateSnapshot.IsZero())
}

// Several taxes on one line sum into the effective rate, so the line can be read without unpacking
// the snapshot.
func TestSeveralComponentsSumIntoTheEffectiveRate(t *testing.T) {
	taxSvc := &stubTaxService{result: resolvedResult(itExt.TaxLineResult{
		LineReference: "a",
		Status:        itExt.DeterminationResolved,
		TotalTax:      dec("12000"),
		Components: []itExt.TaxComponentResult{
			{TaxCode: "VAT", Rate: dec("10"), TaxAmount: dec("10000")},
			{TaxCode: "ENV", Rate: dec("2"), TaxAmount: dec("2000")},
		},
	})}

	tax, _, err := ResolveBasketTax(nil, taxSvc, testContext(),
		[]pricing.LineResult{pricedLine("a", "110000")})

	require.NoError(t, err)
	assert.True(t, tax.ByLineKey["a"].RateSnapshot.Equal(dec("0.12")),
		"10%% and 2%% must read as 0.12, got %s", tax.ByLineKey["a"].RateSnapshot)
}
