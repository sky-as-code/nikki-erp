package tax

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
)

func dec(value string) decimal.Decimal {
	parsed, err := decimal.NewFromString(value)
	if err != nil {
		panic(err)
	}
	return parsed
}

func percentageSpec(code string, rate string, sequence int32) ComponentSpec {
	return ComponentSpec{
		TaxId:           code,
		TaxCode:         code,
		Sequence:        sequence,
		CalculationType: models.CalculationPercentage,
		Treatment:       models.TaxTreatmentTaxable,
		InclusionMode:   models.PriceInclusionInherit,
		Rate:            dec(rate),
	}
}

// AC-TAX-07 and AC-TAX-12: percentage tax on a tax-excluded price.
func TestPercentageTaxExcluded(t *testing.T) {
	result := CalculateLine(LineInput{
		LineReference:  "L1",
		CommercialBase: dec("100000"),
		PriceMode:      models.PriceInclusionExcluded,
		Components:     []ComponentSpec{percentageSpec("VAT10", "10", 1)},
	})

	assert.True(t, result.TotalExcluded.Equal(dec("100000")), "excluded = %s", result.TotalExcluded)
	assert.True(t, result.TotalTax.Equal(dec("10000")), "tax = %s", result.TotalTax)
	assert.True(t, result.TotalIncluded.Equal(dec("110000")), "included = %s", result.TotalIncluded)
}

// AC-TAX-11: a tax-inclusive price has the tax extracted, not added on top.
//
// The worked example from BR-TAX-ESS-016: 110,000 at 10% inclusive is 100,000 base and 10,000 tax,
// never 110,000 base and 11,000 tax.
func TestPercentageTaxIncluded(t *testing.T) {
	result := CalculateLine(LineInput{
		LineReference:  "L1",
		CommercialBase: dec("110000"),
		PriceMode:      models.PriceInclusionIncluded,
		Components:     []ComponentSpec{percentageSpec("VAT10", "10", 1)},
	})

	assert.True(t, result.TotalExcluded.Equal(dec("100000")), "excluded = %s", result.TotalExcluded)
	assert.True(t, result.TotalTax.Equal(dec("10000")), "tax = %s", result.TotalTax)
	assert.True(t, result.TotalIncluded.Equal(dec("110000")), "included = %s", result.TotalIncluded)
}

// BR-TAX-ESS-017: the tax's own inclusion mode overrides the document's.
func TestPriceInclusionPrecedence(t *testing.T) {
	spec := percentageSpec("VAT10", "10", 1)
	spec.InclusionMode = models.PriceInclusionExcluded

	// The document says inclusive, but the tax insists it is excluded, so nothing is extracted.
	result := CalculateLine(LineInput{
		LineReference:  "L1",
		CommercialBase: dec("100000"),
		PriceMode:      models.PriceInclusionIncluded,
		Components:     []ComponentSpec{spec},
	})

	assert.True(t, result.TotalExcluded.Equal(dec("100000")))
	assert.True(t, result.TotalTax.Equal(dec("10000")))
}

// BR-TAX-ESS-SUP-013: division tax, with the doc's worked example of base 180 at 10% giving 20.
func TestDivisionTax(t *testing.T) {
	spec := percentageSpec("DIV10", "10", 1)
	spec.CalculationType = models.CalculationDivision

	result := CalculateLine(LineInput{
		LineReference:  "L1",
		CommercialBase: dec("180"),
		PriceMode:      models.PriceInclusionExcluded,
		Components:     []ComponentSpec{spec},
	})

	assert.True(t, result.TotalTax.Equal(dec("20")), "tax = %s", result.TotalTax)
	assert.True(t, result.TotalIncluded.Equal(dec("200")), "included = %s", result.TotalIncluded)
}

// AC-TAX-08: fixed tax multiplies a converted quantity by an amount per unit.
func TestFixedTax(t *testing.T) {
	result := CalculateLine(LineInput{
		LineReference:  "L1",
		CommercialBase: dec("50000"),
		PriceMode:      models.PriceInclusionExcluded,
		Components: []ComponentSpec{{
			TaxId:           "ENV",
			TaxCode:         "ENV",
			Sequence:        1,
			CalculationType: models.CalculationFixed,
			Treatment:       models.TaxTreatmentTaxable,
			InclusionMode:   models.PriceInclusionInherit,
			FixedAmount:     dec("3000"),
			Quantity:        dec("12"),
		}},
	})

	assert.True(t, result.TotalTax.Equal(dec("36000")), "tax = %s", result.TotalTax)
}

