package tax

import (
	"github.com/shopspring/decimal"

	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
)

// workingScale is the precision intermediate results keep. Rounding happens once, later, where the
// policy says; until then keep far more precision than any currency needs so a division in a
// compound chain does not shed fractions the final rounding would have accounted for.
const workingScale = 24

// hundred converts a percentage to a fraction. Rates are stored as percentages (8 means 8%), as law
// and invoices state them; this package is the single place that divides by 100, so no caller can
// forget it and charge 800%.
var hundred = decimal.NewFromInt(100)

// ComponentSpec is one tax to compute against a line. It carries configuration already resolved by
// determination, which runs first; the calculator must not choose a different version.
type ComponentSpec struct {
	TaxId           string
	TaxCode         string
	Sequence        int32
	CalculationType models.CalculationType
	Treatment       models.TaxTreatment
	InclusionMode   models.PriceInclusionMode

	// Rate is a percentage (not a fraction) for percentage and division taxes; fixed taxes use
	// FixedAmount and Quantity instead. CalculationType decides which is populated.
	Rate        decimal.Decimal
	FixedAmount decimal.Decimal

	// Quantity is already converted into the unit the fixed rate is quoted in; by this point the
	// conversion has happened or the line was rejected.
	Quantity decimal.Decimal

	// AffectSubsequentBase adds this tax's amount into the base of later components.
	AffectSubsequentBase bool
	// BaseAffectedByPrevious lets this tax's own base pick up earlier components' amounts.
	BaseAffectedByPrevious bool
}

// ComponentAmount is what one component came to.
type ComponentAmount struct {
	TaxId     string
	TaxCode   string
	Sequence  int32
	Treatment models.TaxTreatment

	// TaxableBase is the base this component was computed on, after any compound adjustment. It is
	// reported rather than recomputed downstream: a compound chain makes it differ per component.
	TaxableBase decimal.Decimal

	// Amount is unrounded. Rounding is a separate, policy-driven step.
	Amount decimal.Decimal
}

// LineInput is one line entering calculation.
type LineInput struct {
	LineReference string

	// CommercialBase is the line's pre-tax amount from Sales, already net of discount. Tax takes it
	// as given; recomputing discounts here would let the two modules disagree on a total.
	CommercialBase decimal.Decimal

	// PriceMode is the document's default tax-inclusive/exclusive mode, used by any component whose
	// inclusion mode is "inherit".
	PriceMode models.PriceInclusionMode

	Components []ComponentSpec
}

// LineAmounts is the calculated result for one line.
type LineAmounts struct {
	LineReference string
	TotalExcluded decimal.Decimal
	TotalTax      decimal.Decimal
	TotalIncluded decimal.Decimal
	Components    []ComponentAmount
}

// CalculateLine computes every component of one line, honouring compounding and price inclusion.
// Components run in the given order, which determination has already sorted by sequence; that order
// is load-bearing whenever one component feeds a later one's base. Results are unrounded.
func CalculateLine(line LineInput) LineAmounts {
	base := effectiveExcludedBase(line)

	// compounded accumulates amounts of components flagged to feed later bases. It is separate from
	// the running total because the two flags are independent: a tax can feed others without being
	// fed itself.
	compounded := decimal.Zero
	totalTax := decimal.Zero
	amounts := make([]ComponentAmount, 0, len(line.Components))

	for _, spec := range line.Components {
		componentBase := base
		if spec.BaseAffectedByPrevious {
			componentBase = componentBase.Add(compounded)
		}

		amount := computeComponent(spec, componentBase)

		amounts = append(amounts, ComponentAmount{
			TaxId:       spec.TaxId,
			TaxCode:     spec.TaxCode,
			Sequence:    spec.Sequence,
			Treatment:   spec.Treatment,
			TaxableBase: componentBase,
			Amount:      amount,
		})

		totalTax = totalTax.Add(amount)
		if spec.AffectSubsequentBase {
			compounded = compounded.Add(amount)
		}
	}

	return LineAmounts{
		LineReference: line.LineReference,
		TotalExcluded: base,
		TotalTax:      totalTax,
		TotalIncluded: base.Add(totalTax),
		Components:    amounts,
	}
}

// effectiveExcludedBase turns the line's commercial base into a tax-excluded one. When the price
// already contains tax, the pre-tax base must be extracted first, or the tax is taxed. Extraction
// uses the combined rate of every inclusive component, since the gross contains all of them.
func effectiveExcludedBase(line LineInput) decimal.Decimal {
	inclusiveRate := decimal.Zero
	for _, spec := range line.Components {
		if !isIncluded(spec, line.PriceMode) {
			continue
		}
		// Only rate-driven taxes can be extracted proportionally. A fixed amount inside an
		// inclusive price is subtracted directly, below.
		if spec.CalculationType == models.CalculationPercentage {
			inclusiveRate = inclusiveRate.Add(spec.Rate)
		}
	}

	base := line.CommercialBase
	for _, spec := range line.Components {
		if isIncluded(spec, line.PriceMode) && spec.CalculationType == models.CalculationFixed {
			base = base.Sub(spec.FixedAmount.Mul(spec.Quantity))
		}
	}

	if inclusiveRate.IsZero() {
		return base
	}
	// base = gross / (1 + rate/100)
	divisor := decimal.NewFromInt(1).Add(inclusiveRate.Div(hundred))
	if divisor.IsZero() {
		return base
	}
	return base.DivRound(divisor, workingScale)
}

// isIncluded resolves a component's price inclusion against the document default: an explicit
// included or excluded wins, inherit defers to the document mode. This lets one VAT definition serve
// both tax-inclusive retail and tax-exclusive wholesale.
func isIncluded(spec ComponentSpec, documentMode models.PriceInclusionMode) bool {
	switch spec.InclusionMode {
	case models.PriceInclusionIncluded:
		return true
	case models.PriceInclusionExcluded:
		return false
	default:
		return documentMode == models.PriceInclusionIncluded
	}
}

// computeComponent applies the arithmetic for one calculation type.
func computeComponent(spec ComponentSpec, base decimal.Decimal) decimal.Decimal {
	switch spec.CalculationType {
	case models.CalculationPercentage:
		// tax = base x rate / 100
		return base.Mul(spec.Rate).DivRound(hundred, workingScale)

	case models.CalculationDivision:
		// Here the rate is a share of the tax-INCLUSIVE total, not of the base, so recover the total
		// first and take the difference:
		//   total = base / (1 - rate/100);  tax = total - base
		// A rate of 100% would divide by zero, so return zero rather than panic mid-order.
		divisor := decimal.NewFromInt(1).Sub(spec.Rate.DivRound(hundred, workingScale))
		if !divisor.IsPositive() {
			return decimal.Zero
		}
		return base.DivRound(divisor, workingScale).Sub(base)

	case models.CalculationFixed:
		// tax = quantity x amount per unit, the quantity already converted into the rate's unit.
		return spec.Quantity.Mul(spec.FixedAmount)

	case models.CalculationGroup:
		// A group's components are expanded into the line earlier, so it contributes nothing here.
		return decimal.Zero

	case models.CalculationNone:
		// Legal semantics only: an exemption needs a code on the invoice but produces no amount.
		return decimal.Zero
	}
	return decimal.Zero
}
