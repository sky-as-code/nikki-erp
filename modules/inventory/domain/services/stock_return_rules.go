package services

import (
	"github.com/shopspring/decimal"

	ft "github.com/sky-as-code/nikki-erp/common/fault"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// The arithmetic behind a return, with no repository in sight so it tests without a database.
//
// AC-STOCK-022 caps a return at what is still returnable, and in this phase that cap is absolute:
// there is no override, by any caller, at any permission level. See [INV-STK-307] for the note on
// the supervisor override the AC anticipates. Nothing here takes a waiver parameter, deliberately —
// if an override is ever built it must be an authorisation decision at the call site, not an
// argument threaded through the arithmetic where every caller could reach it.

// ReturnableLine is one move's contribution to what may still be sent back.
type ReturnableLine struct {
	MoveId           string
	ProductVariantId string
	// Completed is what the original move actually executed, not what it demanded.
	Completed decimal.Decimal
	// AlreadyReturned is the sum of done return lines raised against this move.
	AlreadyReturned decimal.Decimal
}

// Returnable is what may still be sent back for this move.
//
// BR §4.2.10.3 is explicit that it is computed from what was *completed*, never from
// demand_quantity: a transfer of 100 that shipped only 80 can have 80 returned, because the other
// 20 never left. This is the whole reason Phase 2 kept demand and execution as separate layers
// (STOCK-INV-019), and spending that separation correctly here is the point of the function.
//
// It floors at zero. A negative would mean more has been returned than ever shipped, which is a
// data problem rather than a licence to ship stock back out.
func (this ReturnableLine) Returnable() decimal.Decimal {
	remaining := this.Completed.Sub(this.AlreadyReturned)
	if remaining.LessThan(decimal.Zero) {
		return decimal.Zero
	}
	return remaining
}

// AssertReturnable refuses a requested quantity that exceeds what is still returnable.
//
// The refusal names both numbers because the caller's next step depends on the gap: a request for
// slightly too much is usually a typo, while a request far beyond the shipment suggests the wrong
// transfer was chosen.
func AssertReturnable(line ReturnableLine, requested decimal.Decimal) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()

	if requested.LessThanOrEqual(decimal.Zero) {
		vErrs.Append(*ft.NewBusinessViolation(
			models.StockMoveSchemaName, "stock_return.quantity_not_positive",
			"a return must send back a quantity greater than zero"))
		return vErrs
	}

	returnable := line.Returnable()
	if requested.GreaterThan(returnable) {
		vErrs.Append(*ft.NewBusinessViolation(
			models.StockMoveSchemaName, "stock_return.exceeds_returnable",
			"only "+returnable.String()+" of move '"+line.MoveId+
				"' is still returnable, which is less than the "+requested.String()+" requested"))
	}
	return vErrs
}

// TotalReturnable sums what may still be sent back across every line of a transfer.
//
// Used to refuse a return of a transfer that has already been returned in full, which is worth
// catching before any per-line arithmetic: the message "nothing is left to return" is far clearer
// than a list of per-move zeroes.
func TotalReturnable(lines []ReturnableLine) decimal.Decimal {
	total := decimal.Zero
	for _, line := range lines {
		total = total.Add(line.Returnable())
	}
	return total
}
