package pricing

import (
	"testing"

	"github.com/shopspring/decimal"
)

// Two invariants recur here and matter more than any individual number: Σ line net == order net at
// every step, and Σ component allocated == the combo line's amount, exactly.

func dec(value string) decimal.Decimal {
	return decimal.RequireFromString(value)
}

func lineOf(key string, number int32, quantity, price string) LineInput {
	return LineInput{
		Key:                 key,
		LineNumber:          number,
		ProductVariantId:    "variant-" + key,
		UomId:               "uom-each",
		Quantity:            dec(quantity),
		CatalogueUnitPrice:  dec(price),
		ProductCode:         "CODE-" + key,
		ProductName:         "Product " + key,
		RequiresFulfillment: true,
	}
}

func vndContext() Context {
	// VND has no minor unit. No tax entries means every line is taxed at zero.
	return Context{CurrencyScale: 0}
}

// taxedContext taxes each named line at the given amount and fraction rate, standing in for
// ResolveBasketTax.
func taxedContext(rate string, amountByKey map[string]string) Context {
	byLineKey := make(map[string]LineTax, len(amountByKey))
	for key, amount := range amountByKey {
		byLineKey[key] = LineTax{RateSnapshot: dec(rate), Amount: dec(amount)}
	}
	return Context{CurrencyScale: 0, TaxByLineKey: byLineKey}
}

func lineByKey(t *testing.T, result Result, key string) LineResult {
	t.Helper()
	for _, line := range result.Lines {
		if line.Key == key {
			return line
		}
	}
	t.Fatalf("no line %q in the result", key)
	return LineResult{}
}

// assertLineNetSumsToOrderNet checks the core allocation invariant.
func assertLineNetSumsToOrderNet(t *testing.T, result Result) {
	t.Helper()
	total := decimal.Zero
	for _, line := range result.Lines {
		total = total.Add(line.NetAmount)
	}
	// The grand total may differ by the rounding adjustment, which applies to the order rather than
	// the lines, so compare against the pre-rounding sum.
	expected := result.Subtotal.Sub(result.DiscountTotal)
	if !total.Equal(expected) {
		t.Errorf("Σ line net = %s, but subtotal - discount = %s", total, expected)
	}
}

func TestEmptyBasket(t *testing.T) {
	result := Calculate(Input{Context: vndContext()})

	if len(result.Lines) != 0 {
		t.Errorf("an empty basket must produce no lines, got %d", len(result.Lines))
	}
	if len(result.Adjustments) != 0 {
		t.Errorf("an empty basket must produce no adjustments, got %d", len(result.Adjustments))
	}
	for name, amount := range map[string]decimal.Decimal{
		"subtotal": result.Subtotal, "discount": result.DiscountTotal,
		"tax": result.TaxTotal, "grand": result.GrandTotal,
	} {
		if !amount.IsZero() {
			t.Errorf("an empty basket must total zero, %s = %s", name, amount)
		}
	}
}

// Step 2 with nothing to override it: the catalogue price applies.
func TestCataloguePriceWhenNoPricelistMatches(t *testing.T) {
	result := Calculate(Input{
		Lines:   []LineInput{lineOf("a", 1, "3", "10000")},
		Context: vndContext(),
	})

	line := lineByKey(t, result, "a")
	if line.PricingSource != "catalogue" {
		t.Errorf("pricing_source = %q, want catalogue", line.PricingSource)
	}
	if !line.GrossAmount.Equal(dec("30000")) {
		t.Errorf("gross = %s, want 30000", line.GrossAmount)
	}
	if !result.GrandTotal.Equal(dec("30000")) {
		t.Errorf("grand total = %s, want 30000", result.GrandTotal)
	}
	assertLineNetSumsToOrderNet(t, result)
}

// A pricelist item overrides the catalogue price.
func TestPricelistOverridesCatalogue(t *testing.T) {
	result := Calculate(Input{
		Lines: []LineInput{lineOf("a", 1, "2", "10000")},
		PricelistItems: []PricelistItem{{
			ProductVariantId: "variant-a", UomId: "uom-each",
			UnitPrice: dec("8000"), MinQuantity: decimal.Zero, Specificity: 1,
		}},
		Context: vndContext(),
	})

	line := lineByKey(t, result, "a")
	if line.PricingSource != "pricelist" {
		t.Errorf("pricing_source = %q, want pricelist", line.PricingSource)
	}
	if !line.GrossAmount.Equal(dec("16000")) {
		t.Errorf("gross = %s, want 16000", line.GrossAmount)
	}
}

