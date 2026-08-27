package services

import (
	"testing"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The pure parts of applying a voucher: reading a program row into a candidate, and the conflict
// resolution that decides whether the voucher joins or is refused.
//
// The repository-touching paths (resolveCodeByString, loadVoucherProgram, loadConditionGroups) are
// not covered here. `engineFor` is a package var precisely so it could be stubbed, but nothing in
// this module stubs it yet and inventing a fake registry for one test would pin the fake rather than
// the behaviour. They are exercised live instead — see the SALES-023 note in 02-progress.md.

func TestCandidateFromReadsTheFieldsResolutionSortsOn(t *testing.T) {
	record := dmodel.DynamicFields{
		models.SalesPromotionProgramFieldId:             "PG1",
		models.SalesPromotionProgramFieldPriority:       int32(50),
		models.SalesPromotionProgramFieldStackPolicy:    string(models.PromotionStackExclusive),
		models.SalesPromotionProgramFieldExclusiveGroup: "seasonal",
		"created_at": "2026-01-01T00:00:00Z",
	}

	candidate := candidateFrom(record)

	if candidate.Id != "PG1" {
		t.Errorf("id = %q, want PG1", candidate.Id)
	}
	if candidate.Priority != 50 {
		t.Errorf("priority = %d, want 50", candidate.Priority)
	}
	if candidate.StackPolicy != string(models.PromotionStackExclusive) {
		t.Errorf("stack policy = %q", candidate.StackPolicy)
	}
	if candidate.ExclusiveGroup != "seasonal" {
		t.Errorf("exclusive group = %q, want seasonal", candidate.ExclusiveGroup)
	}
	if candidate.CreatedAt == "" {
		t.Error("created_at must be carried: it is the tie-break when priorities match (D-10)")
	}
}

// A priority that arrived through jsonb is a float64, not an int32. A reader accepting only int32
// would silently see every priority as zero, which would make conflict resolution order programs by
// their created_at alone — wrong, and invisible.
func TestPriorityIsReadFromEveryNumericShape(t *testing.T) {
	for name, value := range map[string]any{
		"int32":   int32(50),
		"int64":   int64(50),
		"int":     50,
		"float64": float64(50),
	} {
		t.Run(name, func(t *testing.T) {
			got := candidateFrom(dmodel.DynamicFields{
				models.SalesPromotionProgramFieldPriority: value,
			})
			if got.Priority != 50 {
				t.Errorf("priority from %s = %d, want 50", name, got.Priority)
			}
		})
	}
}

// A missing priority reads as zero rather than panicking. Zero is the strongest priority (lower
// first, D-10), which is the safe direction: a misconfigured program applies early and visibly
// rather than silently last.
func TestAbsentPriorityIsZero(t *testing.T) {
	if got := candidateFrom(dmodel.DynamicFields{}); got.Priority != 0 {
		t.Errorf("absent priority = %d, want 0", got.Priority)
	}
}

// The compatibility column is an enum, not a boolean, and only the exact string `allowed` may be
// read as permission. Anything else — `denied`, a typo, a value from a future migration — must
// refuse, because D-09 says denied wins over everything.
func TestOnlyAllowedIsReadAsPermission(t *testing.T) {
	cases := map[string]bool{
		string(models.PromotionCompatibilityAllowed): true,
		string(models.PromotionCompatibilityDenied):  false,
		"":              false,
		"ALLOWED":       false,
		"something_new": false,
	}
	for value, wantAllowed := range cases {
		got := value == string(models.PromotionCompatibilityAllowed)
		if got != wantAllowed {
			t.Errorf("compatibility %q read as allowed=%v, want %v", value, got, wantAllowed)
		}
	}
}

// A voucher that survives resolution alongside what is already applied is accepted, and nothing is
// displaced. The plain case, pinned so the interesting ones below are read as departures from it.
func TestAStackableVoucherJoinsWithoutDisplacing(t *testing.T) {
	voucher := CandidateProgram{
		Id: "VOUCHER", Priority: 10, CreatedAt: "2026-01-01",
		StackPolicy: string(models.PromotionStackStackable),
	}
	incumbent := CandidateProgram{
		Id: "AUTO", Priority: 20, CreatedAt: "2026-01-01",
		StackPolicy: string(models.PromotionStackStackable),
	}

	survivors := ResolvePromotions([]CandidateProgram{voucher, incumbent}, nil)

	if len(survivors) != 2 {
		t.Fatalf("both stackable programs must survive, got %d", len(survivors))
	}
}

// An exclusive voucher with the better priority DISPLACES what was already applied. That is why
// ApplyVoucher returns a full accepted list rather than a boolean: the caller must re-price against
// what survived, not append to what it had.
func TestAnExclusiveVoucherDisplacesAWeakerIncumbent(t *testing.T) {
	voucher := CandidateProgram{
		Id: "VOUCHER", Priority: 10, CreatedAt: "2026-01-01",
		StackPolicy: string(models.PromotionStackExclusive),
	}
	incumbent := CandidateProgram{
		Id: "AUTO", Priority: 20, CreatedAt: "2026-01-01",
		StackPolicy: string(models.PromotionStackStackable),
	}

	survivors := ResolvePromotions([]CandidateProgram{voucher, incumbent}, nil)

	if len(survivors) != 1 || survivors[0].Id != "VOUCHER" {
		t.Fatalf("the exclusive voucher must win and displace the incumbent, got %+v", survivors)
	}
}

// The other direction: an exclusive program already on the order refuses the voucher. This is the
// case ApplyVoucher reports as ReasonIncompatible, and it is decided by the same call — which is why
// the voucher and the incumbents go into resolution together rather than being compared pairwise.
func TestAnExclusiveIncumbentRefusesTheVoucher(t *testing.T) {
	incumbent := CandidateProgram{
		Id: "AUTO", Priority: 10, CreatedAt: "2026-01-01",
		StackPolicy: string(models.PromotionStackExclusive),
	}
	voucher := CandidateProgram{
		Id: "VOUCHER", Priority: 20, CreatedAt: "2026-01-01",
		StackPolicy: string(models.PromotionStackStackable),
	}

	survivors := ResolvePromotions([]CandidateProgram{voucher, incumbent}, nil)

	if len(survivors) != 1 || survivors[0].Id != "AUTO" {
		t.Fatalf("the exclusive incumbent must refuse the voucher, got %+v", survivors)
	}
}

// An explicit `denied` row beats stack policy, whichever direction it was written in (D-09).
func TestAnExplicitDenialRefusesTwoOtherwiseStackablePrograms(t *testing.T) {
	voucher := CandidateProgram{
		Id: "VOUCHER", Priority: 10, CreatedAt: "2026-01-01",
		StackPolicy: string(models.PromotionStackStackable),
	}
	incumbent := CandidateProgram{
		Id: "AUTO", Priority: 20, CreatedAt: "2026-01-01",
		StackPolicy: string(models.PromotionStackStackable),
	}

	// Written incumbent-first, to prove direction does not matter.
	rules := []CompatibilityRule{{ProgramAId: "AUTO", ProgramBId: "VOUCHER", Allowed: false}}

	survivors := ResolvePromotions([]CandidateProgram{voucher, incumbent}, rules)

	if len(survivors) != 1 || survivors[0].Id != "VOUCHER" {
		t.Fatalf("an explicit denial must refuse the pair, keeping the better priority; got %+v",
			survivors)
	}
}
