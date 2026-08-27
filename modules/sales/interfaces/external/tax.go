package external

import (
	accmodels "github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
	itTax "github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/tax"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// TaxCalculationExtService is Sales' port onto Accounting's tax determination and calculation.
//
// Sales computes what is owed commercially — the price, the discounts, the promotions — and
// Accounting computes what is owed legally on top of that. The split is not arbitrary: BR-TAX-ESS-026
// and TAX-INV-17 both say outright that Tax must not know about promotions, and Sales must not decide
// a rate. Each module would get the other's job wrong.
//
// Note what is NOT here: Simulate. It exists upstream and is deliberately left out, because it is the
// tax administrator's debugging surface — it assembles a full trace of how the pipeline reached its
// answer, which is expensive and pointless on the path that prices an order. A screen that needs the
// explanation calls Accounting's own REST API; Sales does not proxy it.
//
// Reversal IS here although nothing calls it yet. Phase 6 returns need it, and the shape of a
// reversal request constrains what a sale must have stored — declaring it now is what makes that
// obligation visible while the storing code is being written, rather than after.
type TaxCalculationExtService interface {
	// Calculate determines and computes tax for a whole document.
	//
	// Document-level rather than per-line, and that is load-bearing: a document-scoped rounding
	// policy rounds the total once and distributes the residual across lines (BR-TAX-ESS-022). Calling
	// this per line and summing would produce a different number — one no rounding policy asked for.
	//
	// It has no business side effects and is safely repeatable, which is what lets a draft order
	// recalculate on every edit.
	Calculate(ctx corectx.Context, request CalculationRequest) (*CalculateResult, error)

	// ReverseFull reverses an entire original charge from its frozen snapshot.
	//
	// It negates what was charged rather than re-determining it: a refund of a sale made under last
	// year's rate returns last year's tax, whatever the rate is today.
	ReverseFull(ctx corectx.Context, request FullReversalRequest) (*ReverseResult, error)

	// ReversePartial reverses part of an original charge.
	//
	// The already-reversed figures travel in the request because Accounting stores no reversal state
	// (BR-TAX-ESS-SUP-025) — Sales owns the running total of what has been refunded so far.
	ReversePartial(ctx corectx.Context, request PartialReversalRequest) (*ReverseResult, error)
}

// The request and result types, re-exported so Sales names them without importing accounting outside
// infra/external.
type (
	CalculationRequest       = itTax.CalculationRequest
	CalculationLine          = itTax.CalculationLine
	TaxPartyContext          = itTax.TaxPartyContext
	TaxRegistration          = itTax.TaxRegistration
	CalculateResult          = itTax.CalculateResult
	CalculationResult        = itTax.CalculationResult
	TaxLineResult            = itTax.LineResult
	TaxComponentResult       = itTax.ComponentResult
	TaxSnapshot              = itTax.Snapshot
	FullReversalRequest      = itTax.FullReversalRequest
	PartialReversalRequest   = itTax.PartialReversalRequest
	ReversalComponentRequest = itTax.ReversalComponentRequest
	ReverseResult            = itTax.ReverseResult
	ReversalResult           = itTax.ReversalResult

	// The enum types, so a caller can declare a field of one without importing accounting.
	DeterminationStatus = accmodels.DeterminationStatus
	OperationType       = accmodels.OperationType
	PriceInclusionMode  = accmodels.PriceInclusionMode
)

// The determination outcomes, re-exported so Sales can branch on them.
//
// Unresolved is the one that matters to Sales: it means Accounting could not decide what is owed —
// a missing rate version, an ambiguous one, no rounding policy in force. BR-TAX-ESS forbids reading
// it as zero, so Sales refuses to confirm the order rather than storing a total it cannot defend.
const (
	DeterminationResolved        = accmodels.DeterminationResolved
	DeterminationNoTaxApplicable = accmodels.DeterminationNoTaxApplicable
	DeterminationUnresolved      = accmodels.DeterminationUnresolved
)

// The operation types Sales uses. Purchase values exist upstream as a reserved contract and are
// rejected by the engine, so they are not re-exported here.
const (
	OperationSale       = accmodels.OperationSale
	OperationSaleRefund = accmodels.OperationSaleRefund
)

// The price inclusion modes. Sales sells tax-inclusive, so PriceInclusionIncluded is the expected
// document default; the constants are re-exported so a caller names one without importing accounting.
const (
	PriceInclusionInherit  = accmodels.PriceInclusionInherit
	PriceInclusionIncluded = accmodels.PriceInclusionIncluded
	PriceInclusionExcluded = accmodels.PriceInclusionExcluded
)
