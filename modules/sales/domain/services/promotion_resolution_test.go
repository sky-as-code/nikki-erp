package services

import (
	"math/rand"
	"testing"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

func program(id string, priority int32, policy models.PromotionStackPolicy) CandidateProgram {
	return CandidateProgram{
		Id:          id,
		Priority:    priority,
		CreatedAt:   "2026-01-01T00:00:00Z",
		StackPolicy: string(policy),
	}
}

func grouped(id string, priority int32, policy models.PromotionStackPolicy,
	group string) CandidateProgram {
	candidate := program(id, priority, policy)
	candidate.ExclusiveGroup = group
	return candidate
}

func appliedIds(results []CandidateProgram) []string {
	ids := make([]string, len(results))
	for index, result := range results {
		ids[index] = result.Id
	}
	return ids
}

func assertIds(t *testing.T, got []CandidateProgram, want ...string) {
	t.Helper()
	ids := appliedIds(got)
	if len(ids) != len(want) {
		t.Fatalf("applied %v, want %v", ids, want)
	}
	for index := range want {
		if ids[index] != want[index] {
			t.Fatalf("applied %v, want %v", ids, want)
		}
	}
}

// D-10: lower priority applies first, and the ordering is total.
func TestResolveOrdersByPriorityThenCreatedAtThenId(t *testing.T) {
	candidates := []CandidateProgram{
		{Id: "z", Priority: 5, CreatedAt: "2026-01-01", StackPolicy: "stackable"},
		{Id: "a", Priority: 5, CreatedAt: "2026-01-01", StackPolicy: "stackable"},
		{Id: "m", Priority: 1, CreatedAt: "2026-03-01", StackPolicy: "stackable"},
		{Id: "b", Priority: 5, CreatedAt: "2025-01-01", StackPolicy: "stackable"},
	}

	// m first (lowest priority); then among the priority-5 programs, b has the earliest
	// created_at; a and z tie on both, so the id separates them.
	assertIds(t, ResolvePromotions(candidates, nil), "m", "b", "a", "z")
}

// BR §29 forbids any dependence on database record order. This is the test the requirement asks for
// by name: shuffle the input, assert the same output.
func TestResolveIsIndependentOfInputOrder(t *testing.T) {
	candidates := []CandidateProgram{
		grouped("seasonal", 10, models.PromotionStackExclusiveWithinGroup, "seasonal"),
		grouped("summer", 20, models.PromotionStackExclusiveWithinGroup, "seasonal"),
		program("loyalty", 5, models.PromotionStackStackable),
		program("clearance", 30, models.PromotionStackExclusive),
		program("staff", 15, models.PromotionStackStackable),
	}
	rules := []CompatibilityRule{
		{ProgramAId: "loyalty", ProgramBId: "staff", Allowed: false},
	}

	baseline := appliedIds(ResolvePromotions(candidates, rules))
	if len(baseline) == 0 {
		t.Fatal("the fixture must apply at least one program for this test to mean anything")
	}

	random := rand.New(rand.NewSource(7))
	for attempt := 0; attempt < 50; attempt++ {
		shuffled := make([]CandidateProgram, len(candidates))
		copy(shuffled, candidates)
		random.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})
		shuffledRules := make([]CompatibilityRule, len(rules))
		copy(shuffledRules, rules)

		got := appliedIds(ResolvePromotions(shuffled, shuffledRules))
		if len(got) != len(baseline) {
			t.Fatalf("shuffling changed the result: %v vs %v", got, baseline)
		}
		for index := range baseline {
			if got[index] != baseline[index] {
				t.Fatalf("shuffling changed the result: %v vs %v", got, baseline)
			}
		}
	}
}

// D-09: an explicit denied row wins over everything, including two stackable programs.
func TestExplicitDenyBeatsStackablePolicy(t *testing.T) {
	candidates := []CandidateProgram{
		program("a", 1, models.PromotionStackStackable),
		program("b", 2, models.PromotionStackStackable),
	}
	rules := []CompatibilityRule{{ProgramAId: "a", ProgramBId: "b", Allowed: false}}

	assertIds(t, ResolvePromotions(candidates, rules), "a")
}

// The direction a rule was stored in must not matter: the relation is symmetric in meaning.
func TestCompatibilityRuleIsSymmetric(t *testing.T) {
	candidates := []CandidateProgram{
		program("a", 1, models.PromotionStackStackable),
		program("b", 2, models.PromotionStackStackable),
	}

	forward := ResolvePromotions(candidates,
		[]CompatibilityRule{{ProgramAId: "a", ProgramBId: "b", Allowed: false}})
	reverse := ResolvePromotions(candidates,
		[]CompatibilityRule{{ProgramAId: "b", ProgramBId: "a", Allowed: false}})

	if len(forward) != len(reverse) || forward[0].Id != reverse[0].Id {
		t.Errorf("a rule stored b->a must behave like a->b: %v vs %v",
			appliedIds(forward), appliedIds(reverse))
	}
}

// BR §27's "combine only with these vouchers": exclusive plus explicit allowed rows. This is the
// combination D-09's ordering exists to make expressible.
func TestExclusivePlusAllowedRowsExpressesAWhitelist(t *testing.T) {
	candidates := []CandidateProgram{
		program("headline", 1, models.PromotionStackExclusive),
		program("friend", 2, models.PromotionStackStackable),
		program("stranger", 3, models.PromotionStackStackable),
	}
	rules := []CompatibilityRule{
		{ProgramAId: "headline", ProgramBId: "friend", Allowed: true},
	}

	// headline is exclusive, so stranger is refused; friend is explicitly allowed, so it survives.
	assertIds(t, ResolvePromotions(candidates, rules), "headline", "friend")
}

