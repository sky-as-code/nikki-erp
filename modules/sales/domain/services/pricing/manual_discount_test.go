package pricing

import (
	"testing"
)

// Manual overrides (BR §87.4, SALES-039).
//
// The rule the whole feature turns on: the base price is NEVER overwritten. The line keeps what the
// catalogue and pricelist gave it, and the override is one visible link in the chain — which is what
// keeps BR §87.9's explanation readable, and what stops an override being indistinguishable from a
// catalogue price that happened to be low.

func manualOf(id, lineKey, amount, reason string) ManualDiscountInput {
	return ManualDiscountInput{
		Id:      id,
		LineKey: lineKey,
		Amount:  dec(amount),
		Reason:  reason,
	}
}

// THE rule. The gross amount — what the goods list at — is untouched; only the discount and the net
// move. An override that rewrote the base price would leave the order saying the goods cost less
// than they do, and nothing would record that anyone had decided so.
func TestAManualDiscountNeverRewritesTheBasePrice(t *testing.T) {
	result := Calculate(Input{
		Lines:           []LineInput{lineOf("a", 1, "2", "50000")},
		ManualDiscounts: []ManualDiscountInput{manualOf("MD1", "a", "10000", "Damaged packaging")},
		Context:         vndContext(),
	})

	line := lineByKey(t, result, "a")
	if !line.GrossAmount.Equal(dec("100000")) {
		t.Errorf("gross = %s, want 100000 — the base price must survive the override",
			line.GrossAmount)
	}
	if !line.DiscountAmount.Equal(dec("10000")) {
		t.Errorf("discount = %s, want 10000", line.DiscountAmount)
	}
	if !line.NetAmount.Equal(dec("90000")) {
		t.Errorf("net = %s, want 90000", line.NetAmount)
	}
}

// The override appears in the chain, typed and attributed, with the operator's reason as its
// description. Without this the explanation would show a number changing for no stated cause.
func TestAManualDiscountIsExplainable(t *testing.T) {
	result := Calculate(Input{
		Lines:           []LineInput{lineOf("a", 1, "1", "80000")},
		ManualDiscounts: []ManualDiscountInput{manualOf("MD9", "a", "8000", "Loyalty gesture")},
		Context:         vndContext(),
	})

	var found bool
	for _, adjustment := range result.Adjustments {
		if adjustment.Type != AdjustmentManualDiscount {
			continue
		}
		found = true

		if adjustment.Description != "Loyalty gesture" {
			t.Errorf("description = %q, want the operator's reason", adjustment.Description)
		}
		if adjustment.SourceId != "MD9" {
			t.Errorf("source id = %q, want the stored override MD9", adjustment.SourceId)
		}
		if !adjustment.Amount.Equal(dec("-8000")) {
			t.Errorf("amount = %s, want -8000: an adjustment that reduces a price is signed "+
				"negative, like every other discount in the chain", adjustment.Amount)
		}
		if !adjustment.BaseAmount.Equal(dec("80000")) {
			t.Errorf("base = %s, want 80000 — what it was calculated from", adjustment.BaseAmount)
		}
	}
	if !found {
		t.Fatal("a manual discount must appear in the adjustment chain")
	}
}

// An order-level override — no line named — must actually reduce what the customer pays.
//
// The failure this guards is specific and was real: step 6 ignores the adjustment list and totalise()
// sums the LINES alone, so an adjustment touching no line would show on the paperwork as a discount
// the customer never received.
func TestAnOrderLevelOverrideReachesTheTotal(t *testing.T) {
	result := Calculate(Input{
		Lines: []LineInput{
			lineOf("a", 1, "1", "60000"),
			lineOf("b", 2, "1", "40000"),
		},
		ManualDiscounts: []ManualDiscountInput{manualOf("MD2", "", "10000", "Goodwill")},
		Context:         vndContext(),
	})

	if !result.GrandTotal.Equal(dec("90000")) {
		t.Errorf("grand total = %s, want 90000 — an order-level override must reach the total, "+
			"not merely appear beside it", result.GrandTotal)
	}

	// And it lands proportionally, by the same allocator every other order-level discount uses.
	if got := lineByKey(t, result, "a").DiscountAmount; !got.Equal(dec("6000")) {
		t.Errorf("line a discount = %s, want 6000 (60%% of the basket)", got)
	}
	if got := lineByKey(t, result, "b").DiscountAmount; !got.Equal(dec("4000")) {
		t.Errorf("line b discount = %s, want 4000 (40%% of the basket)", got)
	}
}

// The spread is EXACT: the parts sum to the whole, with no residual lost to rounding. D-04's
// allocator is what guarantees it, and this asserts the guarantee holds through this path too.
func TestAnOrderLevelOverrideSpreadsExactly(t *testing.T) {
	result := Calculate(Input{
		Lines: []LineInput{
			lineOf("a", 1, "1", "33333"),
			lineOf("b", 2, "1", "33333"),
			lineOf("c", 3, "1", "33334"),
		},
		ManualDiscounts: []ManualDiscountInput{manualOf("MD3", "", "10000", "Goodwill")},
		Context:         vndContext(),
	})

	total := dec("0")
	for _, line := range result.Lines {
		total = total.Add(line.DiscountAmount)
	}
	if !total.Equal(dec("10000")) {
		t.Errorf("the spread discounts sum to %s, want exactly 10000", total)
	}
}

