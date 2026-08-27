package services

import (
	"time"

	"github.com/shopspring/decimal"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
	taxsvc "github.com/sky-as-code/nikki-erp/modules/accounting/domain/services/tax"
	it "github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/tax"
)

// ReverseFull negates an entire original charge.
//
// The rate is deliberately not re-resolved. A full reversal undoes what was charged, and what was
// charged is in the snapshot the caller supplies — resolving today's rate would refund a different
// amount than was taken whenever a rate has changed since (BR-TAX-ESS-055).
func (this *TaxCalculationDomainServiceImpl) ReverseFull(
	ctx corectx.Context, request it.FullReversalRequest,
) (*it.ReverseResult, error) {
	inputs := toReversalInputs(request.Components)
	reversed := taxsvc.ReverseFull(inputs)

	return this.reversalResult(request.OriginalSnapshot, request.TaxDate, reversed), nil
}

// ReversePartial reverses part of an original charge.
//
// The rounding policy applies to the refund because a proportion of a rounded amount is rarely a
// round number itself. The final reversal absorbs whatever residual the proportions left, so a
// sequence of partial refunds sums to exactly the original charge rather than approximately
// (BR-TAX-ESS-033).
func (this *TaxCalculationDomainServiceImpl) ReversePartial(
	ctx corectx.Context, request it.PartialReversalRequest,
) (*it.ReverseResult, error) {
	policyRepo, err := RepoFor(models.TaxRoundingPolicySchemaName)
	if err != nil {
		return nil, err
	}

	policy, problem, err := this.resolveRoundingPolicy(ctx, policyRepo,
		it.CalculationRequest{
			RoundingPolicyCode: request.RoundingPolicyCode,
			TaxDate:            request.TaxDate,
		})
	if err != nil {
		return nil, err
	}
	if problem != "" {
		return &it.ReverseResult{
			HasData: true,
			Data: it.ReversalResult{
				Status:           models.DeterminationUnresolved,
				TotalReversedTax: decimal.Zero.String(),
				Snapshot:         this.reversalSnapshot(request.OriginalSnapshot, request.TaxDate, nil),
			},
		}, nil
	}

	inputs := toReversalInputs(request.Components)
	reversed := taxsvc.ReversePartial(inputs, policy)

	return this.reversalResult(request.OriginalSnapshot, request.TaxDate, reversed), nil
}

func (this *TaxCalculationDomainServiceImpl) reversalResult(
	original it.Snapshot, taxDate string, reversed []taxsvc.ReversalComponentResult,
) *it.ReverseResult {
	total := decimal.Zero
	components := make([]it.ReversalComponentResult, 0, len(reversed))
	for _, entry := range reversed {
		total = total.Add(entry.ReversalTaxAmount)
		components = append(components, it.ReversalComponentResult{
			OriginalComponentReference: entry.OriginalComponentReference,
			ReversalTaxAmount:          entry.ReversalTaxAmount.String(),
			RemainingTaxAmount:         entry.RemainingTaxAmount.String(),
		})
	}

	return &it.ReverseResult{
		HasData: true,
		Data: it.ReversalResult{
			Status:           models.DeterminationResolved,
			TotalReversedTax: total.String(),
			Components:       components,
			Snapshot:         this.reversalSnapshot(original, taxDate, components),
		},
	}
}

// reversalSnapshot is the refund's own frozen record.
//
// It carries the original's rounding configuration rather than resolving it again: a refund must
// round the way the sale did, or the reversals will not close to zero against it.
func (this *TaxCalculationDomainServiceImpl) reversalSnapshot(
	original it.Snapshot, taxDate string, components []it.ReversalComponentResult,
) it.Snapshot {
	snapshot := it.Snapshot{
		SchemaVersion:         it.SnapshotSchemaVersion,
		TaxDate:               taxDate,
		CalculatedAt:          this.stamp(),
		Status:                models.DeterminationResolved,
		CurrencyCode:          original.CurrencyCode,
		RoundingPolicyId:      original.RoundingPolicyId,
		RoundingPolicyVersion: original.RoundingPolicyVersion,
		RoundingScope:         original.RoundingScope,
		RoundingMethod:        original.RoundingMethod,
		RoundingIncrement:     original.RoundingIncrement,
	}
	if len(components) == 0 {
		snapshot.Status = models.DeterminationUnresolved
	}
	return snapshot
}

func toReversalInputs(requested []it.ReversalComponentRequest) []taxsvc.ReversalComponentInput {
	inputs := make([]taxsvc.ReversalComponentInput, 0, len(requested))
	for _, component := range requested {
		inputs = append(inputs, taxsvc.ReversalComponentInput{
			OriginalComponentReference: component.OriginalComponentReference,
			OriginalReversibleBasis:    parseAmount(component.OriginalReversibleBasis),
			OriginalTaxAmount:          parseAmount(component.OriginalTaxAmount),
			AlreadyReversedBasis:       parseAmount(component.AlreadyReversedBasis),
			AlreadyReversedTaxAmount:   parseAmount(component.AlreadyReversedTaxAmount),
			RequestedReversalBasis:     parseAmount(component.RequestedReversalBasis),
			IsFinalReversal:            component.IsFinalReversal,
		})
	}
	return inputs
}

// parseAmount reads a decimal that travelled as a string.
//
// Money crosses the wire as text precisely so no float ever touches it; an unparseable value is
// treated as zero rather than as an error, because the reversal arithmetic clamps to what was
// originally charged and a zero basis reverses nothing.
func parseAmount(value string) decimal.Decimal {
	if value == "" {
		return decimal.Zero
	}
	parsed, err := decimal.NewFromString(value)
	if err != nil {
		return decimal.Zero
	}
	return parsed
}

// buildSnapshot freezes how a calculation was reached.
//
// Self-contained by requirement (BR-TAX-ESS-SUP-032): a screen showing a three-year-old invoice
// reads this and never the current tax master, which is the entire mechanism by which a rate change
// cannot reinterpret a historical sale.
func (this *TaxCalculationDomainServiceImpl) buildSnapshot(
	request it.CalculationRequest, result it.CalculationResult, policy taxsvc.RoundingPolicy,
) it.Snapshot {
	return it.Snapshot{
		SchemaVersion:     it.SnapshotSchemaVersion,
		TaxDate:           request.TaxDate,
		CalculatedAt:      this.stamp(),
		Status:            result.Status,
		CurrencyCode:      request.CurrencyCode,
		RoundingScope:     policy.Scope,
		RoundingMethod:    policy.Method,
		RoundingIncrement: policy.Increment,
		AppliedRuleIds:    result.AppliedRuleIds,
		Lines:             result.Lines,
	}
}

// stamp is the calculated-at time, in RFC 3339.
//
// Distinct from the tax date in every way that matters: this records when the arithmetic ran, while
// the tax date decides which configuration governed it and is always the caller's.
func (this *TaxCalculationDomainServiceImpl) stamp() string {
	if this.now != nil {
		return this.now()
	}
	return time.Now().UTC().Format(time.RFC3339)
}
