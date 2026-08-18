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

// The money. PUR-R4 and D8: every line's subtotal, tax_amount and total are stored, and so are the
// order's untaxed_amount, tax_amount and total_amount — none of them is computed at read time.
//
// Storing them is what makes an order readable years later. A total derived on read is derived
// from today's prices and today's rounding rules, so an order would quietly change its own value
// whenever either changed; a stored one is what the business agreed to pay.
//
// The cost of storing is that the stored value can disagree with the lines. Where it does, the
// LINES WIN — they are what a reader can verify by adding up, and a header that disagrees with
// them is the thing that is wrong. This is the invoice precedent from PAY-009.

// defaultScale is the number of decimal places money is rounded to when the order's currency cannot
// be resolved — because it has none yet, or because the lookup failed.
//
// The real scale comes from the currency (OrderReferenceValidator.ScaleFor, [PUR-018]): 0 for VND,
// 2 for USD, 3 for KWD. Two is the right FALLBACK because it is what the overwhelming majority of
// currencies use, and because an order with no currency yet is an ordinary draft rather than an
// error — refusing to total it would make the money unreadable over a problem that is not about
// money.
const defaultScale int32 = 2

// orderScaleResolver resolves an order's rounding scale from its currency.
//
// It is a package-level hook rather than a parameter threaded through every call because the
// recompute is reached from four places and none of them holds a currency port. Init sets it;
// unset, every order rounds to defaultScale, which is exactly the behaviour before [PUR-018].
var orderScaleResolver func(ctx corectx.Context, currencyId string) int32

// SetOrderScaleResolver installs the currency-aware scale lookup. Called once during Init.
func SetOrderScaleResolver(resolve func(ctx corectx.Context, currencyId string) int32) {
	orderScaleResolver = resolve
}

// scaleForOrder returns the scale an order's amounts round to.
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

// ComputeLineTotals derives one line's three money fields from its inputs.
//
// subtotal = quantity x unit_price, less discount_percent. total = subtotal + tax_amount.
//
// tax_amount is an INPUT here, not a calculation (D9). There is no tax engine in this module and
// §55.15 forbids building one, so the client supplies a line's tax and the server sums it. That is
// why tax_amount is writable on the line while subtotal and total are not: one is a number only
// the caller knows, and the other two are arithmetic the caller must not be able to contradict.
//
// A non-product line — a section, a subsection, a note — contributes nothing. It exists to organise
// the printed order, and giving it a quantity and a price would let a heading carry money.
func ComputeLineTotals(line dmodel.DynamicFields, scale int32) LineTotals {
	if !isMoneyBearingLine(line) {
		return LineTotals{Subtotal: decimal.Zero, Tax: decimal.Zero, Total: decimal.Zero}
	}

	quantity := decimalOf(line, models.PurchaseOrderLineFieldQuantity)
	unitPrice := decimalOf(line, models.PurchaseOrderLineFieldUnitPrice)
	discount := decimalOf(line, models.PurchaseOrderLineFieldDiscountPercent)
	tax := decimalOf(line, models.PurchaseOrderLineFieldTaxAmount)

	gross := quantity.Mul(unitPrice)
	// The discount is applied to the whole line rather than to the unit price, so that a 3.333%
	// discount is rounded once instead of once per unit.
	kept := decimal.NewFromInt(100).Sub(discount).Div(decimal.NewFromInt(100))
	subtotal := gross.Mul(kept).Round(scale)
	tax = tax.Round(scale)

	return LineTotals{
		Subtotal: subtotal,
		Tax:      tax,
		Total:    subtotal.Add(tax),
	}
}

// isMoneyBearingLine reports whether a line contributes to the order's totals.
//
// The default is TRUE, deliberately: a line whose type is missing or unrecognised is treated as a
// product line and priced. The alternative — silently contributing zero — would drop money out of
// an order's total with nothing to show for it, which is the worse failure of the two.
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

