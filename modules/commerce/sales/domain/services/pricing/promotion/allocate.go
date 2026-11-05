package promotion

import (
	"sort"

	"github.com/shopspring/decimal"
)

// shareTarget is one recipient of a proportional split. capacity, when positive, is the most it may
// receive: a line is never discounted beyond what remains on it (BR-R06).
type shareTarget struct {
	key        string
	lineNumber int32
	weight     decimal.Decimal
	capacity   decimal.Decimal
}

// allocateLargestRemainder splits a non-negative total, already on the currency grain, in proportion
// to the weights. Each share is first floored to the grain; the leftover increments then go one at a
// time to the largest fractional remainder, ties by line number then key — the same semantics as
// accounting's AllocateDocumentRounding — so the shares sum to the total exactly and identical input
// always yields the identical split.
func allocateLargestRemainder(total decimal.Decimal, targets []shareTarget, scale int32) []decimal.Decimal {
	shares := make([]decimal.Decimal, len(targets))
	weightTotal := decimal.Zero
	for _, target := range targets {
		weightTotal = weightTotal.Add(target.weight)
	}
	if len(targets) == 0 || !weightTotal.IsPositive() || !total.IsPositive() {
		for i := range shares {
			shares[i] = decimal.Zero
		}
		return shares
	}

	remainders := make([]decimal.Decimal, len(targets))
	allocated := decimal.Zero
	for i, target := range targets {
		exact := total.Mul(target.weight).Div(weightTotal)
		shares[i] = exact.Truncate(scale)
		remainders[i] = exact.Sub(shares[i])
		allocated = allocated.Add(shares[i])
	}

	increment := decimal.New(1, -scale)
	steps := total.Sub(allocated).Div(increment).Round(0).IntPart()
	order := remainderOrder(targets, remainders)
	// Normally one pass suffices; a capped target can push an increment to a second pass.
	for steps > 0 {
		given := distributeIncrements(shares, targets, order, increment, steps)
		if given == 0 {
			break
		}
		steps -= given
	}
	return shares
}

func distributeIncrements(
	shares []decimal.Decimal, targets []shareTarget, order []int, increment decimal.Decimal, steps int64,
) int64 {
	given := int64(0)
	for _, index := range order {
		if given == steps {
			break
		}
		next := shares[index].Add(increment)
		if targets[index].capacity.IsPositive() && next.GreaterThan(targets[index].capacity) {
			continue
		}
		shares[index] = next
		given++
	}
	return given
}

func remainderOrder(targets []shareTarget, remainders []decimal.Decimal) []int {
	order := make([]int, len(targets))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(i, j int) bool {
		left, right := order[i], order[j]
		if comparison := remainders[left].Cmp(remainders[right]); comparison != 0 {
			return comparison > 0
		}
		if targets[left].lineNumber != targets[right].lineNumber {
			return targets[left].lineNumber < targets[right].lineNumber
		}
		return targets[left].key < targets[right].key
	})
	return order
}