// An override cannot take a line below zero. Past the end is a REFUND — its own workflow with its own
// money movement — not a discount that overshot.
func TestAnOverrideIsCappedAtWhatIsOwed(t *testing.T) {
	result := Calculate(Input{
		Lines:           []LineInput{lineOf("a", 1, "1", "50000")},
		ManualDiscounts: []ManualDiscountInput{manualOf("MD4", "a", "80000", "Too generous")},
		Context:         vndContext(),
	})

	line := lineByKey(t, result, "a")
	if line.NetAmount.IsNegative() {
		t.Errorf("net = %s: an override must never make the customer owe a negative amount",
			line.NetAmount)
	}
	if !line.NetAmount.IsZero() {
		t.Errorf("net = %s, want 0 — capped at what was owed", line.NetAmount)
	}
	if result.GrandTotal.IsNegative() {
		t.Errorf("grand total = %s must not be negative", result.GrandTotal)
	}
}

// A non-positive override is ignored. Zero changes nothing, and a NEGATIVE one is a surcharge —
// silently adding money to a customer's bill is the worst failure available here, and BR §87.4
// authorises a discount, not a markup.
func TestANegativeOverrideIsNotASurcharge(t *testing.T) {
	for _, amount := range []string{"0", "-5000"} {
		result := Calculate(Input{
			Lines:           []LineInput{lineOf("a", 1, "1", "50000")},
			ManualDiscounts: []ManualDiscountInput{manualOf("MDx", "a", amount, "Nope")},
			Context:         vndContext(),
		})

		if !result.GrandTotal.Equal(dec("50000")) {
			t.Errorf("amount %s changed the total to %s; a non-positive override must be ignored",
				amount, result.GrandTotal)
		}
		for _, adjustment := range result.Adjustments {
			if adjustment.Type == AdjustmentManualDiscount {
				t.Errorf("amount %s produced an adjustment; it must not", amount)
			}
		}
	}
}

// An override naming a line that does not exist must not panic the whole calculation. The engine is
// pure and takes its input on trust, so it declines one bad row rather than failing the sale.
func TestAnOverrideNamingNoLineIsDeclined(t *testing.T) {
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("an override naming an unknown line panicked: %v", recovered)
		}
	}()

	result := Calculate(Input{
		Lines:           []LineInput{lineOf("a", 1, "1", "50000")},
		ManualDiscounts: []ManualDiscountInput{manualOf("MD5", "ghost", "1000", "Wrong line")},
		Context:         vndContext(),
	})

	if !result.GrandTotal.Equal(dec("50000")) {
		t.Errorf("grand total = %s, want 50000 — an override naming no line changes nothing",
			result.GrandTotal)
	}
}

// Overrides REPLAY. This is why they are engine input rather than adjustments written afterwards:
// repricing deletes the whole chain and rewrites it from engine output, and confirm reprices
// unconditionally — so an override written outside the engine would vanish before the sale completed.
func TestOverridesSurviveRecalculation(t *testing.T) {
	input := Input{
		Lines:           []LineInput{lineOf("a", 1, "2", "50000")},
		ManualDiscounts: []ManualDiscountInput{manualOf("MD6", "a", "15000", "Manager approval")},
		Context:         vndContext(),
	}

	first := Calculate(input)
	second := Calculate(input)

	if !first.GrandTotal.Equal(second.GrandTotal) {
		t.Errorf("recalculating moved the total from %s to %s; the same input must give the "+
			"same output (BR 13)", first.GrandTotal, second.GrandTotal)
	}
	if !second.GrandTotal.Equal(dec("85000")) {
		t.Errorf("grand total = %s, want 85000 on every run", second.GrandTotal)
	}
}

// Several overrides on one order all apply, each explainable on its own. An operator may discount one
// line for damage and another as a gesture, and collapsing them would lose both reasons.
func TestSeveralOverridesEachApply(t *testing.T) {
	result := Calculate(Input{
		Lines: []LineInput{
			lineOf("a", 1, "1", "50000"),
			lineOf("b", 2, "1", "50000"),
		},
		ManualDiscounts: []ManualDiscountInput{
			manualOf("MD7", "a", "5000", "Damaged"),
			manualOf("MD8", "b", "3000", "Gesture"),
		},
		Context: vndContext(),
	})

	if !result.GrandTotal.Equal(dec("92000")) {
		t.Errorf("grand total = %s, want 92000", result.GrandTotal)
	}

	reasons := map[string]bool{}
	for _, adjustment := range result.Adjustments {
		if adjustment.Type == AdjustmentManualDiscount {
			reasons[adjustment.Description] = true
		}
	}
	for _, reason := range []string{"Damaged", "Gesture"} {
		if !reasons[reason] {
			t.Errorf("the reason %q must survive into the explanation", reason)
		}
	}
}

// The override is taxed. It happens before the tax step because it changes what is being taxed —
// discounting goods and then charging tax on the undiscounted price would overcharge the customer.
func TestAnOverrideChangesWhatIsTaxed(t *testing.T) {
	// Tax supplied for the DISCOUNTED base, which is what ResolveBasketTax would compute: the
	// caller reprices, then asks Accounting about the net it arrived at.
	result := Calculate(Input{
		Lines:           []LineInput{lineOf("a", 1, "1", "110000")},
		ManualDiscounts: []ManualDiscountInput{manualOf("MD10", "a", "10000", "Gesture")},
		Context:         taxedContext("0.1", map[string]string{"a": "10000"}),
	})

	line := lineByKey(t, result, "a")
	if !line.NetAmount.Equal(dec("100000")) {
		t.Fatalf("net = %s, want 100000 before tax is added", line.NetAmount)
	}
	if !line.TaxAmount.Equal(dec("10000")) {
		t.Errorf("tax = %s, want the 10000 Accounting returned for the discounted base",
			line.TaxAmount)
	}
}
