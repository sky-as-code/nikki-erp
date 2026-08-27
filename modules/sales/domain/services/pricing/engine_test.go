package pricing

import (
	"testing"

	"github.com/shopspring/decimal"
)

// BR §13's "same input ⇒ same output, always" is the acceptance test for this package, and the
// engine has no I/O, so there is no excuse for thin coverage.
//
// Two invariants recur and matter more than any individual number:
//   - Σ line net == order net at every step (D-05's reason for using post-discount net as the
//     allocation basis)
//   - Σ component allocated == the combo line's amount, exactly (D-04)

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
	// VND has no minor unit. No tax entries means every line is taxed at zero, which is what an
	// organization with no default tax code configured gets.
	return Context{CurrencyScale: 0}
}

// taxedContext taxes each named line at the given amount and fraction rate, standing in for what
// ResolveBasketTax returns after Accounting has computed it.
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

// assertLineNetSumsToOrderNet is the invariant D-05 exists to preserve.
func assertLineNetSumsToOrderNet(t *testing.T, result Result) {
	t.Helper()
	total := decimal.Zero
	for _, line := range result.Lines {
		total = total.Add(line.NetAmount)
	}
	// The grand total may differ by the rounding adjustment, which is applied to the order rather
	// than to the lines. Compare against the pre-rounding sum instead.
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

	// A line of 2 cannot reach the ten-break, so the specific list's base price applies — and it
	// beats the global list despite that list's much higher priority.
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

// BR §18's worked example, through the engine: references 30/20/10, combo price 48,000, expected
// component shares 24,000/16,000/8,000.
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

	// D-04's invariant, which is the whole reason the residual rule exists.
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

// A percentage discount, allocated across lines by D-05's post-discount-net basis.
func TestPercentageDiscount(t *testing.T) {
	result := Calculate(Input{
		Lines: []LineInput{
			lineOf("a", 1, "1", "60000"),
			lineOf("b", 2, "1", "40000"),
		},
		Programs: []AppliedProgram{{
			ProgramId: "p1", ProgramName: "Ten percent off",
			Rewards: []RewardInput{{
				RewardId: "r1", Sequence: 1, Type: "percentage_discount",
				Value: dec("10"), TargetScope: "order",
			}},
		}},
		Context: vndContext(),
	})

	if !result.DiscountTotal.Equal(dec("10000")) {
		t.Errorf("discount total = %s, want 10000", result.DiscountTotal)
	}
	if !result.GrandTotal.Equal(dec("90000")) {
		t.Errorf("grand total = %s, want 90000", result.GrandTotal)
	}
	// Proportional to net: 60/40 of a 10,000 discount.
	if got := lineByKey(t, result, "a").DiscountAmount; !got.Equal(dec("6000")) {
		t.Errorf("line a discount = %s, want 6000", got)
	}
	if got := lineByKey(t, result, "b").DiscountAmount; !got.Equal(dec("4000")) {
		t.Errorf("line b discount = %s, want 4000", got)
	}
	assertLineNetSumsToOrderNet(t, result)

	// BR §13 requires a list of adjustments, never a silent total.
	if len(result.Adjustments) != 1 {
		t.Fatalf("expected one adjustment, got %d", len(result.Adjustments))
	}
	if !result.Adjustments[0].Amount.Equal(dec("-10000")) {
		t.Errorf("the adjustment must be signed negative, got %s", result.Adjustments[0].Amount)
	}
}

// A fixed discount larger than the basket is capped: a negative total would pay the customer to shop.
func TestFixedDiscountIsCappedAtTheBasketValue(t *testing.T) {
	result := Calculate(Input{
		Lines: []LineInput{lineOf("a", 1, "1", "5000")},
		Programs: []AppliedProgram{{
			ProgramId: "p1", ProgramName: "Big voucher", VoucherCode: "SAVE100K",
			Rewards: []RewardInput{{
				RewardId: "r1", Sequence: 1, Type: "fixed_amount_discount",
				Value: dec("100000"), TargetScope: "order",
			}},
		}},
		Context: vndContext(),
	})

	if result.GrandTotal.IsNegative() {
		t.Errorf("grand total = %s: a discount must never take an order below zero",
			result.GrandTotal)
	}
	if !result.GrandTotal.IsZero() {
		t.Errorf("grand total = %s, want 0", result.GrandTotal)
	}
	// A voucher-activated program records its adjustment as a voucher, not a plain discount, so the
	// price explanation can say where the discount came from.
	if result.Adjustments[0].Type != AdjustmentVoucher {
		t.Errorf("adjustment type = %q, want voucher", result.Adjustments[0].Type)
	}
}

// D-11: a free item becomes a REAL line at zero price, not an adjustment.
func TestFreeQuantityCreatesARealLine(t *testing.T) {
	result := Calculate(Input{
		Lines: []LineInput{lineOf("a", 1, "2", "50000")},
		Programs: []AppliedProgram{{
			ProgramId: "p1", ProgramName: "Buy two get one",
			Rewards: []RewardInput{{
				RewardId: "free-1", Sequence: 1, Type: "free_quantity", Value: dec("1"),
				FreeProductVariantId: "variant-a", FreeUomId: "uom-each",
				FreeProductCode: "CODE-a", FreeProductName: "Product a",
			}},
		}},
		Context: vndContext(),
	})

	if len(result.Lines) != 2 {
		t.Fatalf("a free_quantity reward must add a line, got %d lines", len(result.Lines))
	}
	free := lineByKey(t, result, "free-1")
	if free.LineType != "promotion_reward" {
		t.Errorf("line_type = %q, want promotion_reward", free.LineType)
	}
	if !free.EffectiveUnitPrice.IsZero() || !free.NetAmount.IsZero() {
		t.Errorf("a giveaway line must be free, got unit %s net %s",
			free.EffectiveUnitPrice, free.NetAmount)
	}
	if free.SourcePromotionProgramId != "p1" {
		t.Errorf("the giveaway must name the campaign that gave it away, got %q",
			free.SourcePromotionProgramId)
	}
	// Inventory must physically fulfil it, which is half the reason it is a line at all.
	if !free.RequiresFulfillment {
		t.Error("a giveaway line must require fulfilment: Inventory has to hand it over")
	}
	if !result.GrandTotal.Equal(dec("100000")) {
		t.Errorf("grand total = %s, want 100000: the free line adds nothing", result.GrandTotal)
	}
}

// A giveaway line is never itself discounted — it is already free, and allocating a share to it
// would take that share from a line the customer is actually paying for.
func TestGiveawayLineIsNotDiscounted(t *testing.T) {
	result := Calculate(Input{
		Lines: []LineInput{lineOf("a", 1, "1", "100000")},
		Programs: []AppliedProgram{
			{
				ProgramId: "p1", ProgramName: "Free gift",
				Rewards: []RewardInput{{
					RewardId: "free-1", Sequence: 1, Type: "free_quantity", Value: dec("1"),
					FreeProductVariantId: "v-gift", FreeUomId: "uom-each",
				}},
			},
			{
				ProgramId: "p2", ProgramName: "Ten percent",
				Rewards: []RewardInput{{
					RewardId: "r2", Sequence: 1, Type: "percentage_discount",
					Value: dec("10"), TargetScope: "order",
				}},
			},
		},
		Context: vndContext(),
	})

	if got := lineByKey(t, result, "free-1").DiscountAmount; !got.IsZero() {
		t.Errorf("the giveaway line was discounted by %s: it is already free", got)
	}
	if got := lineByKey(t, result, "a").DiscountAmount; !got.Equal(dec("10000")) {
		t.Errorf("the paid line should absorb the whole discount, got %s", got)
	}
}

// Rewards do not commute, so the sequence decides the answer. This is the reason adjustments carry
// one and the reason BR §13 fixes the step order.
func TestRewardOrderChangesTheTotal(t *testing.T) {
	percentFirst := Calculate(Input{
		Lines: []LineInput{lineOf("a", 1, "1", "100000")},
		Programs: []AppliedProgram{{
			ProgramId: "p1", ProgramName: "Stacked",
			Rewards: []RewardInput{
				{RewardId: "r1", Sequence: 1, Type: "percentage_discount", Value: dec("10")},
				{RewardId: "r2", Sequence: 2, Type: "fixed_amount_discount", Value: dec("20000")},
			},
		}},
		Context: vndContext(),
	})

	fixedFirst := Calculate(Input{
		Lines: []LineInput{lineOf("a", 1, "1", "100000")},
		Programs: []AppliedProgram{{
			ProgramId: "p1", ProgramName: "Stacked",
			Rewards: []RewardInput{
				{RewardId: "r2", Sequence: 1, Type: "fixed_amount_discount", Value: dec("20000")},
				{RewardId: "r1", Sequence: 2, Type: "percentage_discount", Value: dec("10")},
			},
		}},
		Context: vndContext(),
	})

	// 10% of 100,000 then 20,000 = 70,000; 20,000 then 10% of 80,000 = 72,000.
	if !percentFirst.GrandTotal.Equal(dec("70000")) {
		t.Errorf("percentage first = %s, want 70000", percentFirst.GrandTotal)
	}
	if !fixedFirst.GrandTotal.Equal(dec("72000")) {
		t.Errorf("fixed first = %s, want 72000", fixedFirst.GrandTotal)
	}
	if percentFirst.GrandTotal.Equal(fixedFirst.GrandTotal) {
		t.Error("reward order must change the total: discounts do not commute")
	}
}

// Step 8: rounding is applied once, to the grand total, and recorded as its own adjustment so the
// difference between the line sum and the total is explained rather than unaccounted for.
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

// An already-round total must not produce a rounding adjustment: rounding is applied once and only
// when it changes something.
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

// Tax is taken from what Accounting decided, and the rate is snapshotted onto every line so a later
// rate change cannot reinterpret a historical sale.
//
// The engine no longer extracts tax itself. It used to compute net × rate / (1 + rate); that
// arithmetic now lives in accounting, which also handles the compound, division and fixed cases this
// engine never did. What is tested here is that the engine carries Accounting's figures through
// faithfully and does not re-derive them.
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
	// A FRACTION, not a percentage. The boundary converts; see services.effectiveRateFraction.
	if !line.TaxRateSnapshot.Equal(dec("0.1")) {
		t.Errorf("tax rate snapshot = %s, want the fraction 0.1 rather than the percentage 10",
			line.TaxRateSnapshot)
	}
}