// ComputeOrderTotals sums the lines into the order's three totals.
//
// It sums the lines' STORED subtotal and tax rather than recomputing each line from its inputs,
// because RecomputeOrderTotals has already made the stored values correct — and because the header
// must equal the sum of what the lines display. A header derived independently could differ from
// the visible column by a rounding step, which is the one discrepancy a reader is guaranteed to
// notice.
func ComputeOrderTotals(lines []dmodel.DynamicFields) OrderTotals {
	totals := OrderTotals{Untaxed: decimal.Zero, Tax: decimal.Zero, Total: decimal.Zero}
	for _, line := range lines {
		totals.Untaxed = totals.Untaxed.Add(decimalOf(line, models.PurchaseOrderLineFieldSubtotal))
		totals.Tax = totals.Tax.Add(decimalOf(line, models.PurchaseOrderLineFieldTaxAmount))
	}
	totals.Total = totals.Untaxed.Add(totals.Tax)
	return totals
}

// RecomputeOrderTotals rewrites whatever is stale on an order and its lines.
//
// The caller supplies the transaction through ctx: a recompute that committed separately from the
// write that triggered it would leave a window in which the header and the lines disagree, and a
// reader landing in that window sees an order whose total is wrong.
//
// Nothing already correct is rewritten. That is not only an optimisation: every write bumps the
// etag, so rewriting an unchanged line would invalidate a client's in-flight edit for no reason.
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
		// The order is gone — deleted by the same request that asked for this, most likely. There
		// is nothing to recompute and nothing wrong.
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

// rewriteStaleLines stores each line's computed totals where they differ from what is there, and
// returns the lines with the computed values in place so the header sums the new numbers rather
// than the ones it just replaced.
func rewriteStaleLines(
	ctx corectx.Context, lineEngine drif.DynamicResourceEngine, lines []dmodel.DynamicFields,
	scale int32,
) ([]dmodel.DynamicFields, error) {
	for _, line := range lines {
		computed := ComputeLineTotals(line, scale)

		stale := !decimalOf(line, models.PurchaseOrderLineFieldSubtotal).Equal(computed.Subtotal) ||
			!decimalOf(line, models.PurchaseOrderLineFieldTaxAmount).Equal(computed.Tax) ||
			!decimalOf(line, models.PurchaseOrderLineFieldTotal).Equal(computed.Total)

		// The in-memory line is updated either way, so that a line already correct still sums
		// correctly into the header below.
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

// StampLineTotals fills a line's three computed fields into the params of a create or update.
//
// It runs before the write rather than recomputing after it, because subtotal and total are
// required_for_create: a create that omitted them would be refused by the schema before any
// recompute could fix it. A client-supplied subtotal or total is OVERWRITTEN rather than rejected —
// they are not part of the request's meaning, and a client echoing a record back should not fail
// for carrying them.
func StampLineTotals(params dmodel.DynamicFields) {
	computed := ComputeLineTotals(params, defaultScale)
	params[models.PurchaseOrderLineFieldSubtotal] = computed.Subtotal
	params[models.PurchaseOrderLineFieldTaxAmount] = computed.Tax
	params[models.PurchaseOrderLineFieldTotal] = computed.Total
}

// StampOrderTotalsForCreate zeroes a new order's three totals.
//
// An order is created without lines, so its totals are zero by construction; they become real when
// the first line lands and the recompute runs. They are stamped rather than left out because they
// are required_for_create, and stamped to zero rather than to whatever the client sent because a
// header total is a summary of lines that do not exist yet.
func StampOrderTotalsForCreate(params dmodel.DynamicFields) {
	params[models.PurchaseOrderFieldUntaxedAmount] = decimal.Zero
	params[models.PurchaseOrderFieldTaxAmount] = decimal.Zero
	params[models.PurchaseOrderFieldTotalAmount] = decimal.Zero
}