// An exclusive program refuses others whichever side of the comparison it lands on. Checking one
// direction only would let it be dragged in by a stackable program evaluated first.
func TestExclusiveRefusesInBothDirections(t *testing.T) {
	// The stackable one sorts first, so the exclusive one is the candidate being tested against an
	// already-accepted incumbent.
	candidates := []CandidateProgram{
		program("stackable_first", 1, models.PromotionStackStackable),
		program("exclusive_second", 2, models.PromotionStackExclusive),
	}
	assertIds(t, ResolvePromotions(candidates, nil), "stackable_first")

	// And the other way round: the exclusive one sorts first and refuses what follows.
	candidates = []CandidateProgram{
		program("exclusive_first", 1, models.PromotionStackExclusive),
		program("stackable_second", 2, models.PromotionStackStackable),
	}
	assertIds(t, ResolvePromotions(candidates, nil), "exclusive_first")
}

// exclusive_within_group excludes only programs sharing its group.
func TestExclusiveWithinGroup(t *testing.T) {
	candidates := []CandidateProgram{
		grouped("seasonal_a", 1, models.PromotionStackExclusiveWithinGroup, "seasonal"),
		grouped("seasonal_b", 2, models.PromotionStackExclusiveWithinGroup, "seasonal"),
		grouped("staff", 3, models.PromotionStackExclusiveWithinGroup, "staff"),
	}

	// seasonal_b is excluded by seasonal_a; staff is in a different group and survives.
	assertIds(t, ResolvePromotions(candidates, nil), "seasonal_a", "staff")
}

// An empty group is not a group. Two programs that both left it blank are not thereby in the same
// one — otherwise every unconfigured exclusive_within_group program would silently exclude every
// other, which is the opposite of what leaving a field blank should do.
func TestEmptyExclusiveGroupDoesNotExclude(t *testing.T) {
	candidates := []CandidateProgram{
		grouped("a", 1, models.PromotionStackExclusiveWithinGroup, ""),
		grouped("b", 2, models.PromotionStackExclusiveWithinGroup, ""),
	}
	assertIds(t, ResolvePromotions(candidates, nil), "a", "b")
}

// A policy this build does not recognise refuses to stack. The value came from a database row, and
// the safe reading of "I do not know how this combines" is that it does not.
func TestUnknownStackPolicyRefusesToCombine(t *testing.T) {
	candidates := []CandidateProgram{
		program("known", 1, models.PromotionStackStackable),
		{Id: "mystery", Priority: 2, CreatedAt: "2026-01-01", StackPolicy: "from_the_future"},
	}
	assertIds(t, ResolvePromotions(candidates, nil), "known")
}

// A contradictory pair of rows resolves to denied, whichever order they were read in. Otherwise the
// answer would depend on database record order, which BR §29 forbids.
func TestContradictoryRulesResolveToDenied(t *testing.T) {
	candidates := []CandidateProgram{
		program("a", 1, models.PromotionStackStackable),
		program("b", 2, models.PromotionStackStackable),
	}

	denyFirst := ResolvePromotions(candidates, []CompatibilityRule{
		{ProgramAId: "a", ProgramBId: "b", Allowed: false},
		{ProgramAId: "a", ProgramBId: "b", Allowed: true},
	})
	allowFirst := ResolvePromotions(candidates, []CompatibilityRule{
		{ProgramAId: "a", ProgramBId: "b", Allowed: true},
		{ProgramAId: "a", ProgramBId: "b", Allowed: false},
	})

	if len(denyFirst) != 1 || len(allowFirst) != 1 {
		t.Errorf("a contradictory pair must resolve to denied whichever came first: %v vs %v",
			appliedIds(denyFirst), appliedIds(allowFirst))
	}
}

// The greedy consequence, asserted so it is a decision rather than a surprise: a low-priority
// program can exclude a later one that would have been worth more.
func TestGreedyInPriorityOrderCanExcludeABetterOffer(t *testing.T) {
	candidates := []CandidateProgram{
		program("small_but_first", 1, models.PromotionStackExclusive),
		program("large_but_later", 2, models.PromotionStackStackable),
	}

	assertIds(t, ResolvePromotions(candidates, nil), "small_but_first")
}

func TestResolveEmptyInput(t *testing.T) {
	if got := ResolvePromotions(nil, nil); got != nil {
		t.Errorf("resolving no candidates = %v, want nil", got)
	}
}

// All stackable and no rules: everything applies, in priority order.
func TestAllStackableApplyTogether(t *testing.T) {
	candidates := []CandidateProgram{
		program("c", 3, models.PromotionStackStackable),
		program("a", 1, models.PromotionStackStackable),
		program("b", 2, models.PromotionStackStackable),
	}
	assertIds(t, ResolvePromotions(candidates, nil), "a", "b", "c")
}

// Resolution must not mutate the caller's slice. A caller that passed its own candidate list and
// found it reordered afterwards would have a bug nothing here reported.
func TestResolveDoesNotMutateInput(t *testing.T) {
	candidates := []CandidateProgram{
		program("z", 9, models.PromotionStackStackable),
		program("a", 1, models.PromotionStackStackable),
	}
	ResolvePromotions(candidates, nil)

	if candidates[0].Id != "z" || candidates[1].Id != "a" {
		t.Errorf("the input slice was reordered: %v", appliedIds(candidates))
	}
}
