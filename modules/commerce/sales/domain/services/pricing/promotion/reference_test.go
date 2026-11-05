package promotion

import (
	"math/rand"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func determinismInput() Input {
	lines := []Line{
		withCategories(buyLine("a", 1, "3", "10000"), "cat-1"),
		withCategories(buyLine("b", 2, "1", "33333"), "cat-1"),
		buyLine("c", 3, "7", "1429"),
		giftLine("g", 4, "gift", "rw", "1"),
	}
	cheapest := automatic("cheap", 5, anyRule(), percentReward("rw", ScopeCheapestProduct, "50", "0"))
	ten := voucherProgram("ten", 10, percentReward("rw", ScopeOrder, "10", "0"), "v-1")
	fifteen := automatic("fifteen", 20, anyRule(), percentReward("rw", ScopeOrder, "15", "0"))
	gift := automatic("gift", 30, anyRule(), giftReward("rw", "var-g", "1"))
	input := orderOf(lines, gift, fifteen, ten, cheapest)
	input.Selections = []Selection{{PromotionId: "gift", RewardId: "rw"}}
	return input
}

func referenceOf(result Result) string {
	return CalculationReference(ReferenceInput{
		OrderId: "order-1", StageVersion: 1, CurrencyScale: 0, Programs: result.Programs, GrandTotal: dec("90000"),
	})
}

// Identical input in any order yields the identical result and reference (D4, BR-R06 tie-breaks).
func TestApplyIsDeterministic(t *testing.T) {
	baseline := Apply(determinismInput())
	reference := referenceOf(baseline)
	random := rand.New(rand.NewSource(42))
	for round := 0; round < 50; round++ {
		input := determinismInput()
		random.Shuffle(len(input.Lines), func(i, j int) { input.Lines[i], input.Lines[j] = input.Lines[j], input.Lines[i] })
		random.Shuffle(len(input.Programs), func(i, j int) {
			input.Programs[i], input.Programs[j] = input.Programs[j], input.Programs[i]
		})
		result := Apply(input)
		require.Equal(t, baseline, result)
		require.Equal(t, reference, referenceOf(result))
	}
}

func TestReferenceChangesWithWhatTheCustomerGets(t *testing.T) {
	baseline := Apply(determinismInput())
	reference := referenceOf(baseline)
	require.Len(t, reference, 64)

	bumped := determinismInput()
	bumped.Programs[1].ConfigVersion = 2
	require.NotEqual(t, reference, referenceOf(Apply(bumped)), "config_version of an applied program")

	moreGoods := determinismInput()
	moreGoods.Lines[2].Quantity = dec("8")
	moreGoods.Lines[2].UntaxedAmount = dec("11432")
	require.NotEqual(t, reference, referenceOf(Apply(moreGoods)), "amounts")

	otherTotal := CalculationReference(ReferenceInput{
		OrderId: "order-1", StageVersion: 1, Programs: baseline.Programs, GrandTotal: dec("90001"),
	})
	require.NotEqual(t, reference, otherTotal, "grand total")

	otherStage := CalculationReference(ReferenceInput{
		OrderId: "order-1", StageVersion: 2, Programs: baseline.Programs, GrandTotal: dec("90000"),
	})
	require.NotEqual(t, reference, otherStage, "stage version")
}

// A candidate that applies nothing must not force a requote; equal decimals hash equal.
func TestReferenceIgnoresNonAppliedAndScaleNoise(t *testing.T) {
	baseline := Apply(determinismInput())
	reference := referenceOf(baseline)

	withLoser := determinismInput()
	withLoser.Programs = append(withLoser.Programs, automatic("never", 1, []Rule{applyRule("r", "999", "0")},
		fixedReward("rw", ScopeOrder, "1")))
	require.Equal(t, reference, referenceOf(Apply(withLoser)))

	rescaled := Apply(determinismInput())
	for i := range rescaled.Programs {
		if rescaled.Programs[i].Gift != nil {
			gift := *rescaled.Programs[i].Gift
			gift.Quantity = decimal.RequireFromString("1.000000")
			rescaled.Programs[i].Gift = &gift
		}
	}
	require.Equal(t, reference, referenceOf(rescaled))
}

// Invariants on the determinism basket: per line the promotion discount never exceeds the base and
// Σ allocations equals each program's discount.
func TestAllocationInvariants(t *testing.T) {
	result := Apply(determinismInput())
	for _, line := range result.Lines {
		require.False(t, line.RemainingUntaxedAmount.IsNegative(), line.Key)
		require.True(t, line.BaseUntaxedAmount.Sub(line.DiscountUntaxedAmount).Equal(line.RemainingUntaxedAmount))
	}
	for _, program := range result.Programs {
		sum := decimal.Zero
		for _, allocation := range program.Allocations {
			sum = sum.Add(allocation.DiscountUntaxedAmount)
			require.Equal(t, program.CalculationSequence, allocation.CalculationSequence)
		}
		require.True(t, sum.Equal(program.DiscountUntaxedAmount), program.PromotionId)
	}
	require.Equal(t, OutcomeNotApplicable, programResult(t, result, "ten").Outcome, "15% beats 10%")
	require.Equal(t, []string(nil), nilIfEmpty(result.RemovedGiftLineKeys))
}

func nilIfEmpty(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	return values
}
