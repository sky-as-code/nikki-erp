package external

import (
	accmodels "github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
	itTax "github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/tax"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// TaxCalculationExtService is Sales' port onto accounting's tax determination and calculation. Sales
// computes what is owed commercially — price, discounts, promotions — and accounting computes what is
// owed legally on top: tax must not know about promotions, and Sales must not decide a rate.
//
// Simulate is deliberately omitted: it assembles a full trace of how the pipeline reached its
// answer, which is expensive and pointless on the path that prices an order. A screen needing the
// explanation calls accounting's own REST API.
type TaxCalculationExtService interface {
	// Calculate determines and computes tax for a whole document. Document-level rather than
	// per-line is load-bearing: a document-scoped rounding policy rounds the total once and
	// distributes the residual across lines, so calling this per line and summing gives a different
	// number. It has no business side effects and is safely repeatable.
	Calculate(ctx corectx.Context, request CalculationRequest) (*CalculateResult, error)

	// ReverseFull reverses an entire original charge from its frozen snapshot: it negates what was
	// charged rather than re-determining it, so a refund of a sale made under last year's rate
	// returns last year's tax.
	ReverseFull(ctx corectx.Context, request FullReversalRequest) (*ReverseResult, error)

	// ReversePartial reverses part of an original charge. The already-reversed figures travel in the
	// request because accounting stores no reversal state — Sales owns the running total.
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

// The determination outcomes, re-exported so Sales can branch on them. Unresolved means accounting
// could not decide what is owed; it must never be read as zero, so Sales refuses to confirm the
// order rather than storing a total it cannot defend.
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
// document default.
const (
	PriceInclusionInherit  = accmodels.PriceInclusionInherit
	PriceInclusionIncluded = accmodels.PriceInclusionIncluded
	PriceInclusionExcluded = accmodels.PriceInclusionExcluded
)