// AC-TAX-09 and AC-TAX-10: a compound tax feeds the base of the one after it.
//
// The worked example from BR-TAX-ESS-019: base 100, A at 10% gives 10, and B at 5% is computed on
// 110 rather than 100, giving 5.5.
func TestCompoundTaxAffectsSubsequentBase(t *testing.T) {
	first := percentageSpec("A", "10", 1)
	first.AffectSubsequentBase = true

	second := percentageSpec("B", "5", 2)
	second.BaseAffectedByPrevious = true

	result := CalculateLine(LineInput{
		LineReference:  "L1",
		CommercialBase: dec("100"),
		PriceMode:      models.PriceInclusionExcluded,
		Components:     []ComponentSpec{first, second},
	})

	assert.True(t, result.Components[0].Amount.Equal(dec("10")),
		"first = %s", result.Components[0].Amount)
	assert.True(t, result.Components[1].TaxableBase.Equal(dec("110")),
		"second base = %s", result.Components[1].TaxableBase)
	assert.True(t, result.Components[1].Amount.Equal(dec("5.5")),
		"second = %s", result.Components[1].Amount)
	assert.True(t, result.TotalTax.Equal(dec("15.5")), "total = %s", result.TotalTax)
}

// The two compound flags are independent: a tax may feed later bases without picking up earlier
// ones. Without BaseAffectedByPrevious the second component stays on the original base.
func TestCompoundFlagsAreIndependent(t *testing.T) {
	first := percentageSpec("A", "10", 1)
	first.AffectSubsequentBase = true
	second := percentageSpec("B", "5", 2) // BaseAffectedByPrevious deliberately false

	result := CalculateLine(LineInput{
		LineReference:  "L1",
		CommercialBase: dec("100"),
		PriceMode:      models.PriceInclusionExcluded,
		Components:     []ComponentSpec{first, second},
	})

	assert.True(t, result.Components[1].TaxableBase.Equal(dec("100")),
		"second base = %s", result.Components[1].TaxableBase)
	assert.True(t, result.Components[1].Amount.Equal(dec("5")))
}

// AC-TAX-15: zero-rated is a real tax computed at 0%, and keeps its treatment on the result.
func TestZeroRatedIsCalculatedNotSkipped(t *testing.T) {
	spec := percentageSpec("VAT0", "0", 1)
	spec.Treatment = models.TaxTreatmentZeroRated

	result := CalculateLine(LineInput{
		LineReference:  "L1",
		CommercialBase: dec("100000"),
		PriceMode:      models.PriceInclusionExcluded,
		Components:     []ComponentSpec{spec},
	})

	assert.Len(t, result.Components, 1, "the component must appear even though it charges nothing")
	assert.True(t, result.TotalTax.IsZero())
	assert.Equal(t, models.TaxTreatmentZeroRated, result.Components[0].Treatment)
}

// BR-TAX-ESS-SUP-015: a "none" calculation produces no amount but still reports its treatment, so
// that an exemption keeps its legal identity on the document.
func TestNoneCalculationProducesNoAmount(t *testing.T) {
	result := CalculateLine(LineInput{
		LineReference:  "L1",
		CommercialBase: dec("100000"),
		PriceMode:      models.PriceInclusionExcluded,
		Components: []ComponentSpec{{
			TaxId:           "EXEMPT",
			TaxCode:         "EXEMPT",
			Sequence:        1,
			CalculationType: models.CalculationNone,
			Treatment:       models.TaxTreatmentExempt,
			InclusionMode:   models.PriceInclusionInherit,
		}},
	})

	assert.True(t, result.TotalTax.IsZero())
	assert.Equal(t, models.TaxTreatmentExempt, result.Components[0].Treatment)
	assert.True(t, result.TotalExcluded.Equal(dec("100000")))
}

// AC-TAX-34 in miniature: two different rates on one document, computed per line.
func TestDifferentRatesAcrossLines(t *testing.T) {
	reduced := CalculateLine(LineInput{
		LineReference:  "L1",
		CommercialBase: dec("100000"),
		PriceMode:      models.PriceInclusionExcluded,
		Components:     []ComponentSpec{percentageSpec("VAT8", "8", 1)},
	})
	standard := CalculateLine(LineInput{
		LineReference:  "L2",
		CommercialBase: dec("100000"),
		PriceMode:      models.PriceInclusionExcluded,
		Components:     []ComponentSpec{percentageSpec("VAT10", "10", 1)},
	})

	assert.True(t, reduced.TotalTax.Equal(dec("8000")))
	assert.True(t, standard.TotalTax.Equal(dec("10000")))
}

// TAX-INV-20 / AC-TAX-21: the same input yields the same result, and the arithmetic is decimal.
//
// The rate 8.5% on 1,234,567 is chosen because it has no exact binary representation: a float
// implementation drifts here, a decimal one does not.
func TestCalculationIsDeterministicAndExact(t *testing.T) {
	line := LineInput{
		LineReference:  "L1",
		CommercialBase: dec("1234567"),
		PriceMode:      models.PriceInclusionExcluded,
		Components:     []ComponentSpec{percentageSpec("VAT85", "8.5", 1)},
	}

	first := CalculateLine(line)
	second := CalculateLine(line)

	assert.True(t, first.TotalTax.Equal(second.TotalTax))
	assert.True(t, first.TotalTax.Equal(dec("104938.195")), "tax = %s", first.TotalTax)
}
