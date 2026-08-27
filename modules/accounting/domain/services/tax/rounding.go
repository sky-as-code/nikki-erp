// Package tax implements determination, calculation, rounding and reversal.
//
// Everything here is pure arithmetic over values the caller supplies. Nothing in this package
// reads a database, writes one, or knows that Sales exists — which is what makes a calculation
// safely repeatable (BR-TAX-ESS-046) and what lets the whole engine be tested without one.
package tax

import (
	"github.com/shopspring/decimal"

	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
)

// RoundingPolicy is the subset of a stored policy the arithmetic needs.
//
// It is a plain value rather than the stored model so that the calculator can be exercised without
// constructing a database row, and so that a caller holding a policy resolved for one currency
// cannot accidentally apply it to another by passing the whole record around.
type RoundingPolicy struct {
	Scope     models.RoundingScope
	Method    models.RoundingMethod
	Increment decimal.Decimal
}

// Round applies the policy's method and increment to an amount.
//
// Rounding is expressed as a quantum rather than a number of decimal places because currencies
// exist that round to something other than a power of ten — a 0.05 increment for cash settlement,
// or 1 for VND. Dividing by the increment, rounding to a whole number and multiplying back handles
// both without a special case (BR-TAX-ESS-SUP-017).
//
// A non-positive increment would divide by zero or invert the sign, so it is treated as "do not
// round" rather than allowed to produce nonsense. Configuration validation rejects such a policy at
// write time; this is the belt to that braces.
func (this RoundingPolicy) Round(amount decimal.Decimal) decimal.Decimal {
	if !this.Increment.IsPositive() {
		return amount
	}

	quotient := amount.Div(this.Increment)

	var rounded decimal.Decimal
	switch this.Method {
	case models.RoundingHalfUp:
		// decimal.Round rounds half away from zero, so -0.5 goes to -1 as 0.5 goes to 1. That
		// symmetry matters for refunds: rounding a reversal towards zero while its sale rounded
		// away would leave a residual the reversal could never clear.
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
	// LineReference and ComponentSequence identify the component the caller wants the result
	// attributed back to. They are opaque here: the engine never resolves a line reference against
	// Sales, it only echoes it (BR-TAX-ESS-025).
	LineReference     string
	ComponentSequence int32

	// GroupKey is what the amount is summed within — tax, rate version, treatment and jurisdiction
	// together. Two components of different taxes must never be pooled, because the document total
	// per tax is what a VAT return reports.
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

	// Adjustment is what this component absorbed of the group's rounding delta, and is recorded in
	// the snapshot so a later refund reverses the exact component that carried it rather than
	// recomputing the whole document (BR-TAX-ESS-SUP-019).
	Adjustment decimal.Decimal
}

// AllocateDocumentRounding rounds each group's total once and distributes the difference.
//
// The naive alternative — round every component and add them up — produces a document total that
// can differ from the correctly rounded one by a few units of the increment, and it is the document
// total that appears on the invoice and the VAT return. So the group total is rounded first and
// treated as authoritative; the components are then adjusted to sum to exactly that
// (BR-TAX-ESS-022, AC-TAX-SUP-15).
//
// The delta is handed out one increment at a time, to the components with the largest fractional
// remainder first. Ties break by line reference and then component sequence, so the same input
// always yields the same allocation — a refund computed later must be able to reproduce which
// component carried the adjustment (TAX-SUP-INV-12).
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

// groupIndexes buckets input positions by GroupKey, preserving first-seen order.
//
// Insertion order rather than map order, because iterating a Go map is deliberately randomised and
// an allocation that depended on it would differ between two runs on identical input.
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

// allocationOrder sorts a group's positions by who most deserves the next increment.
//
// Largest absolute fractional remainder first: the component that lost the most to rounding gets
// the adjustment back. Ties fall back to line reference and component sequence so the order is
// total, never dependent on the caller's input order alone.
func allocationOrder(
	inputs []AllocationInput, group []int, policy RoundingPolicy,
) []int {
	ordered := make([]int, len(group))
	copy(ordered, group)

	remainder := func(index int) decimal.Decimal {
		return inputs[index].Unrounded.Sub(policy.Round(inputs[index].Unrounded)).Abs()
	}

	// Insertion sort: a group is a handful of components, and this keeps the comparison rule in one
	// readable place without pulling in a comparator closure over three fields.
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