// A line Accounting was not asked about is taxed at zero rather than left undefined.
//
// This is the giveaway-line case: a free item may be excluded from the tax request, and the engine
// must still produce a complete, readable line for it.
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

// With no tax configured the step still runs, writing real zeros rather than leaving the fields unset.
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

// BR §13's acceptance criterion: the same input produces the same output, every time.
func TestSameInputSameOutput(t *testing.T) {
	build := func() Input {
		comboLine := lineOf("combo", 3, "2", "0")
		comboLine.ComboId = "combo-1"
		return Input{
			Lines: []LineInput{
				lineOf("a", 1, "3", "17000"),
				lineOf("b", 2, "1", "23500"),
				comboLine,
			},
			PricelistItems: []PricelistItem{{
				ProductVariantId: "variant-a", UomId: "uom-each",
				UnitPrice: dec("15500"), MinQuantity: dec("2"), Specificity: 1,
			}},
			Combos: []ComboDefinition{{
				ComboId: "combo-1", ComboPrice: dec("48000"),
				Components: []ComboComponentInput{
					{Key: "c1", Sequence: 1, ProductVariantId: "v1", UomId: "u",
						Quantity: dec("1"), ReferencePrice: dec("30")},
					{Key: "c2", Sequence: 2, ProductVariantId: "v2", UomId: "u",
						Quantity: dec("2"), ReferencePrice: dec("20")},
				},
			}},
			Programs: []AppliedProgram{{
				ProgramId: "p1", ProgramName: "Seasonal", VoucherCode: "SUMMER",
				Rewards: []RewardInput{
					{RewardId: "r1", Sequence: 1, Type: "percentage_discount", Value: dec("7.5")},
					{RewardId: "r2", Sequence: 2, Type: "fixed_amount_discount", Value: dec("1234")},
				},
			}},
			Context: Context{CurrencyScale: 0},
		}
	}

	first := Calculate(build())
	for attempt := 0; attempt < 10; attempt++ {
		again := Calculate(build())

		if !again.GrandTotal.Equal(first.GrandTotal) ||
			!again.Subtotal.Equal(first.Subtotal) ||
			!again.DiscountTotal.Equal(first.DiscountTotal) ||
			!again.TaxTotal.Equal(first.TaxTotal) {
			t.Fatalf("totals differed between runs: %+v vs %+v", again, first)
		}
		if len(again.Lines) != len(first.Lines) ||
			len(again.Adjustments) != len(first.Adjustments) {
			t.Fatalf("shape differed between runs")
		}
		for index := range first.Lines {
			if !again.Lines[index].NetAmount.Equal(first.Lines[index].NetAmount) {
				t.Fatalf("line %s net differed: %s vs %s", first.Lines[index].Key,
					again.Lines[index].NetAmount, first.Lines[index].NetAmount)
			}
		}
	}
	assertLineNetSumsToOrderNet(t, first)
}