// Specificity beats priority, and a higher quantity break beats a lower one at equal specificity.
func TestPricelistSelection(t *testing.T) {
	items := []PricelistItem{
		{ProductVariantId: "variant-a", UomId: "uom-each", UnitPrice: dec("10000"),
			MinQuantity: decimal.Zero, Specificity: 0, Priority: 99},
		{ProductVariantId: "variant-a", UomId: "uom-each", UnitPrice: dec("9000"),
			MinQuantity: decimal.Zero, Specificity: 2, Priority: 0},
		{ProductVariantId: "variant-a", UomId: "uom-each", UnitPrice: dec("7000"),
			MinQuantity: dec("10"), Specificity: 2, Priority: 0},
	}

	// A line of 2 cannot reach the ten-break, so the specific list's base price applies despite the
	// global list's much higher priority.
	small := Calculate(Input{
		Lines: []LineInput{lineOf("a", 1, "2", "50000")}, PricelistItems: items,
		Context: vndContext(),
	})
	if got := lineByKey(t, small, "a").EffectiveUnitPrice; !got.Equal(dec("9000")) {
		t.Errorf("unit price = %s, want 9000: specificity beats priority", got)
	}

	// A line of 12 reaches the break, and the break wins at equal specificity.
	large := Calculate(Input{
		Lines: []LineInput{lineOf("a", 1, "12", "50000")}, PricelistItems: items,
		Context: vndContext(),
	})
	if got := lineByKey(t, large, "a").EffectiveUnitPrice; !got.Equal(dec("7000")) {
		t.Errorf("unit price = %s, want 7000: the reached quantity break wins", got)
	}
}

// Worked combo example: references 30/20/10, combo price 48,000, shares 24,000/16,000/8,000.
func TestComboAllocationWorkedExample(t *testing.T) {
	comboLine := lineOf("combo", 1, "1", "0")
	comboLine.ComboId = "combo-1"

	result := Calculate(Input{
		Lines: []LineInput{comboLine},
		Combos: []ComboDefinition{{
			ComboId:    "combo-1",
			ComboPrice: dec("48000"),
			Components: []ComboComponentInput{
				{Key: "c1", Sequence: 1, ProductVariantId: "v1", UomId: "uom-each",
					Quantity: dec("1"), ReferencePrice: dec("30")},
				{Key: "c2", Sequence: 2, ProductVariantId: "v2", UomId: "uom-each",
					Quantity: dec("1"), ReferencePrice: dec("20")},
				{Key: "c3", Sequence: 3, ProductVariantId: "v3", UomId: "uom-each",
					Quantity: dec("1"), ReferencePrice: dec("10")},
			},
		}},
		Context: vndContext(),
	})

	line := lineByKey(t, result, "combo")
	if line.LineType != "combo" {
		t.Errorf("line_type = %q, want combo", line.LineType)
	}
	if !line.GrossAmount.Equal(dec("48000")) {
		t.Errorf("combo gross = %s, want 48000 (the bundle price, not the parts)", line.GrossAmount)
	}

	want := map[string]string{"c1": "24000", "c2": "16000", "c3": "8000"}
	for _, component := range line.Components {
		if !component.AllocatedNetAmount.Equal(dec(want[component.Key])) {
			t.Errorf("%s allocated %s, want %s",
				component.Key, component.AllocatedNetAmount, want[component.Key])
		}
	}

	// The exact-sum invariant the residual rule exists for.
	total := decimal.Zero
	for _, component := range line.Components {
		total = total.Add(component.AllocatedNetAmount)
	}
	if !total.Equal(dec("48000")) {
		t.Errorf("Σ component allocated = %s, want exactly 48000", total)
	}
}

// A combo whose price does not divide evenly across its components still sums exactly.
func TestComboAllocationWithResidual(t *testing.T) {
	comboLine := lineOf("combo", 1, "1", "0")
	comboLine.ComboId = "combo-r"

	result := Calculate(Input{
		Lines: []LineInput{comboLine},
		Combos: []ComboDefinition{{
			ComboId:    "combo-r",
			ComboPrice: dec("100"),
			Components: []ComboComponentInput{
				{Key: "c1", Sequence: 1, ProductVariantId: "v1", UomId: "u",
					Quantity: dec("1"), ReferencePrice: dec("1")},
				{Key: "c2", Sequence: 2, ProductVariantId: "v2", UomId: "u",
					Quantity: dec("1"), ReferencePrice: dec("1")},
				{Key: "c3", Sequence: 3, ProductVariantId: "v3", UomId: "u",
					Quantity: dec("1"), ReferencePrice: dec("1")},
			},
		}},
		Context: vndContext(),
	})

	line := lineByKey(t, result, "combo")
	total := decimal.Zero
	for _, component := range line.Components {
		total = total.Add(component.AllocatedNetAmount)
	}
	if !total.Equal(dec("100")) {
		t.Errorf("Σ component allocated = %s, want exactly 100 — the residual was lost", total)
	}
}

