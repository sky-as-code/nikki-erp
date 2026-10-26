package services

import (
	"sort"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// Promotion conflict resolution: pure, taking the candidate programs and the compatibility rows and
// returning the ordered application list.
//
// The result must not depend on database record order, so the sort key is total — priority, then
// created_at, then id — leaving no pair the comparison cannot order.

// CandidateProgram is one program that has already passed its own eligibility conditions. Eligibility
// and conflict resolution stay separate: mixing them would make a compatibility rule depend on basket
// contents.
type CandidateProgram struct {
	Id string

	// Priority orders application, lower first.
	Priority int32

	// CreatedAt breaks a priority tie. A string rather than a time, because it is only ever compared.
	CreatedAt string

	StackPolicy    string
	ExclusiveGroup string
}

// CompatibilityRule is one explicit pairwise directive. The relation is symmetric, so the resolver
// looks a pair up both ways round.
type CompatibilityRule struct {
	ProgramAId string
	ProgramBId string
	Allowed    bool
}

// ResolvePromotions picks which of the candidates apply, in application order: sort by priority, then
// created_at, then id, and accept a program only if it is compatible with every one already accepted.
//
// Greedy in priority order rather than searching for the best combination, so a low-priority program
// can exclude a later one worth more to the customer. Maximising total discount instead would make
// the applied set depend on prices, so two carts differing only in quantity could get structurally
// different promotions.
func ResolvePromotions(
	candidates []CandidateProgram, rules []CompatibilityRule,
) []CandidateProgram {
	if len(candidates) == 0 {
		return nil
	}

	ordered := make([]CandidateProgram, len(candidates))
	copy(ordered, candidates)
	sortCandidates(ordered)

	index := indexRules(rules)
	accepted := make([]CandidateProgram, 0, len(ordered))

	for _, candidate := range ordered {
		compatible := true
		for _, incumbent := range accepted {
			if !programsMayCombine(candidate, incumbent, index) {
				compatible = false
				break
			}
		}
		if compatible {
			accepted = append(accepted, candidate)
		}
	}
	return accepted
}

// sortCandidates imposes a total order: priority ascending, then created_at, then id. All three are
// needed — priority alone leaves ties, created_at collides when rows are written in one transaction,
// and the id is a ULID that always separates.
func sortCandidates(candidates []CandidateProgram) {
	sort.SliceStable(candidates, func(i, j int) bool {
		left, right := candidates[i], candidates[j]
		if left.Priority != right.Priority {
			return left.Priority < right.Priority
		}
		if left.CreatedAt != right.CreatedAt {
			return left.CreatedAt < right.CreatedAt
		}
		return left.Id < right.Id
	})
}

// pairKey normalises a pair so a rule is found whichever way round it was stored.
func pairKey(a, b string) [2]string {
	if a <= b {
		return [2]string{a, b}
	}
	return [2]string{b, a}
}

// indexRules turns the rule rows into a lookup. A denied row beats an allowed one for the same pair,
// so contradictory rows resolve the safe way rather than by which was read first.
func indexRules(rules []CompatibilityRule) map[[2]string]bool {
	index := make(map[[2]string]bool, len(rules))
	for _, rule := range rules {
		key := pairKey(rule.ProgramAId, rule.ProgramBId)
		if existing, seen := index[key]; seen && !existing {
			// Already denied; a later allowed row does not override it.
			continue
		}
		index[key] = rule.Allowed
	}
	return index
}

// programsMayCombine resolves one pair: explicit denied wins always, then explicit allowed permits,
// then stack_policy decides. That order makes "combine only with these vouchers" expressible as
// stack_policy exclusive plus explicit allowed rows.
func programsMayCombine(a, b CandidateProgram, index map[[2]string]bool) bool {
	if allowed, found := index[pairKey(a.Id, b.Id)]; found {
		return allowed
	}
	return stackPolicyPermits(a, b) && stackPolicyPermits(b, a)
}

// stackPolicyPermits asks whether subject tolerates other in the absence of an explicit rule. The
// caller asks in both directions, so one exclusive program refuses the pair; checking one side would
// let an exclusive program be dragged in by a stackable one evaluated first.
func stackPolicyPermits(subject, other CandidateProgram) bool {
	policy := models.PromotionStackPolicy(subject.StackPolicy)
	switch policy {
	case models.PromotionStackExclusive:
		return false
	case models.PromotionStackExclusiveWithinGroup:
		// Excludes only programs sharing its group. An empty group is not a group, or every
		// unconfigured exclusive_within_group program would silently exclude every other.
		if subject.ExclusiveGroup == "" {
			return true
		}
		return subject.ExclusiveGroup != other.ExclusiveGroup
	case models.PromotionStackStackable:
		return true
	}
	// An unrecognised policy refuses to stack: the value came from a database row, and a policy this
	// build does not understand must not be assumed permissive.
	return false
}
