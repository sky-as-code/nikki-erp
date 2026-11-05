package promotion

import (
	"github.com/shopspring/decimal"
)

// discountTarget is a line receiving a discount: basis is the untaxed value open to this reward,
// quantity the units that basis covers.
type discountTarget struct {
	line     Line
	basis    decimal.Decimal
	quantity decimal.Decimal
}

// discountTargets resolves BR-R02: the receiving scope is independent from the qualifying lines.
// Lines outside the program's pricelists never receive a discount either (D2).
func discountTargets(ctx EvalContext, program Program, reward Reward) []discountTarget {
	selector := reward.Selector
	if reward.DiscountScope == ScopeOrder {
		selector = Selector{}
	}
	targets := []discountTarget{}
	for _, line := range purchaseLines(ctx.Input) {
		remaining := ctx.Remaining[line.Key]
		if !remaining.IsPositive() || !pricelistAllows(program, line) || !selector.Matches(line) {
			continue
		}
		targets = append(targets, discountTarget{line: line, basis: remaining, quantity: line.Quantity})
	}
	if reward.DiscountScope == ScopeCheapestProduct {
		return cheapestUnit(targets)
	}
	return targets
}

// cheapestUnit keeps one unit of the cheapest line (BR-R05). The choice is made on the unit price
// after manual discounts, before any promotion, ties by line order then key. A line holding less
// than one unit offers only what it has.
func cheapestUnit(targets []discountTarget) []discountTarget {
	if len(targets) == 0 {
		return targets
	}
	best := 0
	for i := 1; i < len(targets); i++ {
		if cheaper(targets[i].line, targets[best].line) {
			best = i
		}
	}
	chosen := targets[best]
	units := decimal.Min(chosen.line.Quantity, decimal.NewFromInt(1))
	chosen.basis = chosen.basis.Mul(units).Div(chosen.line.Quantity)
	chosen.quantity = units
	return []discountTarget{chosen}
}

func cheaper(candidate Line, incumbent Line) bool {
	candidatePrice := candidate.UntaxedAmount.Div(candidate.Quantity)
	incumbentPrice := incumbent.UntaxedAmount.Div(incumbent.Quantity)
	if comparison := candidatePrice.Cmp(incumbentPrice); comparison != 0 {
		return comparison < 0
	}
	return lineLess(candidate, incumbent)
}
