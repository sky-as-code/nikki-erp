package services

import (
	"github.com/shopspring/decimal"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services/pricing"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// Resolving tax for a basket: the seam between what Sales computes and what Accounting computes.
//
// Sales decides the commercial base — the price, the discounts, the promotions — and Accounting
// decides the tax on it. BR-TAX-ESS-026 is explicit about the direction: "Sales owns discount logic.
// Tax must not decide a promotion or discount. Sales passes commercial_base_amount after applying
// business discounts. Tax uses this as the starting tax base."
//
// This file is where the two meet, and it exists as its own step rather than inside the pricing
// engine because the engine is pure (D-13). A pure engine can be replayed from its inputs, which is
// what makes a three-year-old sale reproducible; an engine that called another module could not be.

// hundred converts between Accounting's units and Sales'.
//
// Accounting stores a rate as a PERCENTAGE — a rate version holding 8 means 8% — and divides by a
// hundred internally (accounting/domain/services/tax/calculation.go). Sales stores
// tax_rate_snapshot as a FRACTION, documented as such in sales_order_line.json. The conversion
// happens once, here, at the boundary. Getting it wrong is a hundredfold error that no total would
// visibly flag as impossible, so it has its own test.
var hundred = decimal.NewFromInt(100)

// BasketTax is the resolved tax for one basket, ready for the pricing engine and for storage.
type BasketTax struct {
	// ByLineKey is what pricing.Context.TaxByLineKey takes.
	ByLineKey map[string]pricing.LineTax

	// Total is the document's tax, as Accounting rounded it. It is taken from Accounting rather than
	// summed from the lines: under a document-scoped rounding policy the two differ, and the document
	// figure is the one that was actually charged.
	Total decimal.Decimal

	// Snapshot is Accounting's immutable record of how it reached this answer. Sales stores it —
	// Accounting keeps no copy and holds no foreign key into Sales (BR-TAX-ESS-030) — and a partial
	// return later hands it straight back as PartialReversalRequest.OriginalSnapshot.
	Snapshot itExt.TaxSnapshot
}

// ZeroBasketTax is the answer when no tax applies to anything.
//
// It is a real answer rather than an absence: every line is taxed, at zero. That distinction matters
// because "taxed at zero" and "tax undetermined" are different facts, and only the first may be
// stored on a confirmed order.
func ZeroBasketTax(lines []pricing.LineResult) BasketTax {
	byLineKey := make(map[string]pricing.LineTax, len(lines))
	for _, line := range lines {
		byLineKey[line.Key] = pricing.LineTax{
			RateSnapshot: decimal.Zero,
			Amount:       decimal.Zero,
		}
	}
	return BasketTax{ByLineKey: byLineKey, Total: decimal.Zero}
}

// TaxRequestContext is what Sales knows about the sale that Accounting needs.
type TaxRequestContext struct {
	OrgId string

	// TaxDate is the date the sale legally occurred, formatted YYYY-MM-DD. It is MANDATORY and must
	// come from the caller: BR-TAX-ESS-SUP-020 deleted the server-clock fallback, because a request
	// that forgot the date would otherwise be priced against whatever configuration happened to be
	// effective when it was processed rather than when the sale happened.
	TaxDate string

	CurrencyCode string

	// TaxCode names the tax to apply to every line, from the sales_org_settings default.
	//
	// One code for the whole basket is a deliberate interim. The per-product association that doc 3
	// specifies — product_template_sales_taxes and effective_sales_tax_ids — does not exist in
	// essential or inventory yet, so nothing can say that this product is 8% and that one 5%. When it
	// is built, this field becomes a per-line lookup and nothing else here changes shape.
	TaxCode string

	// PriceMode is the document default a tax whose own inclusion mode is "inherit" resolves against.
	// Sales sells tax-inclusive, so this is normally itExt.PriceInclusionIncluded.
	PriceMode itExt.PriceInclusionMode

	// BusinessChannelCode and OutletReference are opaque context Accounting carries for audit and
	// never resolves back against Sales.
	BusinessChannelCode string
	OutletReference     string
}

// ResolveBasketTax asks Accounting for the tax on a priced basket.
//
// **It fails closed.** When Accounting cannot determine the tax — no rate version in force on the tax
// date, two that both are, no rounding policy — it answers `unresolved`, and BR-TAX-ESS forbids
// reading that as zero. This returns ClientErrors, which the caller must surface as a 400: the
// configuration is wrong and an administrator can fix it. Storing a zero-tax order instead would
// under-charge VAT silently, and the error would only surface at a tax audit.
//
// That reverses what D-38 established. Its fallback-to-zero was correct while zero was the true
// answer — there was no tax master to consult — and is not correct now that one exists and is
// seeded. The reversal is deliberate and recorded in the plan.
//
// A nil port means this deployment has no accounting module: every line is taxed at zero, which is
// accurate rather than optimistic.
func ResolveBasketTax(
	ctx corectx.Context,
	taxSvc itExt.TaxCalculationExtService,
	context TaxRequestContext,
	lines []pricing.LineResult,
) (*BasketTax, *ft.ClientErrors, error) {
	if taxSvc == nil || context.TaxCode == "" || len(lines) == 0 {
		zero := ZeroBasketTax(lines)
		return &zero, nil, nil
	}

	request := buildCalculationRequest(context, lines)
	result, err := taxSvc.Calculate(ctx, request)
	if err != nil {
		// A transport or database failure inside Accounting. Not the caller's fault and not
		// something a form can fix, so it propagates as a 500 rather than a validation error.
		return nil, nil, err
	}
	if result == nil || !result.HasData {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation("tax", "sales_order.tax_unavailable",
			"the tax service returned no result; the order cannot be priced until it does"))
		return nil, vErrs, nil
	}

	if vErrs := assertTaxResolved(result.Data); vErrs != nil {
		return nil, vErrs, nil
	}
	return fromCalculationResult(result.Data), nil, nil
}

