package tax

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

type CalculateResult = dyn.OpResult[CalculationResult]
type SimulateResult = dyn.OpResult[SimulationResult]
type ReverseResult = dyn.OpResult[ReversalResult]

// TaxCalculationAppService is the capability other modules consume. A consuming module declares its
// own port in interfaces/external and binds this once in infra/external rather than importing it
// from a domain or application layer.
type TaxCalculationAppService interface {
	// Calculate determines and computes tax for a whole document. It is pure: no invoice, posting,
	// stock movement or change to tax master data, so a draft order may recalculate on every edit.
	Calculate(ctx corectx.Context, request CalculationRequest) (*CalculateResult, error)

	// Simulate runs the same pipeline and also returns the trace of how it got there. It is separate
	// from Calculate because assembling the explanation is too expensive for the hot path.
	Simulate(ctx corectx.Context, request CalculationRequest) (*SimulateResult, error)

	// ReverseFull reverses an entire original charge from its frozen snapshot.
	ReverseFull(ctx corectx.Context, request FullReversalRequest) (*ReverseResult, error)

	// ReversePartial reverses part of an original charge.
	ReversePartial(ctx corectx.Context, request PartialReversalRequest) (*ReverseResult, error)
}

// TaxCalculationDomainService is the same capability without the permission checks. Authorization
// happens in app/ and nowhere else, so callers must authorize before calling this.
type TaxCalculationDomainService interface {
	Calculate(ctx corectx.Context, request CalculationRequest) (*CalculateResult, error)
	Simulate(ctx corectx.Context, request CalculationRequest) (*SimulateResult, error)
	ReverseFull(ctx corectx.Context, request FullReversalRequest) (*ReverseResult, error)
	ReversePartial(ctx corectx.Context, request PartialReversalRequest) (*ReverseResult, error)
}

// SimulationResult is a calculation plus the trace of how it was reached: matched rules, the
// mapping, the applicable taxes, their rates and components, and the final figure.
type SimulationResult struct {
	Calculation CalculationResult
	Trace       []TraceStep
}

// TraceStep is one stage of the pipeline, as it happened.
type TraceStep struct {
	// Stage names the pipeline step: candidate_taxes, rule_evaluation, mapping, override,
	// version_resolution, calculation, rounding.
	Stage string

	// Detail is human-readable, already localized where it names configuration.
	Detail string

	// TaxIds is the tax set as it stood after this step, showing where a tax entered or left.
	TaxIds []string

	// RuleIds names the rules that fired, when the step is rule evaluation.
	RuleIds []string
}

// FullReversalRequest reverses an original charge in its entirety.
type FullReversalRequest struct {
	OrgId string

	// OriginalSnapshot is the frozen snapshot the caller stored at the time of sale; it is supplied
	// rather than looked up because Tax keeps no copy.
	OriginalSnapshot Snapshot

	// TaxDate is the refund's own date, recorded on the reversal snapshot. It does not re-resolve
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

// ReversalComponentRequest is the caller's reversal state for one original component. The
// already-reversed figures come from the caller because Tax stores no reversal state of its own.
type ReversalComponentRequest struct {
	OriginalComponentReference string
	OriginalReversibleBasis    string
	OriginalTaxAmount          string
	AlreadyReversedBasis       string
	AlreadyReversedTaxAmount   string
	RequestedReversalBasis     string

	// IsFinalReversal marks the last refund against this component; it absorbs the rounding residual
	// so the reversals sum to exactly the original charge.
	IsFinalReversal bool
}

// ReversalResult is the computed reversal, with its own snapshot.
type ReversalResult struct {
	Status models.DeterminationStatus

	TotalReversedTax string

	Components []ReversalComponentResult

	// Snapshot is the refund's own frozen record, referencing the original by the logical identity
	// the caller supplied rather than a foreign key into the caller's schema.
	Snapshot Snapshot
}

// ReversalComponentResult is what to reverse for one component.
type ReversalComponentResult struct {
	OriginalComponentReference string

	// ReversalTaxAmount is negative, so a caller summing charges and refunds gets the net.
	ReversalTaxAmount string

	RemainingTaxAmount string
}
