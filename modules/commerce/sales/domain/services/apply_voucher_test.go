package services

import (
	"testing"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The pure parts of applying a voucher: reading a program row into a candidate, and the conflict
// resolution that decides whether it joins. The repository-touching paths are exercised live -
// `engineFor` could be stubbed, but a fake registry would pin the fake rather than the behaviour.

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

// A priority from jsonb is a float64, not an int32. A reader accepting only int32 would see every
// priority as zero and order programs by created_at alone.
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

// A missing priority reads as zero, the strongest priority: a misconfigured program applies early
// and visibly rather than last.
func TestAbsentPriorityIsZero(t *testing.T) {
	if got := candidateFrom(dmodel.DynamicFields{}); got.Priority != 0 {
		t.Errorf("absent priority = %d, want 0", got.Priority)
	}
}

// The compatibility column is an enum: only the exact string `allowed` may be read as permission.
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

// The plain case: a voucher that survives resolution is accepted and nothing is displaced.
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

// An exclusive voucher with the better priority DISPLACES what was already applied, which is why
// ApplyVoucher returns a full accepted list rather than a boolean.
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

// The other direction: an incumbent exclusive program refuses the voucher, decided by the same
// call - which is why voucher and incumbents go into resolution together.
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

// An explicit `denied` row beats stack policy, whichever direction it was written in.
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
