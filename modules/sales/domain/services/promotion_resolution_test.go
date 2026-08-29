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

// Lower priority applies first, and the ordering is total.
func TestResolveOrdersByPriorityThenCreatedAtThenId(t *testing.T) {
	candidates := []CandidateProgram{
		{Id: "z", Priority: 5, CreatedAt: "2026-01-01", StackPolicy: "stackable"},
		{Id: "a", Priority: 5, CreatedAt: "2026-01-01", StackPolicy: "stackable"},
		{Id: "m", Priority: 1, CreatedAt: "2026-03-01", StackPolicy: "stackable"},
		{Id: "b", Priority: 5, CreatedAt: "2025-01-01", StackPolicy: "stackable"},
	}

	// m first (lowest priority); then b has the earliest created_at; a and z tie, so the id separates.
	assertIds(t, ResolvePromotions(candidates, nil), "m", "b", "a", "z")
}

// The result must not depend on record order: shuffle the input, assert the same output.
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

// An explicit denied row wins over everything, including two stackable programs.
func TestExplicitDenyBeatsStackablePolicy(t *testing.T) {
	candidates := []CandidateProgram{
		program("a", 1, models.PromotionStackStackable),
		program("b", 2, models.PromotionStackStackable),
	}
	rules := []CompatibilityRule{{ProgramAId: "a", ProgramBId: "b", Allowed: false}}

	assertIds(t, ResolvePromotions(candidates, rules), "a")
}

// The direction a rule was stored in must not matter: the relation is symmetric.
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

// "Combine only with these vouchers" is expressed as exclusive plus explicit allowed rows.
func TestExclusivePlusAllowedRowsExpressesAWhitelist(t *testing.T) {
	candidates := []CandidateProgram{
		program("headline", 1, models.PromotionStackExclusive),
		program("friend", 2, models.PromotionStackStackable),
		program("stranger", 3, models.PromotionStackStackable),
	}
	rules := []CompatibilityRule{
		{ProgramAId: "headline", ProgramBId: "friend", Allowed: true},
	}

	// headline is exclusive, so stranger is refused; friend is explicitly allowed and survives.
	assertIds(t, ResolvePromotions(candidates, rules), "headline", "friend")
}

// An exclusive program refuses others whichever side of the comparison it lands on, or a stackable
// program evaluated first could drag it in.
func TestExclusiveRefusesInBothDirections(t *testing.T) {
	// The stackable one sorts first, so the exclusive one is tested against an accepted incumbent.
	candidates := []CandidateProgram{
		program("stackable_first", 1, models.PromotionStackStackable),
		program("exclusive_second", 2, models.PromotionStackExclusive),
	}
	assertIds(t, ResolvePromotions(candidates, nil), "stackable_first")

	// And the other way round.
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

// An empty group is not a group, or every unconfigured exclusive_within_group program would silently
// exclude every other.
func TestEmptyExclusiveGroupDoesNotExclude(t *testing.T) {
	candidates := []CandidateProgram{
		grouped("a", 1, models.PromotionStackExclusiveWithinGroup, ""),
		grouped("b", 2, models.PromotionStackExclusiveWithinGroup, ""),
	}
	assertIds(t, ResolvePromotions(candidates, nil), "a", "b")
}

// A policy this build does not recognise refuses to stack rather than being assumed permissive.
func TestUnknownStackPolicyRefusesToCombine(t *testing.T) {
	candidates := []CandidateProgram{
		program("known", 1, models.PromotionStackStackable),
		{Id: "mystery", Priority: 2, CreatedAt: "2026-01-01", StackPolicy: "from_the_future"},
	}
	assertIds(t, ResolvePromotions(candidates, nil), "known")
}

// A contradictory pair of rows resolves to denied whichever order they were read in, or the answer
// would depend on record order.
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

// The greedy consequence: a low-priority program can exclude a later one worth more.
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

func TestAllStackableApplyTogether(t *testing.T) {
	candidates := []CandidateProgram{
		program("c", 3, models.PromotionStackStackable),
		program("a", 1, models.PromotionStackStackable),
		program("b", 2, models.PromotionStackStackable),
	}
	assertIds(t, ResolvePromotions(candidates, nil), "a", "b", "c")
}

// Resolution must not mutate the caller's slice.
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
