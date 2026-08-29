package services

import (
	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"

	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

// Line and order money fields are stored, never computed at read time, so an order keeps the value
// the business agreed to rather than one derived from today's prices and rounding rules. Where a
// stored header disagrees with its lines, the lines win.

// defaultScale is the decimal places money rounds to when the order's currency cannot be resolved
// (none set yet, or lookup failed). The real scale comes from the currency
// (OrderReferenceValidator.ScaleFor): 0 for VND, 2 for USD, 3 for KWD. Two is the fallback because
// most currencies use it and an order with no currency yet is an ordinary draft, not an error.
const defaultScale int32 = 2

// orderScaleResolver resolves an order's rounding scale from its currency. It is a package-level
// hook rather than a threaded parameter because the recompute is reached from four call sites and
// none of them holds a currency port; when unset, every order rounds to defaultScale.
var orderScaleResolver func(ctx corectx.Context, currencyId string) int32

// SetOrderScaleResolver installs the currency-aware scale lookup. Called once during Init.
func SetOrderScaleResolver(resolve func(ctx corectx.Context, currencyId string) int32) {
	orderScaleResolver = resolve
}

func scaleForOrder(ctx corectx.Context, order dmodel.DynamicFields) int32 {
	if orderScaleResolver == nil {
		return defaultScale
	}
	return orderScaleResolver(ctx, stringOf(order, models.PurchaseOrderFieldCurrencyId))
}

// OrderTotals is the money summary of one order, computed from its lines.
type OrderTotals struct {
	Untaxed decimal.Decimal
	Tax     decimal.Decimal
	Total   decimal.Decimal
}

// LineTotals is the money summary of one line.
type LineTotals struct {
	Subtotal decimal.Decimal
	Tax      decimal.Decimal
	Total    decimal.Decimal
}

// ComputeLineTotals derives one line's money fields: subtotal = quantity x unit_price less
// discount_percent, total = subtotal + tax_amount. tax_amount is an input, not computed here —
// there is no tax engine in this module, so the client supplies the line's tax and the server only
// sums it. Non-product lines (section, subsection, note) contribute nothing.
func ComputeLineTotals(line dmodel.DynamicFields, scale int32) LineTotals {
	if !isMoneyBearingLine(line) {
		return LineTotals{Subtotal: decimal.Zero, Tax: decimal.Zero, Total: decimal.Zero}
	}

	quantity := decimalOf(line, models.PurchaseOrderLineFieldQuantity)
	unitPrice := decimalOf(line, models.PurchaseOrderLineFieldUnitPrice)
	discount := decimalOf(line, models.PurchaseOrderLineFieldDiscountPercent)
	tax := decimalOf(line, models.PurchaseOrderLineFieldTaxAmount)

	gross := quantity.Mul(unitPrice)
	// discount_percent is a percentage, and is applied to the whole line rather than to the unit
	// price so it is rounded once instead of once per unit.
	kept := decimal.NewFromInt(100).Sub(discount).Div(decimal.NewFromInt(100))
	subtotal := gross.Mul(kept).Round(scale)
	tax = tax.Round(scale)

	return LineTotals{
		Subtotal: subtotal,
		Tax:      tax,
		Total:    subtotal.Add(tax),
	}
}

// isMoneyBearingLine reports whether a line contributes to the order's totals. A missing or
// unrecognised line type deliberately defaults to true: dropping money silently out of a total is
// the worse failure.
func isMoneyBearingLine(line dmodel.DynamicFields) bool {
	switch stringOf(line, models.PurchaseOrderLineFieldLineType) {
	case string(models.PurchaseOrderLineTypeSection),
		string(models.PurchaseOrderLineTypeSubsection),
		string(models.PurchaseOrderLineTypeNote):
		return false
	default:
		return true
	}
}

// ComputeOrderTotals sums the lines' stored subtotal and tax rather than recomputing them from
// their inputs, so the header always equals the sum of what the lines display; deriving it
// independently could differ by a rounding step. Callers must have made the stored line values
// current first (see RecomputeOrderTotals).
func ComputeOrderTotals(lines []dmodel.DynamicFields) OrderTotals {
	totals := OrderTotals{Untaxed: decimal.Zero, Tax: decimal.Zero, Total: decimal.Zero}
	for _, line := range lines {
		totals.Untaxed = totals.Untaxed.Add(decimalOf(line, models.PurchaseOrderLineFieldSubtotal))
		totals.Tax = totals.Tax.Add(decimalOf(line, models.PurchaseOrderLineFieldTaxAmount))
	}
	totals.Total = totals.Untaxed.Add(totals.Tax)
	return totals
}

// RecomputeOrderTotals rewrites whatever is stale on an order and its lines. The caller must supply
// the triggering write's transaction through ctx, or readers can land in a window where the header
// and lines disagree. Values already correct are left alone: every write bumps the etag, which
// would invalidate a client's in-flight edit.
func RecomputeOrderTotals(ctx corectx.Context, orderId string) error {
	orderEngine, err := engineFor(models.PurchaseOrderSchemaName)
	if err != nil {
		return err
	}
	lineEngine, err := engineFor(models.PurchaseOrderLineSchemaName)
	if err != nil {
		return err
	}

	found, err := orderEngine.ResourceRepository().FindByKeys(ctx, dmodel.DynamicFields{
		models.PurchaseOrderFieldId: orderId,
	})
	if err != nil {
		return errors.Wrap(err, "RecomputeOrderTotals")
	}
	if found == nil || !found.HasData {
		// The order is gone, most likely deleted by the same request: nothing to recompute.
		return nil
	}

	lines, err := models.FindOrderLines(
		ctx, lineEngine.ResourceRepository(), orderId, models.MaxOrderLines)
	if err != nil {
		return err
	}

	lines, err = rewriteStaleLines(ctx, lineEngine, lines, scaleForOrder(ctx, found.Data))
	if err != nil {
		return err
	}
	return rewriteStaleHeader(ctx, orderEngine, found.Data, lines)
}

// rewriteStaleLines stores each line's computed totals where they differ, and returns the lines
// with the computed values in place so the header sums the new numbers, not the replaced ones.
func rewriteStaleLines(
	ctx corectx.Context, lineEngine drif.DynamicResourceEngine, lines []dmodel.DynamicFields,
	scale int32,
) ([]dmodel.DynamicFields, error) {
	for _, line := range lines {
		computed := ComputeLineTotals(line, scale)

		stale := !decimalOf(line, models.PurchaseOrderLineFieldSubtotal).Equal(computed.Subtotal) ||
			!decimalOf(line, models.PurchaseOrderLineFieldTaxAmount).Equal(computed.Tax) ||
			!decimalOf(line, models.PurchaseOrderLineFieldTotal).Equal(computed.Total)

		// Updated in memory even when not stale, so the header below sums correct values.
		line[models.PurchaseOrderLineFieldSubtotal] = computed.Subtotal
		line[models.PurchaseOrderLineFieldTaxAmount] = computed.Tax
		line[models.PurchaseOrderLineFieldTotal] = computed.Total

		if !stale {
			continue
		}
		_, err := lineEngine.ResourceRepository().Update(ctx, dmodel.DynamicFields{
			models.PurchaseOrderLineFieldId:        stringOf(line, models.PurchaseOrderLineFieldId),
			models.PurchaseOrderLineFieldSubtotal:  computed.Subtotal,
			models.PurchaseOrderLineFieldTaxAmount: computed.Tax,
			models.PurchaseOrderLineFieldTotal:     computed.Total,
			basemodel.FieldEtag:                    stringOf(line, basemodel.FieldEtag),
		})
		if err != nil {
			return nil, errors.Wrap(err, "rewriteStaleLines")
		}
	}
	return lines, nil
}

// rewriteStaleHeader stores the order's three totals when they differ from the sum of its lines.
func rewriteStaleHeader(
	ctx corectx.Context, orderEngine drif.DynamicResourceEngine,
	order dmodel.DynamicFields, lines []dmodel.DynamicFields,
) error {
	totals := ComputeOrderTotals(lines)

	if decimalOf(order, models.PurchaseOrderFieldUntaxedAmount).Equal(totals.Untaxed) &&
		decimalOf(order, models.PurchaseOrderFieldTaxAmount).Equal(totals.Tax) &&
		decimalOf(order, models.PurchaseOrderFieldTotalAmount).Equal(totals.Total) {
		return nil
	}

	_, err := orderEngine.ResourceRepository().Update(ctx, dmodel.DynamicFields{
		models.PurchaseOrderFieldId:            stringOf(order, models.PurchaseOrderFieldId),
		models.PurchaseOrderFieldUntaxedAmount: totals.Untaxed,
		models.PurchaseOrderFieldTaxAmount:     totals.Tax,
		models.PurchaseOrderFieldTotalAmount:   totals.Total,
		basemodel.FieldEtag:                    stringOf(order, basemodel.FieldEtag),
	})
	return errors.Wrap(err, "rewriteStaleHeader")
}

// StampLineTotals fills a line's computed fields into create or update params. It must run before
// the write, not as a recompute after it: subtotal and total are required_for_create, so the schema
// would reject the create first. Client-supplied subtotal and total are overwritten, not rejected.
func StampLineTotals(params dmodel.DynamicFields) {
	computed := ComputeLineTotals(params, defaultScale)
	params[models.PurchaseOrderLineFieldSubtotal] = computed.Subtotal
	params[models.PurchaseOrderLineFieldTaxAmount] = computed.Tax
	params[models.PurchaseOrderLineFieldTotal] = computed.Total
}

// StampOrderTotalsForCreate zeroes a new order's totals: an order is created without lines, and the
// fields are required_for_create so they cannot simply be omitted. Any client-sent values are
// discarded; the totals become real when the first line triggers a recompute.
func StampOrderTotalsForCreate(params dmodel.DynamicFields) {
	params[models.PurchaseOrderFieldUntaxedAmount] = decimal.Zero
	params[models.PurchaseOrderFieldTaxAmount] = decimal.Zero
	params[models.PurchaseOrderFieldTotalAmount] = decimal.Zero
}
