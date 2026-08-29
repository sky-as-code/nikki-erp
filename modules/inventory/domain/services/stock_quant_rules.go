package services

import (
	"github.com/shopspring/decimal"

	ft "github.com/sky-as-code/nikki-erp/common/fault"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// AvailableQuantity is the stock a new demand may still claim. Derived rather than stored, so it
// cannot disagree with the two numbers behind it. A nil operand reads as zero.
func AvailableQuantity(onHand, reserved *decimal.Decimal) decimal.Decimal {
	return orZero(onHand).Sub(orZero(reserved))
}

func orZero(value *decimal.Decimal) decimal.Decimal {
	if value == nil {
		return decimal.Zero
	}
	return *value
}

// AssertQuantNotClientWritable refuses any client attempt to create, update or delete a quant: its
// on-hand quantity is the running total of completed movements, so a client write would be a
// balance change with nothing to audit. Corrections go through an adjustment, transfer or scrap.
//
// It appends a client error rather than returning one, because the engine renders a returned error
// as a server fault and a failed validation as a 400.
func AssertQuantNotClientWritable(vErrs *ft.ClientErrors) {
	vErrs.Append(*ft.NewBusinessViolation(
		models.StockQuantSchemaName,
		"stock_quant.not_client_writable",
		"stock balances cannot be changed directly; record an inventory adjustment, transfer or scrap instead",
	))
}

// available_quantity is served by the engine's computed-field layer from its declaration in
// stock_quant.json. AvailableQuantity above is the pure rule that formula and the reservation logic
// both express.
