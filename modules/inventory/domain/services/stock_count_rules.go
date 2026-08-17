package services

import (
	"time"

	"github.com/shopspring/decimal"

	ft "github.com/sky-as-code/nikki-erp/common/fault"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// The pure rules behind a physical inventory count: what a variance is, when a snapshot has gone
// stale, and what a count's metadata looks like at each step. None of them touches a repository,
// so all of them test without a database — the same split Phase 1 used for stock_quant_rules.go.

// CountVariance is the difference a count found: counted minus system.
//
// Positive means the shelf held more than the books said and the adjustment gains stock; negative
// means it held less and the adjustment loses it. Returning a signed number rather than a
// magnitude plus a direction flag keeps the sign in one place, where the caller can see it.
func CountVariance(counted, system decimal.Decimal) decimal.Decimal {
	return counted.Sub(system)
}

// IsCountSnapshotStale reports whether the balance has moved since the count was entered.
//
// This is STOCK-INV-008 and AC-STOCK-015, and the worked example in BR §4.2.7.4 is what it exists
// for: system says 100, a counter finds 97, a delivery of 10 lands before the count is applied.
// Applying the −3 variance to the new balance of 90 would give 87, when the shelf actually holds
// 107. The count was a statement about a balance that no longer exists, so it must be refused and
// recounted rather than reinterpreted.
//
// The caller must pass an on-hand read *inside* the row lock. A value read before the lock is
// stale by definition, and comparing it here would reproduce exactly the race this prevents.
func IsCountSnapshotStale(snapshot, currentOnHand decimal.Decimal) bool {
	return !snapshot.Equal(currentOnHand)
}

// AssertCountEnterable applies the rules entering a count must satisfy.
//
// A negative counted quantity is refused: a count is a statement of what is physically present,
// and nothing physical is present in a negative amount. Zero is allowed and is meaningful — "the
// shelf is empty" is a legitimate count result, which is exactly why count_quantity_set exists as
// a separate flag.
func AssertCountEnterable(counted decimal.Decimal) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	if counted.LessThan(decimal.Zero) {
		vErrs.Append(*ft.NewBusinessViolation(
			models.StockQuantSchemaName,
			"stock_quant.counted_quantity_negative",
			"a counted quantity cannot be negative"))
	}
	return vErrs
}

// AssertCountApplicable refuses an apply with no pending count behind it.
//
// The flag is the authority, never the value: a counted quantity of zero is a legitimate result,
// so testing `counted_quantity != 0` as a proxy would silently refuse every count that found an
// empty shelf — the one case a counter most needs recorded.
func AssertCountApplicable(countQuantitySet bool) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	if !countQuantitySet {
		vErrs.Append(*ft.NewBusinessViolation(
			models.StockQuantSchemaName,
			"stock_quant.no_pending_count",
			"this balance has no counted quantity awaiting apply"))
	}
	return vErrs
}

// StaleCountErrors is the refusal an apply reports when the snapshot no longer matches.
//
// It names both numbers, because the counter's next action depends on the difference: a small
// discrepancy is a recount, a large one is worth investigating before recounting.
func StaleCountErrors(snapshot, currentOnHand decimal.Decimal) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(
		models.StockQuantSchemaName,
		"stock_quant.count_snapshot_stale",
		"the balance was "+snapshot.String()+" when this count was entered and is now "+
			currentOnHand.String()+"; the count must be entered again against the current balance"))
	return vErrs
}

// CountEntryFields is the metadata written when a count is entered.
//
// The snapshot is taken here, not at apply time: it is the whole basis of the staleness check, and
// a snapshot taken at apply would always match, turning the check into a no-op that looks like it
// works.
func CountEntryFields(
	counted, snapshot decimal.Decimal, reasonCode, reasonText string,
) map[string]any {
	return map[string]any{
		models.StockQuantFieldCountedQuantity:  counted.String(),
		models.StockQuantFieldCountQuantitySet: true,
		models.StockQuantFieldCountSnapshotQty: snapshot.String(),
		models.StockQuantFieldCountReasonCode:  reasonCode,
		models.StockQuantFieldCountReasonText:  reasonText,
	}
}

// CountResetFields clears a pending count without touching the balance.
//
// Reset is not a correction: it abandons a count that was entered wrongly, leaving the on-hand
// quantity exactly as it was (BR §4.2.7.6). The scheduling fields are deliberately untouched — a
// balance whose count was reset is still due to be counted.
func CountResetFields() map[string]any {
	return map[string]any{
		models.StockQuantFieldCountedQuantity:  nil,
		models.StockQuantFieldCountQuantitySet: false,
		models.StockQuantFieldCountSnapshotQty: nil,
		models.StockQuantFieldCountReasonCode:  "",
		models.StockQuantFieldCountReasonText:  "",
	}
}

// CountAppliedFields clears the pending count and stamps the counting history.
//
// It runs whether or not a movement was generated. BR §4.2.7.5 is explicit that a zero variance is
// a successful apply rather than a no-op to skip: a counter who confirms the balance was right has
// done their job, and the worklist must stop asking. Skipping the stamp on zero variance would
// leave those balances permanently overdue.
func CountAppliedFields(now time.Time, nextCountDate *time.Time) map[string]any {
	fields := CountResetFields()
	fields[models.StockQuantFieldLastCountDate] = now
	if nextCountDate != nil {
		fields[models.StockQuantFieldNextCountDate] = *nextCountDate
	}
	return fields
}
