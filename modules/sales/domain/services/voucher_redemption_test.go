package services

import (
	"testing"
	"time"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The transition rules and the code-usability gates. Both are pure, so they are tested exhaustively
// here; the concurrency requirement of BR 30 needs a real database and lives in
// voucher_concurrency_test.go.

func TestReservationSettlesTwoWaysOnly(t *testing.T) {
	reserved := string(models.VoucherRedemptionStatusReserved)

	for _, to := range []string{
		string(models.VoucherRedemptionStatusRedeemed),
		string(models.VoucherRedemptionStatusReleased),
	} {
		if !CanTransitionVoucherRedemption(reserved, to) {
			t.Errorf("a reservation must be able to settle as %q", to)
		}
	}

	// Straight to reversed would mean refunding a use that was never taken.
	if CanTransitionVoucherRedemption(reserved, string(models.VoucherRedemptionStatusReversed)) {
		t.Error("a reservation must not be reversible: it was never redeemed")
	}
}

// A released or reversed redemption is finished. Reviving one would give away a use the code has
// already had back.
func TestTerminalRedemptionsCannotMove(t *testing.T) {
	for _, from := range []string{
		string(models.VoucherRedemptionStatusReleased),
		string(models.VoucherRedemptionStatusReversed),
	} {
		for _, to := range []string{
			string(models.VoucherRedemptionStatusReserved),
			string(models.VoucherRedemptionStatusRedeemed),
		} {
			if CanTransitionVoucherRedemption(from, to) {
				t.Errorf("%q is terminal but allows a move to %q", from, to)
			}
		}
	}
}

// Re-settling an already-settled redemption must not be an error: a retried confirm has asked for a
// state that already holds, and reporting failure would drive the caller to keep retrying.
func TestSettlingToTheSameStatusIsANoOp(t *testing.T) {
	for _, status := range []string{
		string(models.VoucherRedemptionStatusReserved),
		string(models.VoucherRedemptionStatusRedeemed),
		string(models.VoucherRedemptionStatusReleased),
		string(models.VoucherRedemptionStatusReversed),
	} {
		if !CanTransitionVoucherRedemption(status, status) {
			t.Errorf("%q must allow an idempotent re-settle", status)
		}
	}
}

// An unknown status refuses every move rather than panicking. The value came out of a database row
// and may predate a rename.
func TestUnknownRedemptionStatusRefusesEveryMove(t *testing.T) {
	if CanTransitionVoucherRedemption("something_else", string(models.VoucherRedemptionStatusRedeemed)) {
		t.Error("an unrecognised status must not be treated as reservable")
	}
}

// wasHolding is what decides whether settling gives a use back, so its answer for each status is
// worth pinning: a reservation holds one even though no sale has happened, which is the entire
// reason the reserve step exists.
func TestOnlyReservedAndRedeemedHoldAUse(t *testing.T) {
	cases := map[string]bool{
		string(models.VoucherRedemptionStatusReserved): true,
		string(models.VoucherRedemptionStatusRedeemed): true,
		string(models.VoucherRedemptionStatusReleased): false,
		string(models.VoucherRedemptionStatusReversed): false,
	}
	for status, want := range cases {
		if got := wasHolding(status); got != want {
			t.Errorf("wasHolding(%q) = %v, want %v", status, got, want)
		}
	}
}

func codeWith(fields map[string]any) *models.SalesVoucherCode {
	return models.NewSalesVoucherCodeFrom(fields)
}

func activeCode() map[string]any {
	return map[string]any{
		models.SalesVoucherCodeFieldStatus:     string(models.VoucherCodeStatusActive),
		models.SalesVoucherCodeFieldUsageCount: int32(0),
	}
}

func TestAnActiveUnlimitedCodeIsUsable(t *testing.T) {
	if vErrs := assertCodeUsable(codeWith(activeCode()), time.Now().Unix()); vErrs != nil {
		t.Errorf("an active code with no limit must be usable, got %v", vErrs.ToError())
	}
}

// A null usage_limit means unlimited. Reading a nil limit as zero would exhaust every unlimited
// voucher at its first use — the single most likely way to break every campaign at once.
func TestNilUsageLimitMeansUnlimited(t *testing.T) {
	fields := activeCode()
	fields[models.SalesVoucherCodeFieldUsageCount] = int32(9999)

	if !codeWith(fields).HasUsesRemaining() {
		t.Error("a code with no usage_limit must never run out")
	}
}

func TestUsesRunOutAtTheLimit(t *testing.T) {
	fields := activeCode()
	fields[models.SalesVoucherCodeFieldUsageLimit] = int32(3)

	for count, wantRemaining := range map[int32]bool{0: true, 2: true, 3: false, 4: false} {
		fields[models.SalesVoucherCodeFieldUsageCount] = count
		if got := codeWith(fields).HasUsesRemaining(); got != wantRemaining {
			t.Errorf("count %d of limit 3: HasUsesRemaining = %v, want %v",
				count, got, wantRemaining)
		}
	}
}

// Each refusal names its own reason, because BR 71 requires the till to be told WHICH thing was
// wrong — an expired code and an exhausted one call for different responses.
func TestEachRefusalNamesItsOwnReason(t *testing.T) {
	now := time.Now()
	past := model.ModelDateTime(now.Add(-time.Hour))
	future := model.ModelDateTime(now.Add(time.Hour))

	cases := []struct {
		name   string
		fields map[string]any
		reason string
	}{
		{"archived", map[string]any{
			models.SalesVoucherCodeFieldStatus:     string(models.VoucherCodeStatusActive),
			models.SalesVoucherCodeFieldIsArchived: true,
		}, ReasonArchived},

		{"disabled", map[string]any{
			models.SalesVoucherCodeFieldStatus: string(models.VoucherCodeStatusDisabled),
		}, ReasonDisabled},

		{"exhausted by status", map[string]any{
			models.SalesVoucherCodeFieldStatus: string(models.VoucherCodeStatusExhausted),
		}, ReasonUsageExhausted},

		{"exhausted by count", map[string]any{
			models.SalesVoucherCodeFieldStatus:     string(models.VoucherCodeStatusActive),
			models.SalesVoucherCodeFieldUsageLimit: int32(1),
			models.SalesVoucherCodeFieldUsageCount: int32(1),
		}, ReasonUsageExhausted},

		{"not yet valid", map[string]any{
			models.SalesVoucherCodeFieldStatus:    string(models.VoucherCodeStatusActive),
			models.SalesVoucherCodeFieldValidFrom: future,
		}, ReasonNotYetValid},

		{"expired", map[string]any{
			models.SalesVoucherCodeFieldStatus:     string(models.VoucherCodeStatusActive),
			models.SalesVoucherCodeFieldValidUntil: past,
		}, ReasonExpired},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			vErrs := assertCodeUsable(codeWith(testCase.fields), now.Unix())
			if vErrs == nil {
				t.Fatalf("%s must be refused", testCase.name)
			}
			// Assert on the Key rather than the message: the Key is what a till branches on and
			// what a translation layer renders, so it is the part of the contract that matters.
			if !hasReasonKey(vErrs, testCase.reason) {
				t.Errorf("refusal must carry key %q, got %v", testCase.reason, vErrs.ToError())
			}
		})
	}
}