// The engine must not mutate the caller's input. A caller reusing a basket for a preview and then a
// confirm would otherwise get two different answers from what it believed was one input.
func TestCalculateDoesNotMutateInput(t *testing.T) {
	input := Input{
		Lines:   []LineInput{lineOf("a", 1, "2", "10000")},
		Context: vndContext(),
		Programs: []AppliedProgram{{
			ProgramId: "p1", ProgramName: "Discount",
			Rewards: []RewardInput{
				{RewardId: "r2", Sequence: 2, Type: "percentage_discount", Value: dec("10")},
				{RewardId: "r1", Sequence: 1, Type: "fixed_amount_discount", Value: dec("500")},
			},
		}},
	}

	Calculate(input)

	if !input.Lines[0].Quantity.Equal(dec("2")) ||
		!input.Lines[0].CatalogueUnitPrice.Equal(dec("10000")) {
		t.Error("the input line was mutated")
	}
	// The engine sorts rewards by sequence internally; the caller's slice must be left alone.
	if input.Programs[0].Rewards[0].RewardId != "r2" {
		t.Errorf("the caller's reward slice was reordered: %v", input.Programs[0].Rewards)
	}
}

// A line-scoped reward applies only to the lines it names.
func TestLineScopedRewardTargetsOnlyItsLines(t *testing.T) {
	result := Calculate(Input{
		Lines: []LineInput{
			lineOf("a", 1, "1", "100000"),
			lineOf("b", 2, "1", "100000"),
		},
		Programs: []AppliedProgram{{
			ProgramId: "p1", ProgramName: "Line offer",
			Rewards: []RewardInput{{
				RewardId: "r1", Sequence: 1, Type: "percentage_discount", Value: dec("10"),
				TargetScope: "line", TargetLineKeys: []string{"a"},
			}},
		}},
		Context: vndContext(),
	})

	if got := lineByKey(t, result, "a").DiscountAmount; !got.Equal(dec("10000")) {
		t.Errorf("targeted line discount = %s, want 10000", got)
	}
	if got := lineByKey(t, result, "b").DiscountAmount; !got.IsZero() {
		t.Errorf("untargeted line was discounted by %s", got)
	}
}

