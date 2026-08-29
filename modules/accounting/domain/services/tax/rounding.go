// Package tax implements determination, calculation, rounding and reversal, applied in that order.
//
// Everything here is pure arithmetic over caller-supplied values: no database, no Sales dependency,
// so a calculation is repeatable and testable standalone.
package tax

import (
	"github.com/shopspring/decimal"

	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
)

// RoundingPolicy is the subset of a stored policy the arithmetic needs. A policy is resolved for
// one currency; never apply it to another.
type RoundingPolicy struct {
	Scope     models.RoundingScope
	Method    models.RoundingMethod
	Increment decimal.Decimal
}

// Round applies the policy's method and increment to an amount.
//
// The scale is an increment (a quantum), not a decimal-place count, because some currencies round
// to a non-power-of-ten: 0.05 for cash settlement, 1 for VND. Divide by the increment, round to a
// whole number, multiply back.
//
// A non-positive increment would divide by zero or invert the sign, so it means "do not round".
func (this RoundingPolicy) Round(amount decimal.Decimal) decimal.Decimal {
	if !this.Increment.IsPositive() {
		return amount
	}

	quotient := amount.Div(this.Increment)

	var rounded decimal.Decimal
	switch this.Method {
	case models.RoundingHalfUp:
		// Half-up here is half away from zero: -0.5 goes to -1 as 0.5 goes to 1. Refunds depend
		// on that symmetry, or a reversal leaves a residual it can never clear.
		rounded = quotient.Round(0)
	case models.RoundingHalfEven:
		rounded = quotient.RoundBank(0)
	case models.RoundingUp:
		rounded = quotient.Ceil()
	case models.RoundingDown:
		rounded = quotient.Floor()
	default:
		return amount
	}

	return rounded.Mul(this.Increment)
}

// AllocationInput is one line-component amount entering document-scope rounding.
type AllocationInput struct {
	// LineReference and ComponentSequence are opaque identifiers, echoed back and never resolved
	// against Sales.
	LineReference     string
	ComponentSequence int32

	// GroupKey is the summing bucket: tax, rate version, treatment and jurisdiction together.
	// Different taxes must never be pooled — the per-tax document total is what a VAT return
	// reports.
	GroupKey string

	// Unrounded is the exact computed tax for this component, before any rounding.
	Unrounded decimal.Decimal
}

// AllocationResult is one component's share after document rounding.
type AllocationResult struct {
	LineReference     string
	ComponentSequence int32
	GroupKey          string

	// Rounded is the amount to store for this component.
	Rounded decimal.Decimal

	// Adjustment is this component's share of the group's rounding delta. It is snapshotted so a
	// later refund reverses the exact component that carried it instead of recomputing.
	Adjustment decimal.Decimal
}

// AllocateDocumentRounding does round-per-document, not round-per-line: each group's unrounded
// total is rounded once and is authoritative, then components are adjusted to sum to exactly that.
// Rounding components independently and summing can miss the document total by a few increments,
// and the document total is what the invoice and VAT return show.
//
// The delta goes out one increment at a time, largest fractional remainder first, ties broken by
// line reference then component sequence, so identical input always yields the same allocation and
// a later refund can reproduce which component carried the adjustment.
func AllocateDocumentRounding(
	inputs []AllocationInput, policy RoundingPolicy,
) []AllocationResult {
	results := make([]AllocationResult, len(inputs))
	for index, input := range inputs {
		results[index] = AllocationResult{
			LineReference:     input.LineReference,
			ComponentSequence: input.ComponentSequence,
			GroupKey:          input.GroupKey,
			Rounded:           policy.Round(input.Unrounded),
			Adjustment:        decimal.Zero,
		}
	}
	if !policy.Increment.IsPositive() {
		return results
	}

	for _, group := range groupIndexes(inputs) {
		groupTotal := decimal.Zero
		provisionalTotal := decimal.Zero
		for _, index := range group {
			groupTotal = groupTotal.Add(inputs[index].Unrounded)
			provisionalTotal = provisionalTotal.Add(results[index].Rounded)
		}

		delta := policy.Round(groupTotal).Sub(provisionalTotal)
		if delta.IsZero() {
			continue
		}

		// The delta is always a whole number of increments: both sides are multiples of one.
		steps := int(delta.Div(policy.Increment).Round(0).IntPart())
		step := policy.Increment
		if steps < 0 {
			steps = -steps
			step = step.Neg()
		}

		for _, index := range allocationOrder(inputs, group, policy) {
			if steps == 0 {
				break
			}
			results[index].Rounded = results[index].Rounded.Add(step)
			results[index].Adjustment = results[index].Adjustment.Add(step)
			steps--
		}
	}
	return results
}

// groupIndexes buckets input positions by GroupKey in first-seen order. Map iteration order is
// randomised, and an allocation depending on it would differ between runs on identical input.
func groupIndexes(inputs []AllocationInput) [][]int {
	order := []string{}
	byKey := map[string][]int{}
	for index, input := range inputs {
		if _, seen := byKey[input.GroupKey]; !seen {
			order = append(order, input.GroupKey)
		}
		byKey[input.GroupKey] = append(byKey[input.GroupKey], index)
	}

	groups := make([][]int, 0, len(order))
	for _, key := range order {
		groups = append(groups, byKey[key])
	}
	return groups
}

// allocationOrder sorts a group's positions by largest absolute fractional remainder first, so the
// component that lost most to rounding is repaid first. Ties fall back to line reference then
// component sequence, keeping the order total and independent of caller input order.
func allocationOrder(
	inputs []AllocationInput, group []int, policy RoundingPolicy,
) []int {
	ordered := make([]int, len(group))
	copy(ordered, group)

	remainder := func(index int) decimal.Decimal {
		return inputs[index].Unrounded.Sub(policy.Round(inputs[index].Unrounded)).Abs()
	}

	// Insertion sort: a group is a handful of components.
	for outer := 1; outer < len(ordered); outer++ {
		current := ordered[outer]
		inner := outer - 1
		for inner >= 0 && lessForAllocation(inputs, remainder, current, ordered[inner]) {
			ordered[inner+1] = ordered[inner]
			inner--
		}
		ordered[inner+1] = current
	}
	return ordered
}

func lessForAllocation(
	inputs []AllocationInput, remainder func(int) decimal.Decimal, left int, right int,
) bool {
	leftRemainder, rightRemainder := remainder(left), remainder(right)
	if !leftRemainder.Equal(rightRemainder) {
		return leftRemainder.GreaterThan(rightRemainder)
	}
	if inputs[left].LineReference != inputs[right].LineReference {
		return inputs[left].LineReference < inputs[right].LineReference
	}
	return inputs[left].ComponentSequence < inputs[right].ComponentSequence
}