// assertTaxResolved refuses a document Accounting could not determine.
//
// The document status is checked as well as each line's, because Accounting marks the whole document
// unresolved if any line is — a partial answer would let a caller store a total that silently omits
// one line's tax.
func assertTaxResolved(result itExt.CalculationResult) *ft.ClientErrors {
	if result.Status != itExt.DeterminationUnresolved {
		return nil
	}

	vErrs := ft.NewClientErrors()
	for _, line := range result.Lines {
		if line.Status != itExt.DeterminationUnresolved {
			continue
		}
		// The error code names WHY — a missing rate reads differently from an ambiguous one — so the
		// administrator is pointed at the configuration to fix rather than told only that it failed.
		reason := line.ErrorCode
		if reason == "" {
			reason = "unknown"
		}
		vErrs.Append(*ft.NewBusinessViolation("tax", "sales_order.tax_unresolved",
			"tax could not be determined for line '"+line.LineReference+"' ("+reason+"); "+
				"an undetermined tax must not be recorded as zero"))
	}
	if vErrs.Count() == 0 {
		vErrs.Append(*ft.NewBusinessViolation("tax", "sales_order.tax_unresolved",
			"tax could not be determined for this order; "+
				"an undetermined tax must not be recorded as zero"))
	}
	return vErrs
}

// buildCalculationRequest turns priced lines into one document-level tax request.
//
// One request for the whole basket, never one per line: a document-scoped rounding policy rounds the
// total once and distributes the residual (BR-TAX-ESS-022), and per-line calls summed afterwards
// produce a different number that no policy asked for.
func buildCalculationRequest(
	context TaxRequestContext, lines []pricing.LineResult,
) itExt.CalculationRequest {
	taxLines := make([]itExt.CalculationLine, 0, len(lines))
	for _, line := range lines {
		taxLines = append(taxLines, itExt.CalculationLine{
			LineReference:    line.Key,
			ProductReference: line.ProductVariantId,
			Quantity:         line.Quantity,
			UomId:            line.UomId,
			UnitPrice:        line.EffectiveUnitPrice,
			DiscountAmount:   line.DiscountAmount,

			// The taxable base is the line's NET — gross less every discount and promotion already
			// applied. Passing gross would tax the customer on money they did not pay.
			CommercialBaseAmount: line.NetAmount,

			// The tax to apply. Accounting validates it, resolves its version for the tax date and
			// computes; under simple mode it runs no determination rules of its own.
			CandidateTaxIds: []string{context.TaxCode},
		})
	}

	return itExt.CalculationRequest{
		OrgId:               context.OrgId,
		OperationType:       itExt.OperationSale,
		TaxDate:             context.TaxDate,
		CurrencyCode:        context.CurrencyCode,
		PriceMode:           context.PriceMode,
		BusinessChannelCode: context.BusinessChannelCode,
		OutletReference:     context.OutletReference,
		Lines:               taxLines,
	}
}

// fromCalculationResult maps Accounting's answer into the engine's units.
func fromCalculationResult(result itExt.CalculationResult) *BasketTax {
	byLineKey := make(map[string]pricing.LineTax, len(result.Lines))
	for _, line := range result.Lines {
		byLineKey[line.LineReference] = pricing.LineTax{
			RateSnapshot: effectiveRateFraction(line),
			Amount:       line.TotalTax,
		}
	}
	return &BasketTax{
		ByLineKey: byLineKey,
		Total:     result.TotalTax,
		Snapshot:  result.Snapshot,
	}
}

// effectiveRateFraction is the line's overall rate, as a FRACTION.
//
// Summing the component rates is right for the ordinary case — one VAT, or several taxes each on the
// same base — and approximate for a compound tax, where a later component applies to a base that
// already includes an earlier one. The exact figures are all preserved per component in the snapshot;
// this single number exists so a line can be READ without unpacking it, which is what
// tax_rate_snapshot is for.
func effectiveRateFraction(line itExt.TaxLineResult) decimal.Decimal {
	total := decimal.Zero
	for _, component := range line.Components {
		total = total.Add(component.Rate)
	}
	if total.IsZero() {
		return decimal.Zero
	}
	// Percentage to fraction: 10 becomes 0.1. See the note on `hundred`.
	return total.Div(hundred)
}
