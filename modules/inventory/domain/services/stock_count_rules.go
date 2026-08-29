package services

import (
	"time"

	"github.com/shopspring/decimal"

	ft "github.com/sky-as-code/nikki-erp/common/fault"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// The pure rules behind a physical inventory count. None touches a repository, so all test without
// a database.

// CountVariance is counted minus system: positive means the adjustment gains stock, negative means
// it loses stock.
func CountVariance(counted, system decimal.Decimal) decimal.Decimal {
	return counted.Sub(system)
}

// IsCountSnapshotStale reports whether the balance moved since the count was entered. A count is a
// statement about the balance it was taken against, so applying its variance to a balance that has
// since changed gives a wrong figure; it must be recounted rather than reinterpreted.
//
// Callers must pass an on-hand read inside the row lock; a value read before the lock is stale by
// definition and would reproduce the race this prevents.
func IsCountSnapshotStale(snapshot, currentOnHand decimal.Decimal) bool {
	return !snapshot.Equal(currentOnHand)
}

// AssertCountEnterable refuses a negative counted quantity. Zero is allowed and meaningful — an
// empty shelf is a legitimate count result, which is why count_quantity_set is a separate flag.
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

// AssertCountApplicable refuses an apply with no pending count. The flag is the authority, never
// the value: testing `counted_quantity != 0` would silently refuse every count of an empty shelf.
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

// StaleCountErrors is the refusal an apply reports when the snapshot no longer matches. It names
// both numbers, since the counter's next action depends on the size of the difference.
func StaleCountErrors(snapshot, currentOnHand decimal.Decimal) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(
		models.StockQuantSchemaName,
		"stock_quant.count_snapshot_stale",
		"the balance was "+snapshot.String()+" when this count was entered and is now "+
			currentOnHand.String()+"; the count must be entered again against the current balance"))
	return vErrs
}

// CountEntryFields is the metadata written when a count is entered. The snapshot must be taken
// here, not at apply time: taken at apply it would always match, silently disabling the staleness
// check.
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

// CountResetFields clears a pending count without touching the balance. The scheduling fields are
// left alone: a balance whose count was reset is still due to be counted.
func CountResetFields() map[string]any {
	return map[string]any{
		models.StockQuantFieldCountedQuantity:  nil,
		models.StockQuantFieldCountQuantitySet: false,
		models.StockQuantFieldCountSnapshotQty: nil,
		models.StockQuantFieldCountReasonCode:  "",
		models.StockQuantFieldCountReasonText:  "",
	}
}

// CountAppliedFields clears the pending count and stamps the counting history. It runs whether or
// not a movement was generated: a zero variance is a successful apply, and skipping the stamp would
// leave those balances permanently overdue.
func CountAppliedFields(now time.Time, nextCountDate *time.Time) map[string]any {
	fields := CountResetFields()
	fields[models.StockQuantFieldLastCountDate] = now
	if nextCountDate != nil {
		fields[models.StockQuantFieldNextCountDate] = *nextCountDate
	}
	return fields
}
