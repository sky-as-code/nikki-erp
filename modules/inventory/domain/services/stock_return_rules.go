package services

import (
	"github.com/shopspring/decimal"

	ft "github.com/sky-as-code/nikki-erp/common/fault"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// The arithmetic behind a return, with no repository so it tests without a database.
//
// A return is capped at what is still returnable, absolutely: nothing here takes a waiver
// parameter, deliberately. Any future override must be an authorisation decision at the call site,
// not an argument threaded through the arithmetic where every caller could reach it.

// ReturnableLine is one move's contribution to what may still be sent back.
type ReturnableLine struct {
	MoveId           string
	ProductVariantId string
	// Completed is what the original move actually executed, not what it demanded.
	Completed decimal.Decimal
	// AlreadyReturned is the sum of done return lines raised against this move.
	AlreadyReturned decimal.Decimal
}

// Returnable is what may still be sent back for this move. Computed from what was COMPLETED, never
// from demand_quantity: a transfer of 100 that shipped only 80 can have 80 returned, because the
// other 20 never left.
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

// AssertReturnable refuses a requested quantity exceeding what is still returnable. It names both
// numbers, since the caller's next step depends on the size of the gap.
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

// TotalReturnable sums what may still be sent back across every line of a transfer, so a fully
// returned transfer is refused with one clear message rather than a list of per-move zeroes.
func TotalReturnable(lines []ReturnableLine) decimal.Decimal {
	total := decimal.Zero
	for _, line := range lines {
		total = total.Add(line.Returnable())
	}
	return total
}
