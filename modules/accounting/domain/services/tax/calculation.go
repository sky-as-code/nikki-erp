package tax

import (
	"github.com/shopspring/decimal"

	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
)

// workingScale is the precision intermediate results keep.
//
// Rounding happens once, where the policy says (BR-TAX-ESS-044). Everything before that keeps far
// more precision than any currency needs, so that a division in a compound chain does not shed
// fractions that the final rounding would otherwise have accounted for.
const workingScale = 24

// hundred is the divisor turning a percentage into a fraction.
//
// Rates are stored as percentages — 8 means 8% — because that is how law and invoices state them
// (BR-TAX-ESS-SUP-006). The division lives here rather than at the call sites so that no caller can
// forget it and charge a customer 800%.
var hundred = decimal.NewFromInt(100)

// ComponentSpec is one tax to compute against a line.
//
// It carries the resolved configuration rather than ids, because by the time the calculator runs,
// determination has already chosen the definition version and rate version, and the calculator must
// not be able to reach back and choose differently.
type ComponentSpec struct {
	TaxId           string
	TaxCode         string
	Sequence        int32
	CalculationType models.CalculationType
	Treatment       models.TaxTreatment
	InclusionMode   models.PriceInclusionMode

	// Rate is a percentage for percentage and division taxes; FixedAmount and Quantity are used by
	// fixed ones. Which of them is populated follows from CalculationType, and configuration
	// validation has already enforced that only the right ones are set.
	Rate        decimal.Decimal
	FixedAmount decimal.Decimal

	// Quantity is already converted into the unit the fixed rate is quoted in. The conversion is
	// Essential's (BR-TAX-ESS-SUP-014); by this point it has happened or the line was rejected.
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

	// TaxableBase is the base this component was actually computed on, after any compound
	// adjustment. It is reported rather than recomputed downstream because a compound chain makes
	// it genuinely different per component, and an invoice has to be able to show its working.
	TaxableBase decimal.Decimal

	// Amount is unrounded. Rounding is a separate, policy-driven step.
	Amount decimal.Decimal
}

// LineInput is one line entering calculation.
type LineInput struct {
	LineReference string

	// CommercialBase is the line's pre-tax amount as Sales computed it, already net of discount.
	// Tax takes it as given: discounts and promotions are Sales' business, and Tax deciding them
	// too is how two modules end up disagreeing about a total (BR-TAX-ESS-026, TAX-INV-17).
	CommercialBase decimal.Decimal

	// PriceMode is the document's default, used by any component whose inclusion mode is "inherit".
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
//
// Components run in the order given, which determination has already sorted by sequence. The order
// is load-bearing whenever a component affects a later one's base, which is exactly what a compound
// tax is (BR-TAX-ESS-019).
func CalculateLine(line LineInput) LineAmounts {
	base := effectiveExcludedBase(line)

	// compounded accumulates the amounts of components flagged to feed later bases. It is separate
	// from the running total because a component only picks it up when its own
	// BaseAffectedByPrevious says so — the two flags are independent, so a tax can feed others
	// without being fed itself.
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

// effectiveExcludedBase turns the line's commercial base into a tax-excluded one.
//
// When the price the customer sees already contains tax, the pre-tax base has to be extracted
// before anything can be computed on it — charging tax on a tax-inclusive price would tax the tax
// (BR-TAX-ESS-016). Extraction uses the combined rate of every component that participates in the
// inclusive price, because with two inclusive taxes the gross contains both.
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

// isIncluded resolves a component's price inclusion against the document default.
//
// A tax that states included or excluded wins; one saying inherit defers to the request. That is
// what lets a single VAT definition serve tax-inclusive retail and tax-exclusive wholesale without
// duplicating it (BR-TAX-ESS-017).
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
		// The rate expresses tax as a share of the tax-INCLUSIVE total rather than of the base, so
		// the total is recovered first and the tax is the difference (BR-TAX-ESS-SUP-013):
		//   total = base / (1 - rate/100);  tax = total - base
		// A rate of 100% would divide by zero — a tax that is the entire price is not expressible
		// this way, and returning zero keeps the arithmetic total rather than panicking mid-order.
		divisor := decimal.NewFromInt(1).Sub(spec.Rate.DivRound(hundred, workingScale))
		if !divisor.IsPositive() {
			return decimal.Zero
		}
		return base.DivRound(divisor, workingScale).Sub(base)

	case models.CalculationFixed:
		// tax = quantity x amount per unit, the quantity already converted into the rate's unit.
		return spec.Quantity.Mul(spec.FixedAmount)

	case models.CalculationGroup:
		// A group has no arithmetic of its own; its components are expanded into the line before
		// this point, so it contributes nothing directly.
		return decimal.Zero

	case models.CalculationNone:
		// Legal semantics only — an exemption still needs a code on the invoice, but produces no
		// amount (BR-TAX-ESS-SUP-015).
		return decimal.Zero
	}
	return decimal.Zero
}
