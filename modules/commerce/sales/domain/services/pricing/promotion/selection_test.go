package promotion

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// AC-17: two rewards are alternatives; nothing applies until one is chosen, then only that one.
func TestRewardChoice(t *testing.T) {
	program := automatic("p1", 10, anyRule(),
		percentReward("rw-b", ScopeOrder, "10", "0"), fixedReward("rw-a", ScopeOrder, "5000"))
	lines := []Line{buyLine("l1", 1, "1", "100000")}

	unchosen := programResult(t, Apply(orderOf(lines, program)), "p1")
	require.Equal(t, OutcomeSelectionRequired, unchosen.Outcome)
	require.Equal(t, []string{"rw-a", "rw-b"}, []string{unchosen.Rewards[0].RewardId, unchosen.Rewards[1].RewardId})
	require.Empty(t, unchosen.Allocations)

	input := orderOf(lines, program)
	input.Selections = []Selection{{PromotionId: "p1", RewardId: "rw-b"}}
	chosen := programResult(t, Apply(input), "p1")
	require.Equal(t, OutcomeApplied, chosen.Outcome)
	require.Equal(t, "rw-b", chosen.RewardId)
	requireDec(t, "10000", chosen.DiscountUntaxedAmount, "only the chosen reward")
	require.True(t, chosen.Rewards[1].Selected)
	require.False(t, chosen.Rewards[0].Selected)

	input.Selections = []Selection{{PromotionId: "p1", Declined: true}}
	require.Equal(t, OutcomeDeclined, programResult(t, Apply(input), "p1").Outcome)

	input.Selections = []Selection{{PromotionId: "p1", RewardId: "rw-gone"}}
	require.Equal(t, OutcomeSelectionRequired, programResult(t, Apply(input), "p1").Outcome)
}

// AC-18: a second entry of the same promotion (a second voucher) never doubles the reward.
func TestSamePromotionTwiceAppliesOnce(t *testing.T) {
	first := automatic("p1", 10, anyRule(), fixedReward("rw", ScopeOrder, "1000"))
	first.ProgramType, first.VoucherId = ProgramDiscountCode, "v1"
	second := first
	second.VoucherId = "v2"
	result := Apply(orderOf([]Line{buyLine("l1", 1, "1", "100000")}, first, second))
	require.Len(t, result.Programs, 1)
	require.Equal(t, "v1", result.Programs[0].VoucherId)
	requireDec(t, "1000", result.Lines[0].DiscountUntaxedAmount, "single discount")
}

func TestOrderScopeWinner(t *testing.T) {
	lines := []Line{buyLine("l1", 1, "1", "100000")}
	cases := []struct {
		name     string
		programs []Program
		winner   string
	}{
		{
			name: "AC-19 the 20% wins over 10% whatever the sequence",
			programs: []Program{
				voucherProgram("ten", 1, percentReward("rw", ScopeOrder, "10", "0"), "v-ten"),
				voucherProgram("twenty", 50, percentReward("rw", ScopeOrder, "20", "0"), "v-twenty"),
			},
			winner: "twenty",
		},
		{
			name: "equal D: the smaller sequence wins",
			programs: []Program{
				automatic("late", 20, anyRule(), fixedReward("rw", ScopeOrder, "5000")),
				automatic("early", 10, anyRule(), percentReward("rw", ScopeOrder, "5", "0")),
			},
			winner: "early",
		},
		{
			name: "equal D and sequence: the smaller id wins",
			programs: []Program{
				automatic("p-b", 10, anyRule(), fixedReward("rw", ScopeOrder, "5000")),
				automatic("p-a", 10, anyRule(), fixedReward("rw", ScopeOrder, "5000")),
			},
			winner: "p-a",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			result := Apply(orderOf(lines, testCase.programs...))
			for _, program := range result.Programs {
				if program.PromotionId == testCase.winner {
					require.Equal(t, OutcomeApplied, program.Outcome)
					continue
				}
				require.Equal(t, OutcomeNotApplicable, program.Outcome, program.PromotionId)
				require.Equal(t, ReasonOrderDiscountNotBest, program.ReasonCode)
				require.Empty(t, program.Allocations)
			}
		})
	}
}

func voucherProgram(id string, sequence int32, reward Reward, voucherId string) Program {
	program := automatic(id, sequence, anyRule(), reward)
	program.ProgramType, program.VoucherId = ProgramDiscountCode, voucherId
	return program
}

// Non-order scopes and gifts stack with the order-scope winner (BR-R14).
func TestOtherScopesStackWithWinner(t *testing.T) {
	lines := []Line{buyLine("a", 1, "2", "10000"), buyLine("b", 2, "1", "50000")}
	cheapest := automatic("cheap", 5, anyRule(), percentReward("rw", ScopeCheapestProduct, "100", "0"))
	order := automatic("order", 10, anyRule(), percentReward("rw", ScopeOrder, "10", "0"))
	gift := automatic("gift", 15, anyRule(), giftReward("rw", "var-g", "1"))
	input := orderOf(lines, gift, order, cheapest)
	input.Selections = []Selection{{PromotionId: "gift", RewardId: "rw"}}
	result := Apply(input)

	for _, id := range []string{"cheap", "order", "gift"} {
		require.Equal(t, OutcomeApplied, programResult(t, result, id).Outcome, id)
	}
	requireDec(t, "10000", programResult(t, result, "cheap").DiscountUntaxedAmount, "one unit of a")
	requireDec(t, "6000", programResult(t, result, "order").DiscountUntaxedAmount, "10% of 60,000 left")
	require.Len(t, result.Gifts, 1)
}

// AC-39: free shipping never applies and is never turned into a merchandise discount.
func TestFreeShippingNotApplicable(t *testing.T) {
	shipping := Reward{Id: "rw", RewardType: RewardFreeShipping, MaxDiscountAmount: dec("0")}
	result := Apply(orderOf([]Line{buyLine("l1", 1, "1", "100000")}, automatic("p1", 10, anyRule(), shipping)))
	program := programResult(t, result, "p1")
	require.Equal(t, OutcomeNotApplicable, program.Outcome)
	require.Equal(t, ReasonNoShippingLine, program.ReasonCode)
	require.Empty(t, program.Allocations)
	requireDec(t, "0", result.Lines[0].DiscountUntaxedAmount, "no merchandise discount")
}

// AC-40: qualifying with a zero benefit creates no application.
func TestZeroBenefitDropped(t *testing.T) {
	lines := []Line{buyLine("a", 1, "1", "50000")}
	wipeOut := automatic("p1", 10, anyRule(), fixedReward("rw", ScopeOrder, "50000"))
	onA := automatic("p2", 20, anyRule(), withRewardSelector(fixedReward("rw", ScopeSpecificProducts, "1000"), "tpl-a"))
	result := Apply(orderOf(lines, wipeOut, onA))
	program := programResult(t, result, "p2")
	require.True(t, program.Qualification.Qualified)
	require.Equal(t, OutcomeNotApplicable, program.Outcome)
	require.Equal(t, ReasonZeroBenefit, program.ReasonCode)
	require.Zero(t, program.CalculationSequence)
}