// The window is half-open: valid_until is EXCLUSIVE, matching every other window in this module. A
// code valid until noon does not work at noon.
func TestValidUntilIsExclusive(t *testing.T) {
	boundary := time.Now().Truncate(time.Second)

	fields := activeCode()
	fields[models.SalesVoucherCodeFieldValidUntil] = model.ModelDateTime(boundary)

	if assertCodeUsable(codeWith(fields), boundary.Add(-time.Second).Unix()) != nil {
		t.Error("a second before valid_until the code must work")
	}
	if assertCodeUsable(codeWith(fields), boundary.Unix()) == nil {
		t.Error("at valid_until exactly the code must NOT work: the bound is exclusive")
	}
}

// valid_from is INCLUSIVE, the other half of the same window.
func TestValidFromIsInclusive(t *testing.T) {
	boundary := time.Now().Truncate(time.Second)

	fields := activeCode()
	fields[models.SalesVoucherCodeFieldValidFrom] = model.ModelDateTime(boundary)

	if assertCodeUsable(codeWith(fields), boundary.Add(-time.Second).Unix()) == nil {
		t.Error("a second before valid_from the code must not work yet")
	}
	if assertCodeUsable(codeWith(fields), boundary.Unix()) != nil {
		t.Error("at valid_from exactly the code must work: the bound is inclusive")
	}
}

// A code with no status at all is refused rather than treated as active. The field is required, so
// a missing one means a row written outside the sanctioned path.
func TestAStatuslessCodeIsRefused(t *testing.T) {
	if assertCodeUsable(codeWith(map[string]any{}), time.Now().Unix()) == nil {
		t.Error("a code with no status must not be usable")
	}
}

// hasReasonKey reports whether any violation carries the given translation key.
func hasReasonKey(vErrs *ft.ClientErrors, key string) bool {
	if vErrs == nil {
		return false
	}
	for _, item := range *vErrs {
		if item.Key == key {
			return true
		}
	}
	return false
}
