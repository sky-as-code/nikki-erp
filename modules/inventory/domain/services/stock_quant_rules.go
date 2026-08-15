package services

import (
	"github.com/shopspring/decimal"

	ft "github.com/sky-as-code/nikki-erp/common/fault"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// AvailableQuantity is the stock a new demand may still claim.
//
// It is derived rather than stored, so that it cannot disagree with the two numbers it comes
// from. A nil operand reads as zero: a quant whose reserved quantity has never been written has
// reserved nothing, which is not the same as having an unknown balance. See BR §4.2.2.3.
func AvailableQuantity(onHand, reserved *decimal.Decimal) decimal.Decimal {
	return orZero(onHand).Sub(orZero(reserved))
}

func orZero(value *decimal.Decimal) decimal.Decimal {
	if value == nil {
		return decimal.Zero
	}
	return *value
}

// AssertQuantNotClientWritable refuses any client attempt to create, update or delete a quant.
//
// A quant is current state, not a document: its on-hand quantity is the running total of the
// movements that completed against it, so a client write would be a balance change with no
// movement behind it and nothing to audit. Corrections go through an inventory adjustment, a
// transfer or a scrap, each of which records the movement that justifies the new number.
// See BR §3.3, §4.2.2.6 and AC-STOCK-002.
//
// It appends a client error rather than returning one, because the engine reports a failed
// validation to the caller as a 400 and treats a returned error as a server fault.
func AssertQuantNotClientWritable(vErrs *ft.ClientErrors) {
	vErrs.Append(*ft.NewBusinessViolation(
		models.StockQuantSchemaName,
		"stock_quant.not_client_writable",
		"stock balances cannot be changed directly; record an inventory adjustment, transfer or scrap instead",
	))
}

// available_quantity is served by the engine's computed-field layer: the schema declares it in
// stock_quant.json, so no fill/projection helpers live here anymore. AvailableQuantity above
// stays as the pure rule the reservation logic (and the schema's formula) both express.
