package services

import (
	"sort"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// Promotion conflict resolution (BR §29, with D-09 and D-10).
//
// Pure, like the state machines and the allocator: it takes the candidate programs and the
// compatibility rows and returns the ordered application list. No repository, no context.
//
// BR §29 forbids any dependence on database record order, and that is the property worth testing
// hardest: the same cart evaluated twice must produce the same discounts, whatever order the rows
// came back in. The sort key is total — priority, then created_at, then id — so there is never a
// pair the comparison leaves unordered (D-10).

// CandidateProgram is one program that has already passed its own eligibility conditions.
//
// Eligibility and conflict resolution are separate steps on purpose: whether a program's conditions
// hold is about the cart, while whether two programs may apply together is about the programs. Mixing
// them would make a compatibility rule depend on basket contents, which is not what BR §27 means.
type CandidateProgram struct {
	Id string

	// Priority orders application, LOWER FIRST (D-10).
	Priority int32

	// CreatedAt breaks a priority tie. A string rather than a time, because it is only ever
	// compared and the caller already has it in whatever form the repository returned.
	CreatedAt string

	StackPolicy    string
	ExclusiveGroup string
}

// CompatibilityRule is one explicit pairwise directive. Direction does not matter: the relation is
// symmetric in meaning, so the resolver looks a pair up both ways round.
type CompatibilityRule struct {
	ProgramAId string
	ProgramBId string
	Allowed    bool
}

// ResolvePromotions picks which of the candidates apply, in application order.
//
// The algorithm, per BR §29:
//  1. sort by priority, then created_at, then id — a total order, independent of input order;
//  2. walk the sorted list, accepting a program only if it is compatible with every program already
//     accepted.
//
// Greedy-in-priority-order rather than searching for the best combination. That is a real choice
// with a consequence worth stating: a low-priority program can exclude a later one that would have
// been worth more to the customer. It is what BR §29 describes, and the alternative — maximising
// total discount — would make the applied set depend on prices, so two carts differing only in
// quantity could get structurally different promotions with no explanation an operator could follow.
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

// sortCandidates imposes the D-10 total order: priority ascending, then created_at, then id.
//
// All three levels are needed. Priority alone leaves ties; created_at can collide when rows are
// written in one transaction; the id is a ULID, so it is the key that always separates.
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
// so a contradictory pair of rows resolves the safe way rather than by which was read first.
func indexRules(rules []CompatibilityRule) map[[2]string]bool {
	index := make(map[[2]string]bool, len(rules))
	for _, rule := range rules {
		key := pairKey(rule.ProgramAId, rule.ProgramBId)
		if existing, seen := index[key]; seen && !existing {
			// Already denied. A later allowed row does not override it.
			continue
		}
		index[key] = rule.Allowed
	}
	return index
}

// programsMayCombine applies the D-09 resolution order to one pair:
// explicit denied wins always, then explicit allowed permits, then stack_policy decides.
//
// That order is what makes BR §27's "combine only with these vouchers" expressible — as
// stack_policy exclusive plus explicit allowed rows — and it settles the ambiguity BR §28 leaves
// about a pair with no row at all.
func programsMayCombine(a, b CandidateProgram, index map[[2]string]bool) bool {
	if allowed, found := index[pairKey(a.Id, b.Id)]; found {
		return allowed
	}
	return stackPolicyPermits(a, b) && stackPolicyPermits(b, a)
}

// stackPolicyPermits asks whether `subject` tolerates `other` in the absence of an explicit rule.
//
// It is asked in BOTH directions by the caller, so one exclusive program is enough to refuse the
// pair. Checking only one side would let an exclusive program be dragged in by a stackable one that
// happened to be evaluated first.
func stackPolicyPermits(subject, other CandidateProgram) bool {
	policy := models.PromotionStackPolicy(subject.StackPolicy)
	switch policy {
	case models.PromotionStackExclusive:
		return false
	case models.PromotionStackExclusiveWithinGroup:
		// Excludes only programs sharing its group. An empty group is not a group: two programs
		// that both left it blank are not thereby in the same one, or every unconfigured
		// exclusive_within_group program would silently exclude every other.
		if subject.ExclusiveGroup == "" {
			return true
		}
		return subject.ExclusiveGroup != other.ExclusiveGroup
	case models.PromotionStackStackable:
		return true
	}
	// An unrecognised policy refuses to stack. The value came from a database row, and a policy
	// this build does not understand must not be assumed permissive: the safe reading of "I do not
	// know how this combines" is that it does not.
	return false
}
