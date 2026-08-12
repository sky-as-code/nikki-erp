package services

import (
	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
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

// FillAvailableQuantity writes the derived available quantity onto a quant's field bag.
//
// The engine serves `available_quantity` as a virtual field, which means it has no column and is
// absent from every row the repository returns. Without this the field would be advertised in
// meta/schema and then always read as null.
func FillAvailableQuantity(fields dmodel.DynamicFields) {
	quant := models.NewStockQuantFrom(fields)
	available := AvailableQuantity(quant.GetOnHandQuantity(), quant.GetReservedQuantity())
	quant.SetAvailableQuantity(&available)
}

// availableQuantityOperands are the columns the derived quantity is computed from. A projection
// that names available_quantity must carry them, or there is nothing to subtract.
func availableQuantityOperands() []string {
	return []string{
		models.StockQuantFieldOnHandQuantity,
		models.StockQuantFieldReservedQuantity,
	}
}

// wantsAvailableQuantity reports whether a read should compute the derived quantity. An empty
// projection means "the resource's default field set", which includes it.
func wantsAvailableQuantity(requested []string) bool {
	if len(requested) == 0 {
		return true
	}
	for _, field := range requested {
		if field == models.StockQuantFieldAvailableQuantity {
			return true
		}
	}
	return false
}
