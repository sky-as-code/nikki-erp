package invoicing

import (
	"testing"

	"github.com/shopspring/decimal"

	itInvoicing "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external/invoicing"
)

// The tax rate changes units at this boundary, and getting it wrong is a silent hundredfold error on
// a legal document. That is what these tests exist for.

func dec(value string) decimal.Decimal {
	return decimal.RequireFromString(value)
}

// The conversion the port's own comment calls "the likeliest place for a silent 100x error": Sales
// carries 0.1 for 10%, the invoice line carries 10.
func TestTaxRateConvertsFromFractionToPercent(t *testing.T) {
	cases := []struct {
		fraction string
		percent  string
	}{
		{"0.1", "10"},
		{"0.05", "5"},
		{"0.08", "8"},
		{"0", "0"},

		// A rate with more precision than two places must survive: it is a rate, not a rounded
		// display value.
		{"0.075", "7.5"},
	}

	for _, testCase := range cases {
		t.Run(testCase.fraction, func(t *testing.T) {
			lines := issueLinesOf(itInvoicing.IssueRequest{
				Lines: []itInvoicing.IssueLine{{TaxRateSnapshot: dec(testCase.fraction)}},
			})

			if len(lines) != 1 {
				t.Fatalf("got %d lines, want 1", len(lines))
			}
			if !lines[0].TaxRatePercent.Equal(dec(testCase.percent)) {
				t.Errorf("tax rate = %s, want %s — a fraction reached the invoice unconverted",
					lines[0].TaxRatePercent, testCase.percent)
			}
		})
	}
}

// The rest of a line crosses unchanged. UnitAmount is the historical price the customer was charged;
// re-deriving it anywhere would restate a completed sale.
func TestLineAmountsCrossUnchanged(t *testing.T) {
	lines := issueLinesOf(itInvoicing.IssueRequest{
		Lines: []itInvoicing.IssueLine{{
			Description:     "Coffee beans 1kg",
			Quantity:        dec("1.5"),
			UnitAmount:      dec("249000.50"),
			TaxRateSnapshot: dec("0.1"),
		}},
	})

	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1", len(lines))
	}
	line := lines[0]

	if line.Description != "Coffee beans 1kg" {
		t.Errorf("description = %q", line.Description)
	}
	// A fractional quantity is exactly why the invoice line's column was widened: goods sold by
	// weight cannot be stated as a whole number of units.
	if !line.Quantity.Equal(dec("1.5")) {
		t.Errorf("quantity = %s, want 1.5", line.Quantity)
	}
	if !line.UnitPrice.Equal(dec("249000.50")) {
		t.Errorf("unit price = %s, want 249000.50 — a historical price must not be re-derived",
			line.UnitPrice)
	}
}

// Every line converts, not just the first: an invoice mixing rates is the case a single shared
// conversion would get wrong.
func TestEveryLineConverts(t *testing.T) {
	lines := issueLinesOf(itInvoicing.IssueRequest{
		Lines: []itInvoicing.IssueLine{
			{TaxRateSnapshot: dec("0.1")},
			{TaxRateSnapshot: dec("0.05")},
			{TaxRateSnapshot: dec("0")},
		},
	})

	want := []string{"10", "5", "0"}
	if len(lines) != len(want) {
		t.Fatalf("got %d lines, want %d", len(lines), len(want))
	}
	for index, expected := range want {
		if !lines[index].TaxRatePercent.Equal(dec(expected)) {
			t.Errorf("line %d tax rate = %s, want %s",
				index, lines[index].TaxRatePercent, expected)
		}
	}
}

// An empty request produces no lines rather than a nil slice with a length quirk, so the far side's
// "an invoice must have at least one line" refusal is what reports it.
func TestNoLinesProducesNoLines(t *testing.T) {
	if lines := issueLinesOf(itInvoicing.IssueRequest{}); len(lines) != 0 {
		t.Errorf("got %d lines from an empty request", len(lines))
	}
}
