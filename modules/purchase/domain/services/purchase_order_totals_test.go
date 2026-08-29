package services

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"

	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

func dec(value string) decimal.Decimal {
	parsed, err := decimal.NewFromString(value)
	if err != nil {
		panic(err)
	}
	return parsed
}

func productLine(quantity, unitPrice, discount, tax string) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.PurchaseOrderLineFieldLineType:        string(models.PurchaseOrderLineTypeProduct),
		models.PurchaseOrderLineFieldQuantity:        dec(quantity),
		models.PurchaseOrderLineFieldUnitPrice:       dec(unitPrice),
		models.PurchaseOrderLineFieldDiscountPercent: dec(discount),
		models.PurchaseOrderLineFieldTaxAmount:       dec(tax),
	}
}

// The arithmetic of one line.
func TestComputeLineTotals(t *testing.T) {
	testCases := []struct {
		name                       string
		quantity                   string
		unitPrice                  string
		discount                   string
		tax                        string
		subtotal, taxOut, totalOut string
	}{
		{
			name: "quantity times price", quantity: "3", unitPrice: "10", discount: "0", tax: "0",
			subtotal: "30", taxOut: "0", totalOut: "30",
		},
		{
			name: "tax is added, not derived", quantity: "2", unitPrice: "50", discount: "0", tax: "10",
			subtotal: "100", taxOut: "10", totalOut: "110",
		},
		{
			name: "a percentage discount comes off the line", quantity: "4", unitPrice: "25", discount: "10", tax: "0",
			subtotal: "90", taxOut: "0", totalOut: "90",
		},
		{
			// The discount is applied to the whole line, so the rounding happens once. Per-unit it
			// would be 33.3333 -> 33.33 x 3 = 99.99, a cent adrift.
			name: "the discount is rounded once for the line", quantity: "3", unitPrice: "33.3333", discount: "0", tax: "0",
			subtotal: "100.00", taxOut: "0", totalOut: "100.00",
		},
		{
			name: "a fractional quantity prices normally", quantity: "2.5", unitPrice: "4", discount: "0", tax: "0",
			subtotal: "10", taxOut: "0", totalOut: "10",
		},
		{
			name: "a full discount leaves nothing", quantity: "5", unitPrice: "20", discount: "100", tax: "0",
			subtotal: "0", taxOut: "0", totalOut: "0",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			line := productLine(testCase.quantity, testCase.unitPrice, testCase.discount, testCase.tax)

			totals := ComputeLineTotals(line, defaultScale)

			assert.True(t, totals.Subtotal.Equal(dec(testCase.subtotal)),
				"subtotal: want %s, got %s", testCase.subtotal, totals.Subtotal)
			assert.True(t, totals.Tax.Equal(dec(testCase.taxOut)),
				"tax: want %s, got %s", testCase.taxOut, totals.Tax)
			assert.True(t, totals.Total.Equal(dec(testCase.totalOut)),
				"total: want %s, got %s", testCase.totalOut, totals.Total)
		})
	}
}

// A section, subsection or note organizes the printed order; letting a heading carry money would
// put a number in the total that no product accounts for.
func TestNonProductLinesContributeNothing(t *testing.T) {
	for _, lineType := range []models.PurchaseOrderLineType{
		models.PurchaseOrderLineTypeSection,
		models.PurchaseOrderLineTypeSubsection,
		models.PurchaseOrderLineTypeNote,
	} {
		t.Run(string(lineType), func(t *testing.T) {
			line := productLine("10", "100", "0", "50")
			line[models.PurchaseOrderLineFieldLineType] = string(lineType)

			totals := ComputeLineTotals(line, defaultScale)

			assert.True(t, totals.Subtotal.IsZero())
			assert.True(t, totals.Tax.IsZero())
			assert.True(t, totals.Total.IsZero())
		})
	}
}

// A line whose type is missing or unrecognised is priced, not silently zeroed: dropping money out
// of a total with nothing to show for it is the worse failure.
func TestAnUnknownLineTypeIsStillPriced(t *testing.T) {
	for _, lineType := range []any{nil, "", "something_new"} {
		line := productLine("2", "10", "0", "0")
		line[models.PurchaseOrderLineFieldLineType] = lineType

		totals := ComputeLineTotals(line, defaultScale)

		assert.True(t, totals.Subtotal.Equal(dec("20")), "line type %v must still be priced", lineType)
	}
}

// The header equals the sum of the lines' stored values, so the total always matches the column a
// reader can add up by hand.
func TestComputeOrderTotals(t *testing.T) {
	lines := []dmodel.DynamicFields{
		{
			models.PurchaseOrderLineFieldSubtotal:  dec("100.50"),
			models.PurchaseOrderLineFieldTaxAmount: dec("10.05"),
		},
		{
			models.PurchaseOrderLineFieldSubtotal:  dec("249.50"),
			models.PurchaseOrderLineFieldTaxAmount: dec("24.95"),
		},
	}

	totals := ComputeOrderTotals(lines)

	assert.True(t, totals.Untaxed.Equal(dec("350.00")), "untaxed: %s", totals.Untaxed)
	assert.True(t, totals.Tax.Equal(dec("35.00")), "tax: %s", totals.Tax)
	assert.True(t, totals.Total.Equal(dec("385.00")), "total: %s", totals.Total)
}

// An order with no lines totals zero rather than erroring: that is the state every order starts in.
func TestAnEmptyOrderTotalsZero(t *testing.T) {
	totals := ComputeOrderTotals(nil)

	assert.True(t, totals.Untaxed.IsZero())
	assert.True(t, totals.Tax.IsZero())
	assert.True(t, totals.Total.IsZero())
}

// A client-supplied subtotal or total is overwritten, never trusted. They are no_update in the
// schema for the same reason; this checks the service agrees.
func TestStampLineTotalsOverwritesClientValues(t *testing.T) {
	line := productLine("2", "10", "0", "3")
	line[models.PurchaseOrderLineFieldSubtotal] = dec("999999")
	line[models.PurchaseOrderLineFieldTotal] = dec("999999")

	StampLineTotals(line)

	assert.True(t, decimalOf(line, models.PurchaseOrderLineFieldSubtotal).Equal(dec("20")))
	assert.True(t, decimalOf(line, models.PurchaseOrderLineFieldTotal).Equal(dec("23")))
	// The tax the client sent is kept: it is an input, not a calculation.
	assert.True(t, decimalOf(line, models.PurchaseOrderLineFieldTaxAmount).Equal(dec("3")))
}

// decimalOf is the reader every total is summed through, so it must not panic on whatever shape the
// repository hands back: a panic takes down the request instead of reporting a bad row.
func TestDecimalOfReadsEveryShapeWithoutPanicking(t *testing.T) {
	value := dec("12.34")
	testCases := []struct {
		name  string
		value any
		want  string
	}{
		{"decimal", value, "12.34"},
		{"pointer to decimal", &value, "12.34"},
		{"string", "12.34", "12.34"},
		{"float", 12.34, "12.34"},
		{"int64", int64(12), "12"},
		{"absent", nil, "0"},
		{"nil pointer", (*decimal.Decimal)(nil), "0"},
		{"unparseable string", "not a number", "0"},
		{"wrong type entirely", []string{"nope"}, "0"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fields := dmodel.DynamicFields{"amount": testCase.value}

			assert.True(t, decimalOf(fields, "amount").Equal(dec(testCase.want)))
		})
	}

	assert.True(t, decimalOf(dmodel.DynamicFields{}, "missing").IsZero())
}