// A percentage discount, allocated across lines on the post-discount-net basis.
// A fixed discount larger than the basket is capped: a negative total would pay the customer to shop.
// A free item becomes a REAL line at zero price, not an adjustment.
// A giveaway line is never itself discounted: its share would come off a line being paid for.
// Rewards do not commute, so the sequence decides the answer.
// Step 8: rounding is applied once, to the grand total, and recorded as its own adjustment so the
// gap between the line sum and the total is explained.
func TestRoundingIsRecordedAsAnAdjustment(t *testing.T) {
	result := Calculate(Input{
		// A third of 100 at VND scale cannot come out whole.
		Lines:   []LineInput{lineOf("a", 1, "3", "33.3333")},
		Context: vndContext(),
	})

	var rounding *Adjustment
	for index := range result.Adjustments {
		if result.Adjustments[index].Type == AdjustmentRounding {
			rounding = &result.Adjustments[index]
		}
	}
	if rounding == nil {
		t.Fatal("a total needing rounding must record a rounding adjustment")
	}
	if !result.GrandTotal.Equal(result.GrandTotal.Round(0)) {
		t.Errorf("grand total %s is not rounded to the currency scale", result.GrandTotal)
	}
	// The adjustment must account for exactly the difference it made.
	if !rounding.BaseAmount.Add(rounding.Amount).Equal(result.GrandTotal) {
		t.Errorf("the rounding adjustment does not reconcile: %s + %s != %s",
			rounding.BaseAmount, rounding.Amount, result.GrandTotal)
	}
}

// An already-round total must not produce a rounding adjustment.
func TestNoRoundingAdjustmentWhenNothingToRound(t *testing.T) {
	result := Calculate(Input{
		Lines:   []LineInput{lineOf("a", 1, "2", "5000")},
		Context: vndContext(),
	})

	for _, adjustment := range result.Adjustments {
		if adjustment.Type == AdjustmentRounding {
			t.Errorf("an exact total produced a rounding adjustment of %s", adjustment.Amount)
		}
	}
}

// Tax comes from Accounting and the rate is snapshotted onto every line, so a later rate change
// cannot reinterpret a historical sale. This asserts the engine carries Accounting's figures through
// faithfully rather than re-deriving them.
func TestTaxComesFromAccountingAndIsSnapshotted(t *testing.T) {
	// A 110,000 tax-inclusive line at 10%: Accounting extracts 10,000, not 11,000.
	result := Calculate(Input{
		Lines:   []LineInput{lineOf("a", 1, "1", "110000")},
		Context: taxedContext("0.1", map[string]string{"a": "10000"}),
	})

	line := lineByKey(t, result, "a")
	if !line.TaxAmount.Equal(dec("10000")) {
		t.Errorf("tax = %s, want the 10000 Accounting computed", line.TaxAmount)
	}
	// The grand total is unchanged by tax: with tax-inclusive pricing it was always inside the price.
	if !result.GrandTotal.Equal(dec("110000")) {
		t.Errorf("grand total = %s, want 110000: recording tax must not change what is owed",
			result.GrandTotal)
	}
	// A FRACTION, not a percentage; the boundary converts. See services.effectiveRateFraction.
	if !line.TaxRateSnapshot.Equal(dec("0.1")) {
		t.Errorf("tax rate snapshot = %s, want the fraction 0.1 rather than the percentage 10",
			line.TaxRateSnapshot)
	}
}

// A line Accounting was not asked about is taxed at zero rather than left undefined — the giveaway
// case, where the engine must still produce a complete line.
func TestLineWithNoTaxEntryIsTaxedAtZero(t *testing.T) {
	result := Calculate(Input{
		Lines: []LineInput{
			lineOf("a", 1, "1", "110000"),
			lineOf("b", 2, "1", "50000"),
		},
		Context: taxedContext("0.1", map[string]string{"a": "10000"}),
	})

	absent := lineByKey(t, result, "b")
	if !absent.TaxAmount.IsZero() || !absent.TaxRateSnapshot.IsZero() {
		t.Errorf("a line Accounting did not tax must read as zero, got amount %s rate %s",
			absent.TaxAmount, absent.TaxRateSnapshot)
	}
	if !result.TaxTotal.Equal(dec("10000")) {
		t.Errorf("tax total = %s, want 10000 from the one taxed line", result.TaxTotal)
	}
}

// With no tax configured the step still runs, writing real zeros rather than leaving fields unset.
func TestZeroTaxRateStillPopulatesTheFields(t *testing.T) {
	result := Calculate(Input{
		Lines:   []LineInput{lineOf("a", 1, "1", "50000")},
		Context: vndContext(),
	})

	line := lineByKey(t, result, "a")
	if !line.TaxAmount.IsZero() || !line.TaxRateSnapshot.IsZero() {
		t.Errorf("with no tax configured both must be zero, got amount %s rate %s",
			line.TaxAmount, line.TaxRateSnapshot)
	}
	if !result.TaxTotal.IsZero() {
		t.Errorf("tax total = %s, want zero", result.TaxTotal)
	}
}

// The same input produces the same output, every time.
// The engine must not mutate the caller's input, or a basket reused for preview then confirm would
// give two different answers.
// A line-scoped reward applies only to the lines it names.
// Conditional bundle pricing: a program whose reward sets a per-unit price. Not a combo.
// Adjustment sequences must be strictly increasing; the sequence is what makes the replay exact.
