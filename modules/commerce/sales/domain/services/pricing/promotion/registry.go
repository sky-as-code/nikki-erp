package promotion

import (
	"github.com/shopspring/decimal"
)

// EvalContext is the read-only state an evaluator sees. Remaining holds each purchase line's
// untaxed value still open to discount at this point of the sequence (BR-R03); on the qualifying
// pass it equals the pre-promotion base.
type EvalContext struct {
	Input     *Input
	Remaining map[string]decimal.Decimal
}

// Benefit is what one reward yields over the current remaining base. A NotApplicableReason means the
// reward cannot apply at all (free_shipping without a shipping line); a zero amount with no gift is
// a zero benefit and is dropped by Apply (BR-R15).
type Benefit struct {
	RewardType            string
	DiscountUntaxedAmount decimal.Decimal
	Allocations           []Allocation
	Gift                  *GiftIntent
	NotApplicableReason   string
}

// ProgramEvaluator is the extension point per program_type: a new type is one more Register call,
// never an edit to Apply. Both methods must be pure and must not mutate ctx.
type ProgramEvaluator interface {
	// Qualify runs the apply-phase rules once on the pre-promotion base (BR-R15).
	Qualify(ctx EvalContext, program Program) Qualification

	// Benefit computes the chosen reward over ctx.Remaining.
	Benefit(ctx EvalContext, program Program, reward Reward, qualification Qualification) Benefit
}

// IssueEvaluator decides, on a source order, whether a program issues a next-order voucher.
type IssueEvaluator interface {
	QualifyIssue(ctx EvalContext, program Program) IssueDecision
}

// Registry maps program_type to its evaluators.
type Registry struct {
	evaluators map[string]ProgramEvaluator
	issuers    map[string]IssueEvaluator
}

// NewRegistry returns an empty registry; DefaultRegistry is the one the order flow uses.
func NewRegistry() *Registry {
	return &Registry{
		evaluators: map[string]ProgramEvaluator{},
		issuers:    map[string]IssueEvaluator{},
	}
}

// DefaultRegistry holds every built-in type. automatic, discount_code and coupon share one discount
// evaluator; the two voucher types sit behind the voucher gate.
func DefaultRegistry() *Registry {
	registry := NewRegistry()
	discount := discountEvaluator{}
	registry.Register(ProgramAutomatic, discount)
	registry.Register(ProgramDiscountCode, voucherGate{inner: discount})
	registry.Register(ProgramCoupon, voucherGate{inner: discount})
	registry.Register(ProgramBuyXGetY, buyXGetYEvaluator{})
	registry.Register(ProgramNextOrderCoupon, voucherGate{inner: discount})
	registry.RegisterIssuer(ProgramNextOrderCoupon, nextOrderIssuer{})
	return registry
}

func (this *Registry) Register(programType string, evaluator ProgramEvaluator) {
	this.evaluators[programType] = evaluator
}

func (this *Registry) RegisterIssuer(programType string, evaluator IssueEvaluator) {
	this.issuers[programType] = evaluator
}

func (this *Registry) Evaluator(programType string) (ProgramEvaluator, bool) {
	evaluator, found := this.evaluators[programType]
	return evaluator, found
}

func (this *Registry) Issuer(programType string) (IssueEvaluator, bool) {
	issuer, found := this.issuers[programType]
	return issuer, found
}

// voucherGate refuses a program that no attached voucher unlocks. A valid code never bypasses the
// base conditions (BR-C01), so the inner evaluator still qualifies normally.
type voucherGate struct {
	inner ProgramEvaluator
}

func (this voucherGate) Qualify(ctx EvalContext, program Program) Qualification {
	if program.VoucherId == "" {
		return Qualification{ReasonCode: ReasonVoucherRequired}
	}
	if program.VoucherSourceOrderId != "" && program.VoucherSourceOrderId == ctx.Input.OrderId {
		return Qualification{ReasonCode: ReasonVoucherSourceOrder}
	}
	return this.inner.Qualify(ctx, program)
}

func (this voucherGate) Benefit(
	ctx EvalContext, program Program, reward Reward, qualification Qualification,
) Benefit {
	return this.inner.Benefit(ctx, program, reward, qualification)
}
