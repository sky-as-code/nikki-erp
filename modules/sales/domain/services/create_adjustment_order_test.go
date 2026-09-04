package services

import (
	"testing"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The adjustment order's line arithmetic, which is where a mistake would misstate what a customer
// kept — and therefore what any later document says they owe.

func orderLine(ordered, returned, gross, discount, net, tax, final string) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.SalesOrderLineFieldOrderedQuantity:  dec(ordered),
		models.SalesOrderLineFieldReturnedQuantity: dec(returned),
		models.SalesOrderLineFieldGrossAmount:      dec(gross),
		models.SalesOrderLineFieldDiscountAmount:   dec(discount),
		models.SalesOrderLineFieldNetAmount:        dec(net),
		models.SalesOrderLineFieldTaxAmount:        dec(tax),
		models.SalesOrderLineFieldFinalAmount:      dec(final),
	}
}

// Half a line comes back, so half its money stays. Prorated rather than recomputed from unit price:
// a line's amounts carry promotions and rounding that unit price alone cannot reproduce.
func TestKeptAmountsAreProratedFromTheOriginal(t *testing.T) {
	line := keptLine{
		original: orderLine("4", "2", "400", "40", "360", "36", "396"),
		quantity: dec("2"),
	}

	amounts := proratedLineAmounts(line)

	for _, testCase := range []struct {
		name string
		got  decimal.Decimal
		want string
	}{
		{"gross", amounts.gross, "200"},
		{"discount", amounts.discount, "20"},
		{"net", amounts.net, "180"},
		{"tax", amounts.tax, "18"},
		{"final", amounts.final, "198"},
	} {
		if !testCase.got.Equal(dec(testCase.want)) {
			t.Errorf("%s = %s, want %s", testCase.name, testCase.got, testCase.want)
		}
	}
}

// A line nothing came back from keeps all of its money, exactly. A rounding slip here would change
// the price of goods the customer never touched.
func TestAnUntouchedLineKeepsItsAmountsExactly(t *testing.T) {
	line := keptLine{
		original: orderLine("3", "0", "299.97", "0", "299.97", "29.99", "329.96"),
		quantity: dec("3"),
	}

	amounts := proratedLineAmounts(line)
	if !amounts.final.Equal(dec("329.96")) {
		t.Errorf("final = %s, want 329.96 — an untouched line must not be re-derived", amounts.final)
	}
	if !amounts.gross.Equal(dec("299.97")) {
		t.Errorf("gross = %s, want 299.97", amounts.gross)
	}
}

// A line with no ordered quantity cannot be divided by. It contributes nothing rather than panicking
// or producing an infinity that would poison the order's totals.
func TestAZeroQuantityLineContributesNothing(t *testing.T) {
	line := keptLine{
		original: orderLine("0", "0", "100", "0", "100", "10", "110"),
		quantity: dec("1"),
	}

	amounts := proratedLineAmounts(line)
	if !amounts.final.IsZero() {
		t.Errorf("final = %s, want 0 for a line with no ordered quantity", amounts.final)
	}
}

// The order's totals are the sum of what was kept, line by line.
func TestAdjustmentTotalsSumTheKeptLines(t *testing.T) {
	kept := []keptLine{
		{original: orderLine("2", "1", "200", "0", "200", "20", "220"), quantity: dec("1")},
		{original: orderLine("4", "0", "400", "40", "360", "36", "396"), quantity: dec("4")},
	}

	totals := adjustmentTotalsOf(kept)

	// Half the first line (100 gross) plus all of the second (400).
	if !totals.subtotal.Equal(dec("500")) {
		t.Errorf("subtotal = %s, want 500", totals.subtotal)
	}
	if !totals.discount.Equal(dec("40")) {
		t.Errorf("discount = %s, want 40", totals.discount)
	}
	if !totals.tax.Equal(dec("46")) {
		t.Errorf("tax = %s, want 46", totals.tax)
	}
	if !totals.grand.Equal(dec("506")) {
		t.Errorf("grand total = %s, want 506", totals.grand)
	}
}

// Nothing kept means no adjustment order. A full return leaves an empty document about nothing, and
// an order with no lines would confuse every report that counts them.
func TestNothingKeptProducesNoTotals(t *testing.T) {
	totals := adjustmentTotalsOf(nil)
	if !totals.grand.IsZero() || !totals.subtotal.IsZero() {
		t.Error("an empty set of kept lines must total zero")
	}
}

// The supersession links are read from presence alone, so an order is never half-marked.
func TestSupersessionIsReadFromTheLink(t *testing.T) {
	plain := models.NewSalesOrderFrom(dmodel.DynamicFields{})
	if plain.IsSuperseded() || plain.IsAdjustment() {
		t.Error("an ordinary order is neither superseded nor an adjustment")
	}

	superseded := models.NewSalesOrderFrom(dmodel.DynamicFields{
		models.SalesOrderFieldAdjustedByOrderId: "01ADJ0000000000000000000",
	})
	if !superseded.IsSuperseded() {
		t.Error("an order naming its adjustment must read as superseded")
	}
	if superseded.IsAdjustment() {
		t.Error("being superseded does not make an order an adjustment")
	}

	adjustment := models.NewSalesOrderFrom(dmodel.DynamicFields{
		models.SalesOrderFieldAdjustsOrderId: "01ORD0000000000000000000",
	})
	if !adjustment.IsAdjustment() {
		t.Error("an order naming what it restates must read as an adjustment")
	}
	if adjustment.IsSuperseded() {
		t.Error("an adjustment order is not itself superseded")
	}
}
