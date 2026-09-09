package services

import (
	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The adjustment order: what the customer kept, after part of a sale came back.
//
// WHY A NEW ORDER RATHER THAN AN EDIT. The original sale really did sell what it sold. It was paid
// for, reported on a VAT invoice, and counted in a day's takings, so rewriting its lines would
// restate a transaction that other records already describe — and leave the returned goods
// unaccounted for. The original stays exactly as it was and gains one link saying it has been
// superseded; the adjustment order beside it states what the customer walked away with.
//
// PRICES ARE COPIED, NEVER RECALCULATED. The kept lines carry the effective unit price, the tax rate
// and the amounts the customer actually paid. Running them through pricing again would apply today's
// promotions and today's tax master to a sale that happened last week, and the adjustment order
// would state a sum nobody was ever charged.
//
// IT CARRIES NO BILLS AND NO PAYMENTS. Payment history cannot be meaningfully split — one card
// payment settled a bill covering both the kept and the returned goods — so the original's bills,
// payments and refund legs remain the financial record. The adjustment order is the operational view
// of what the customer has, not a second claim on their money.

// CreateAdjustmentOrderResult reports what was created, if anything.
type CreateAdjustmentOrderResult struct {
	// AdjustmentOrderId is empty when no adjustment order was needed — nothing was kept, or one
	// already exists for this return.
	AdjustmentOrderId string

	Created bool
}

// CreateAdjustmentOrder records what the customer kept after a partial return.
//
// Called when a return completes. A full return creates nothing: with no lines left there is nothing
// to restate, and an empty order would be a document about nothing.
func CreateAdjustmentOrder(
	ctx corectx.Context, salesReturn dmodel.DynamicFields,
) (*CreateAdjustmentOrderResult, error) {
	orderId := stringOf(salesReturn, models.SalesReturnFieldSalesOrderId)

	original, err := loadRecord(ctx, models.SalesOrderSchemaName,
		models.SalesOrderFieldId, orderId)
	if err != nil || original == nil {
		return &CreateAdjustmentOrderResult{}, err
	}

	// Already superseded: this return has been processed before, or another return got there first.
	// Answered rather than refused — the caller's work is done either way, and creating a second
	// adjustment order would leave two documents claiming to say what the customer kept.
	if models.NewSalesOrderFrom(original).IsSuperseded() {
		return &CreateAdjustmentOrderResult{}, nil
	}

	kept, err := keptLinesOf(ctx, orderId)
	if err != nil {
		return nil, err
	}
	if len(kept) == 0 {
		// Everything came back. There is nothing to restate.
		return &CreateAdjustmentOrderResult{}, nil
	}

	adjustmentId, err := model.NewId()
	if err != nil {
		return nil, err
	}

	// One transaction: the new order, its lines, and the link on the original. A crash between them
	// would leave an order superseded by something that does not exist, or an adjustment order
	// nothing points at.
	err = withTransaction(ctx, models.SalesOrderSchemaName, func(tranxCtx corectx.Context) error {
		if err := writeAdjustmentOrder(
			tranxCtx, original, salesReturn, string(*adjustmentId), kept); err != nil {
			return err
		}
		return writeChanges(tranxCtx, models.SalesOrderSchemaName, original,
			dmodel.DynamicFields{
				models.SalesOrderFieldAdjustedByOrderId: string(*adjustmentId),
			})
	})
	if err != nil {
		return nil, err
	}

	return &CreateAdjustmentOrderResult{
		AdjustmentOrderId: string(*adjustmentId),
		Created:           true,
	}, nil
}

// keptLine is one order line and how much of it the customer still has.
type keptLine struct {
	original dmodel.DynamicFields
	quantity decimal.Decimal
}

// keptLinesOf works out what is left of each line after everything that came back.
//
// It reads returned_quantity, which the return process writes when it completes — so this must run
// after that, not before, or every line would still look untouched.
func keptLinesOf(ctx corectx.Context, orderId string) ([]keptLine, error) {
	lines, err := searchBy(ctx, models.SalesOrderLineSchemaName,
		models.SalesOrderLineFieldSalesOrderId, orderId)
	if err != nil {
		return nil, err
	}

	kept := make([]keptLine, 0, len(lines))
	for _, line := range lines {
		remaining := decimalOf(line, models.SalesOrderLineFieldOrderedQuantity).
			Sub(decimalOf(line, models.SalesOrderLineFieldReturnedQuantity))
		if !remaining.IsPositive() {
			continue
		}
		kept = append(kept, keptLine{original: line, quantity: remaining})
	}
	return kept, nil
}

// writeAdjustmentOrder creates the order and its lines.
func writeAdjustmentOrder(
	ctx corectx.Context,
	original dmodel.DynamicFields,
	salesReturn dmodel.DynamicFields,
	adjustmentId string,
	kept []keptLine,
) error {
	engine, err := engineFor(models.SalesOrderSchemaName)
	if err != nil {
		return err
	}

	totals := adjustmentTotalsOf(kept)
	orgId := stringOf(original, basemodel.FieldOrgId)

	fields := dmodel.DynamicFields{
		models.SalesOrderFieldId:             adjustmentId,
		models.SalesOrderFieldAdjustsOrderId: stringOf(original, models.SalesOrderFieldId),

		// Confirmed, not draft: the sale it describes already happened and was paid for. A draft
		// would be swept away by the expiry job as an abandoned basket.
		models.SalesOrderFieldStatus: string(models.SalesOrderStatusConfirmed),

		models.SalesOrderFieldSalesChannelId: stringOf(original, models.SalesOrderFieldSalesChannelId),
		models.SalesOrderFieldSalesPointId:   stringOf(original, models.SalesOrderFieldSalesPointId),
		models.SalesOrderFieldCurrencyCode:   stringOf(original, models.SalesOrderFieldCurrencyCode),

		models.SalesOrderFieldSubtotal:      totals.subtotal,
		models.SalesOrderFieldDiscountTotal: totals.discount,
		models.SalesOrderFieldTaxTotal:      totals.tax,
		models.SalesOrderFieldGrandTotal:    totals.grand,

		// Derived from the return, so processing the same return twice cannot produce two adjustment
		// orders even if the supersession check above were somehow bypassed.
		models.SalesOrderFieldIdempotencyKey: "adj:" + stringOf(salesReturn, models.SalesReturnFieldId),

		basemodel.FieldOrgId: orgId,
	}

	// The parties and the customer carry over: the same person kept the goods.
	for _, field := range []string{
		models.SalesOrderFieldCustomerReference,
		models.SalesOrderFieldSoldToPartyId,
		models.SalesOrderFieldBillToPartyId,
		models.SalesOrderFieldPayerPartyId,
	} {
		if value := stringOf(original, field); value != "" {
			fields[field] = value
		}
	}

	if _, err := engine.ResourceRepository().Insert(ctx, fields); err != nil {
		return err
	}
	return writeAdjustmentOrderLines(ctx, adjustmentId, orgId, kept)
}

// adjustmentTotals is what the kept lines come to.
type adjustmentTotals struct {
	subtotal decimal.Decimal
	discount decimal.Decimal
	tax      decimal.Decimal
	grand    decimal.Decimal
}

// adjustmentTotalsOf sums the kept lines.
//
// Each line's amounts are prorated by the fraction kept, at the historical unit price, so the
// adjustment order's totals are the part of the original the customer did not send back — never a
// fresh calculation.
func adjustmentTotalsOf(kept []keptLine) adjustmentTotals {
	totals := adjustmentTotals{
		subtotal: decimal.Zero,
		discount: decimal.Zero,
		tax:      decimal.Zero,
		grand:    decimal.Zero,
	}

	for _, line := range kept {
		amounts := proratedLineAmounts(line)
		totals.subtotal = totals.subtotal.Add(amounts.gross)
		totals.discount = totals.discount.Add(amounts.discount)
		totals.tax = totals.tax.Add(amounts.tax)
		totals.grand = totals.grand.Add(amounts.final)
	}
	return totals
}

// proratedAmounts is one kept line's share of its original amounts.
type proratedAmounts struct {
	gross    decimal.Decimal
	discount decimal.Decimal
	net      decimal.Decimal
	tax      decimal.Decimal
	final    decimal.Decimal
}

// proratedLineAmounts scales a line's amounts by the fraction the customer kept.
//
// Prorated rather than recomputed from unit price × quantity, because a line's amounts carry
// promotions and rounding that unit price alone cannot reproduce: a buy-two-get-one line's discount
// is not two thirds of itself when one unit comes back, but proration is the only division of it
// that both halves agree on and that sums back to the original.
func proratedLineAmounts(line keptLine) proratedAmounts {
	ordered := decimalOf(line.original, models.SalesOrderLineFieldOrderedQuantity)
	if !ordered.IsPositive() {
		return proratedAmounts{
			gross: decimal.Zero, discount: decimal.Zero,
			net: decimal.Zero, tax: decimal.Zero, final: decimal.Zero,
		}
	}

	fraction := line.quantity.Div(ordered)
	return proratedAmounts{
		gross:    decimalOf(line.original, models.SalesOrderLineFieldGrossAmount).Mul(fraction),
		discount: decimalOf(line.original, models.SalesOrderLineFieldDiscountAmount).Mul(fraction),
		net:      decimalOf(line.original, models.SalesOrderLineFieldNetAmount).Mul(fraction),
		tax:      decimalOf(line.original, models.SalesOrderLineFieldTaxAmount).Mul(fraction),
		final:    decimalOf(line.original, models.SalesOrderLineFieldFinalAmount).Mul(fraction),
	}
}

// writeAdjustmentOrderLines copies the kept lines onto the new order.
func writeAdjustmentOrderLines(
	ctx corectx.Context, adjustmentId, orgId string, kept []keptLine,
) error {
	engine, err := engineFor(models.SalesOrderLineSchemaName)
	if err != nil {
		return err
	}

	for index, line := range kept {
		id, err := model.NewId()
		if err != nil {
			return err
		}
		amounts := proratedLineAmounts(line)

		fields := dmodel.DynamicFields{
			models.SalesOrderLineFieldId:           string(*id),
			models.SalesOrderLineFieldSalesOrderId: adjustmentId,
			models.SalesOrderLineFieldLineNumber:   int32(index + 1),

			models.SalesOrderLineFieldOrderedQuantity: line.quantity,

			// Nothing has come back from THIS order, and the goods are already with the customer.
			models.SalesOrderLineFieldReturnedQuantity:  decimal.Zero,
			models.SalesOrderLineFieldFulfilledQuantity: line.quantity,

			// The prices the customer actually paid, copied rather than re-derived.
			models.SalesOrderLineFieldBaseUnitPrice: decimalOf(
				line.original, models.SalesOrderLineFieldBaseUnitPrice),
			models.SalesOrderLineFieldEffectiveUnitPrice: decimalOf(
				line.original, models.SalesOrderLineFieldEffectiveUnitPrice),
			models.SalesOrderLineFieldTaxRateSnapshot: decimalOf(
				line.original, models.SalesOrderLineFieldTaxRateSnapshot),

			models.SalesOrderLineFieldGrossAmount:    amounts.gross,
			models.SalesOrderLineFieldDiscountAmount: amounts.discount,
			models.SalesOrderLineFieldNetAmount:      amounts.net,
			models.SalesOrderLineFieldTaxAmount:      amounts.tax,
			models.SalesOrderLineFieldFinalAmount:    amounts.final,

			basemodel.FieldOrgId: orgId,
		}

		// The descriptive fields carry over verbatim: a receipt for what the customer kept must name
		// the product as it was named when they bought it.
		for _, field := range []string{
			models.SalesOrderLineFieldLineType,
			models.SalesOrderLineFieldProductVariantId,
			models.SalesOrderLineFieldProductCodeSnapshot,
			models.SalesOrderLineFieldProductNameSnapshot,
			models.SalesOrderLineFieldUomId,
			models.SalesOrderLineFieldPricingSource,
			models.SalesOrderLineFieldSalesComboId,
		} {
			if value := stringOf(line.original, field); value != "" {
				fields[field] = value
			}
		}

		if _, err := engine.ResourceRepository().Insert(ctx, fields); err != nil {
			return err
		}
	}
	return nil
}
