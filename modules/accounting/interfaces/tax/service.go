package tax

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

type CalculateResult = dyn.OpResult[CalculationResult]
type SimulateResult = dyn.OpResult[SimulationResult]
type ReverseResult = dyn.OpResult[ReversalResult]

// TaxCalculationAppService is the capability other modules consume.
//
// A consuming module declares its own port in interfaces/external and binds this once in
// infra/external, rather than importing it from a domain or application layer — the same shape
// Purchase and Sales use for UoM. That is what keeps the eventual split into separate processes a
// change to one file.
type TaxCalculationAppService interface {
	// Calculate determines and computes tax for a whole document.
	//
	// It has no business side effects whatsoever: no invoice, no posting, no stock movement, no
	// change to tax master data (BR-TAX-ESS-046, AC-TAX-35). Calling it twice with the same input
	// produces the same answer and changes nothing, which is what lets a draft order recalculate on
	// every edit.
	Calculate(ctx corectx.Context, request CalculationRequest) (*CalculateResult, error)

	// Simulate runs the same pipeline and additionally returns the trace of how it got there.
	//
	// Separate from Calculate because the explanation is expensive to assemble and pointless on the
	// hot path: an order being priced needs the number, a tax administrator debugging a rule needs
	// the reasoning.
	Simulate(ctx corectx.Context, request CalculationRequest) (*SimulateResult, error)

	// ReverseFull reverses an entire original charge from its frozen snapshot.
	ReverseFull(ctx corectx.Context, request FullReversalRequest) (*ReverseResult, error)

	// ReversePartial reverses part of an original charge.
	ReversePartial(ctx corectx.Context, request PartialReversalRequest) (*ReverseResult, error)
}

// TaxCalculationDomainService is the same capability without the permission checks.
//
// Authorization happens in app/ and nowhere else, so the domain service is what the application
// service calls once it has decided the caller may proceed.
type TaxCalculationDomainService interface {
	Calculate(ctx corectx.Context, request CalculationRequest) (*CalculateResult, error)
	Simulate(ctx corectx.Context, request CalculationRequest) (*SimulateResult, error)
	ReverseFull(ctx corectx.Context, request FullReversalRequest) (*ReverseResult, error)
	ReversePartial(ctx corectx.Context, request PartialReversalRequest) (*ReverseResult, error)
}

// SimulationResult is a calculation plus the explanation of how it was reached.
//
// The trace is what BR-TAX-ESS-051 requires the Tax Simulator to display: matched rules, then the
// mapping, then the applicable taxes, their rates, their components and the final figure.
type SimulationResult struct {
	Calculation CalculationResult
	Trace       []TraceStep
}

// TraceStep is one stage of the pipeline, as it happened.
type TraceStep struct {
	// Stage names the pipeline step: candidate_taxes, rule_evaluation, mapping, override,
	// version_resolution, calculation, rounding.
	Stage string

	// Detail is a human-readable account of what the step did, already localized by the caller's
	// locale where it names configuration.
	Detail string

	// TaxIds is the tax set as it stood after this step, so a reader can see exactly where a tax
	// entered or left.
	TaxIds []string

	// RuleIds names the rules that fired at this step, when the step is rule evaluation.
	RuleIds []string
}

// FullReversalRequest reverses an original charge in its entirety.
type FullReversalRequest struct {
	OrgId string

	// OriginalSnapshot is the frozen snapshot the caller stored at the time of sale. It is supplied
	// rather than looked up because Tax keeps no copy: the transaction and its snapshot belong to
	// the consuming module (TAX-SUP-INV-01).
	OriginalSnapshot Snapshot

	// TaxDate is the refund's own date, recorded on the reversal snapshot. It does NOT re-resolve
	// the rate: a full reversal negates what was charged, whatever the rate is today.
	TaxDate string

	Components []ReversalComponentRequest
}

// PartialReversalRequest reverses part of an original charge.
type PartialReversalRequest struct {
	OrgId            string
	OriginalSnapshot Snapshot
	TaxDate          string

	RoundingPolicyCode string

	Components []ReversalComponentRequest
}

// ReversalComponentRequest is the caller's reversal state for one original component.
//
// The already-reversed figures come from the caller because Tax stores no reversal state of its own
// (BR-TAX-ESS-SUP-025). The alternative would require Tax to track every downstream transaction,
// which is the dependency the whole module is arranged to avoid.
type ReversalComponentRequest struct {
	OriginalComponentReference string
	OriginalReversibleBasis    string
	OriginalTaxAmount          string
	AlreadyReversedBasis       string
	AlreadyReversedTaxAmount   string
	RequestedReversalBasis     string

	// IsFinalReversal marks the last refund against this component. It absorbs any rounding
	// residual, which is what makes the reversals sum to exactly the original charge rather than
	// approximately (BR-TAX-ESS-033).
	IsFinalReversal bool
}

// ReversalResult is the computed reversal, with its own snapshot.
type ReversalResult struct {
	Status models.DeterminationStatus

	TotalReversedTax string

	Components []ReversalComponentResult

	// Snapshot is the refund's own frozen record, referencing the original by the logical identity
	// the caller supplied. Tax needs no foreign key into the caller's schema (BR-TAX-ESS-055).
	Snapshot Snapshot
}

// ReversalComponentResult is what to reverse for one component.
type ReversalComponentResult struct {
	OriginalComponentReference string

	// ReversalTaxAmount is negative, so a caller summing charges and refunds gets the net.
	ReversalTaxAmount string

	RemainingTaxAmount string
}
