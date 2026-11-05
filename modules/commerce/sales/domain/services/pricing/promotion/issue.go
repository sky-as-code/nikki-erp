package promotion

import (
	"time"

	"go.bryk.io/pkg/errors"
)

// IssueDecision says whether a confirmed source order earns a next-order voucher (BR-V09).
type IssueDecision struct {
	PromotionId   string
	ConfigVersion int32
	Qualified     bool
	ReasonCode    string
	Qualification Qualification
}

// nextOrderIssuer runs the issue-phase rules on the source order's pre-promotion base. It never
// grants the program's reward on the source order; that is the apply phase on a later order.
type nextOrderIssuer struct{}

func (this nextOrderIssuer) QualifyIssue(ctx EvalContext, program Program) IssueDecision {
	decision := IssueDecision{PromotionId: program.Id, ConfigVersion: program.ConfigVersion}
	if program.CurrencyId != "" && program.CurrencyId != ctx.Input.CurrencyId {
		decision.ReasonCode = ReasonCurrencyMismatch
		return decision
	}
	decision.Qualification = qualifyFirst(ctx.Input, program, PhaseIssue)
	if !decision.Qualification.Qualified {
		decision.ReasonCode = decision.Qualification.ReasonCode
		return decision
	}
	// Checked after the rules, so the reason is shown only to an order that would have earned one.
	if ctx.Input.CustomerId == "" {
		decision.ReasonCode = ReasonCustomerRequired
		return decision
	}
	decision.Qualified = true
	return decision
}

// EvaluateIssues runs the default registry. See Registry.EvaluateIssues.
func EvaluateIssues(input Input) []IssueDecision {
	return DefaultRegistry().EvaluateIssues(input)
}

// EvaluateIssues returns one decision per program whose type has an issuer, in (sequence, id)
// order. Vouchers are irrelevant here: issuance does not need one and consumes no usage (BR-V11).
func (this *Registry) EvaluateIssues(input Input) []IssueDecision {
	ctx := EvalContext{Input: &input, Remaining: baseRemaining(&input)}
	decisions := []IssueDecision{}
	for _, program := range sortedPrograms(input.Programs) {
		issuer, found := this.Issuer(program.ProgramType)
		if !found {
			continue
		}
		decisions = append(decisions, issuer.QualifyIssue(ctx, program))
	}
	return decisions
}

// IssueValidity is BR-V10: valid_from = confirmed_at, expires_at = min(confirmed_at + days,
// ends_at) with a null end ignored. ok is false when no positive window is left, and then nothing is
// issued. A nil expiresAt means the voucher never expires on its own.
func IssueValidity(
	confirmedAt time.Time, validityDays *int32, promotionEndsAt *time.Time,
) (validFrom time.Time, expiresAt *time.Time, ok bool, err error) {
	if validityDays != nil && *validityDays < 1 {
		return confirmedAt, nil, false, errors.Errorf(
			"next voucher validity days must be at least 1, got %d", *validityDays)
	}
	var end *time.Time
	if validityDays != nil {
		candidate := confirmedAt.AddDate(0, 0, int(*validityDays))
		end = &candidate
	}
	if promotionEndsAt != nil && (end == nil || promotionEndsAt.Before(*end)) {
		candidate := *promotionEndsAt
		end = &candidate
	}
	if end != nil && !end.After(confirmedAt) {
		return confirmedAt, nil, false, nil
	}
	return confirmedAt, end, true, nil
}