// BR §20's conditional bundle pricing: a program whose reward sets a per-unit price. Not a combo.
func TestFixedProductPriceReward(t *testing.T) {
	result := Calculate(Input{
		Lines: []LineInput{lineOf("z", 1, "2", "50000")},
		Programs: []AppliedProgram{{
			ProgramId: "p1", ProgramName: "Buy X and Y, get Z at 10,000",
			Rewards: []RewardInput{{
				RewardId: "r1", Sequence: 1, Type: "fixed_product_price",
				Value: dec("10000"), TargetScope: "line", TargetLineKeys: []string{"z"},
			}},
		}},
		Context: vndContext(),
	})

	line := lineByKey(t, result, "z")
	if !line.EffectiveUnitPrice.Equal(dec("10000")) {
		t.Errorf("unit price = %s, want 10000", line.EffectiveUnitPrice)
	}
	if !result.GrandTotal.Equal(dec("20000")) {
		t.Errorf("grand total = %s, want 20000", result.GrandTotal)
	}
	if result.Adjustments[0].Type != AdjustmentConditionalPrice {
		t.Errorf("adjustment type = %q, want conditional_price", result.Adjustments[0].Type)
	}
}

// Adjustment sequences must be strictly increasing: the sequence is what makes the replay exact, so
// a duplicate or a gap in the wrong direction would make the calculation unreproducible.
func TestAdjustmentSequencesAreStrictlyIncreasing(t *testing.T) {
	result := Calculate(Input{
		Lines: []LineInput{lineOf("a", 1, "3", "33.3333")},
		Programs: []AppliedProgram{{
			ProgramId: "p1", ProgramName: "Several",
			Rewards: []RewardInput{
				{RewardId: "r1", Sequence: 1, Type: "percentage_discount", Value: dec("5")},
				{RewardId: "r2", Sequence: 2, Type: "fixed_amount_discount", Value: dec("3")},
			},
		}},
		Context: vndContext(),
	})

	for index := 1; index < len(result.Adjustments); index++ {
		if result.Adjustments[index].Sequence <= result.Adjustments[index-1].Sequence {
			t.Errorf("adjustment sequences are not strictly increasing: %d then %d",
				result.Adjustments[index-1].Sequence, result.Adjustments[index].Sequence)
		}
	}
}
