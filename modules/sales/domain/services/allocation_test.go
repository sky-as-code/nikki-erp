package services

import (
	"math/rand"
	"testing"

	"github.com/shopspring/decimal"
)

// The allocation invariant is `Σ allocated == total`, EXACTLY, for every input. It is asserted in
// nearly every test below rather than in one, because it is the property the whole algorithm exists
// to guarantee: a combo whose components do not sum to its price makes the receipt and the stock
// valuation disagree, and a split bill that does not sum to its order loses money.

func alloc(key string, reference string, tiebreak int32) AllocationInput {
	return AllocationInput{
		Key:       key,
		Reference: decimal.RequireFromString(reference),
		Tiebreak:  tiebreak,
	}
}

func amountOf(results []AllocationResult, key string) decimal.Decimal {
	for _, result := range results {
		if result.Key == key {
			return result.Amount
		}
	}
	return decimal.RequireFromString("-999999")
}

func assertSumsExactly(t *testing.T, results []AllocationResult, total string) {
	t.Helper()
	want := decimal.RequireFromString(total)
	if got := AllocationSum(results); !got.Equal(want) {
		t.Errorf("allocations sum to %s, want exactly %s — the residual rule failed", got, want)
	}
}

// BR §18's worked example, spelled out: references 30/20/10, combo price 48,000, expected shares
// 24,000/16,000/8,000. It divides evenly, so there is no residual and the proportions are visible.
func TestAllocateWorkedExampleFromBR18(t *testing.T) {
	results := Allocate(
		decimal.RequireFromString("48000"),
		[]AllocationInput{
			alloc("a", "30", 1),
			alloc("b", "20", 2),
			alloc("c", "10", 3),
		},
		0,
	)

	cases := map[string]string{"a": "24000", "b": "16000", "c": "8000"}
	for key, want := range cases {
		if got := amountOf(results, key); !got.Equal(decimal.RequireFromString(want)) {
			t.Errorf("%s allocated %s, want %s", key, got, want)
		}
	}
	assertSumsExactly(t, results, "48000")
}

// The case the residual rule exists for: 100 across three equal shares at two places is 33.33 three
// times, which is 99.99. The missing hundredth must land somewhere reproducible.
func TestResidualIsAssignedAndSumIsExact(t *testing.T) {
	results := Allocate(
		decimal.RequireFromString("100"),
		[]AllocationInput{alloc("a", "1", 1), alloc("b", "1", 2), alloc("c", "1", 3)},
		2,
	)

	assertSumsExactly(t, results, "100")

	// All references are equal, so the tiebreak decides: lowest wins.
	if got := amountOf(results, "a"); !got.Equal(decimal.RequireFromString("33.34")) {
		t.Errorf("the residual must go to the lowest tiebreak among equal references, "+
			"a got %s want 33.34", got)
	}
}

// D-04: the residual goes to the LARGEST reference, not the last input and not the first.
func TestResidualGoesToLargestReference(t *testing.T) {
	// 10 split by 1/1/1 at 0 places: each share is 3.33 -> 3, summing to 9, residual 1.
	results := Allocate(
		decimal.RequireFromString("10"),
		[]AllocationInput{alloc("small", "1", 1), alloc("big", "5", 2), alloc("mid", "3", 3)},
		0,
	)
	assertSumsExactly(t, results, "10")

	// big has the largest reference, so it absorbs whatever rounding left over.
	shares := map[string]decimal.Decimal{
		"small": amountOf(results, "small"),
		"big":   amountOf(results, "big"),
		"mid":   amountOf(results, "mid"),
	}
	proportional := map[string]string{"small": "1", "big": "6", "mid": "3"}
	for key, want := range proportional {
		if !shares[key].Equal(decimal.RequireFromString(want)) {
			t.Errorf("%s allocated %s, want %s (residual belongs to the largest reference)",
				key, shares[key], want)
		}
	}
}

// The order the caller happens to pass inputs in must not change the answer. Anything else would
// make the same order priced twice produce different numbers.
func TestAllocationIsIndependentOfInputOrder(t *testing.T) {
	inputs := []AllocationInput{
		alloc("a", "7", 1), alloc("b", "11", 2), alloc("c", "13", 3), alloc("d", "17", 4),
	}
	total := decimal.RequireFromString("1000")

	baseline := Allocate(total, inputs, 2)
	assertSumsExactly(t, baseline, "1000")

	random := rand.New(rand.NewSource(42))
	for attempt := 0; attempt < 20; attempt++ {
		shuffled := make([]AllocationInput, len(inputs))
		copy(shuffled, inputs)
		random.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})

		got := Allocate(total, shuffled, 2)
		assertSumsExactly(t, got, "1000")
		for _, input := range inputs {
			if !amountOf(got, input.Key).Equal(amountOf(baseline, input.Key)) {
				t.Fatalf("shuffling the input changed %s: %s vs %s",
					input.Key, amountOf(got, input.Key), amountOf(baseline, input.Key))
			}
		}
	}
}

// A bundle of free items: every reference is zero, so proportions are undefined. The total must
// still be allocated somewhere, or the sum invariant breaks and the money vanishes.
func TestAllZeroReferencesStillAllocatesTheTotal(t *testing.T) {
	results := Allocate(
		decimal.RequireFromString("500"),
		[]AllocationInput{alloc("a", "0", 2), alloc("b", "0", 1), alloc("c", "0", 3)},
		2,
	)
	assertSumsExactly(t, results, "500")

	// References tie at zero, so the lowest tiebreak takes it.
	if got := amountOf(results, "b"); !got.Equal(decimal.RequireFromString("500")) {
		t.Errorf("with every reference zero the total goes to the lowest tiebreak, b got %s", got)
	}
}

// A component priced at zero inside a paid bundle takes no share — it is free, and giving it part
// of the bundle price would misstate both its value and its neighbours'.
func TestZeroReferenceTakesNoShare(t *testing.T) {
	results := Allocate(
		decimal.RequireFromString("100"),
		[]AllocationInput{alloc("paid", "10", 1), alloc("free", "0", 2)},
		2,
	)
	assertSumsExactly(t, results, "100")

	if got := amountOf(results, "free"); !got.IsZero() {
		t.Errorf("a zero-reference component must take no share, got %s", got)
	}
	if got := amountOf(results, "paid"); !got.Equal(decimal.RequireFromString("100")) {
		t.Errorf("the paid component takes everything, got %s", got)
	}
}

// A discount allocated across lines is negative. Proportions are unaffected by sign, and the
// residual rule must still land the remainder deterministically.
func TestNegativeTotalAllocatesLikeAPositiveOne(t *testing.T) {
	results := Allocate(
		decimal.RequireFromString("-100"),
		[]AllocationInput{alloc("a", "1", 1), alloc("b", "1", 2), alloc("c", "1", 3)},
		2,
	)
	assertSumsExactly(t, results, "-100")

	for _, result := range results {
		if result.Amount.IsPositive() {
			t.Errorf("%s got %s: allocating a negative total must not produce a positive share",
				result.Key, result.Amount)
		}
	}
}

func TestAllocateEmptyInput(t *testing.T) {
	if got := Allocate(decimal.RequireFromString("100"), nil, 2); got != nil {
		t.Errorf("allocating across no recipients = %v, want nil", got)
	}
}

func TestAllocateSingleRecipientTakesEverything(t *testing.T) {
	results := Allocate(
		decimal.RequireFromString("33.333333"),
		[]AllocationInput{alloc("only", "5", 1)},
		2,
	)
	assertSumsExactly(t, results, "33.333333")

	// Note what this asserts: the single share is the UNROUNDED total, because the residual is
	// added back after rounding. That is correct — the invariant is exactness, not tidiness, and a
	// caller wanting a rounded total rounds the total before allocating it.
	if got := amountOf(results, "only"); !got.Equal(decimal.RequireFromString("33.333333")) {
		t.Errorf("a single recipient takes the whole total exactly, got %s", got)
	}
}

// A zero total allocates zero to everyone and still sums exactly.
func TestAllocateZeroTotal(t *testing.T) {
	results := Allocate(
		decimal.Zero,
		[]AllocationInput{alloc("a", "10", 1), alloc("b", "20", 2)},
		2,
	)
	assertSumsExactly(t, results, "0")
	for _, result := range results {
		if !result.Amount.IsZero() {
			t.Errorf("%s got %s, want zero", result.Key, result.Amount)
		}
	}
}

// The exactness invariant across a wide range of awkward inputs. This is the test that would catch
// a regression in the residual logic on a case nobody thought to write by hand.
func TestSumIsExactAcrossManyShapes(t *testing.T) {
	totals := []string{"0.01", "1", "7", "99.99", "100", "1000.005", "48000", "123456.789"}
	references := [][]string{
		{"1"},
		{"1", "1"},
		{"1", "1", "1"},
		{"1", "2", "3"},
		{"30", "20", "10"},
		{"0.001", "999999", "1"},
		{"7", "7", "7", "7", "7", "7", "7"},
		{"0", "1", "0", "2"},
	}
	scales := []int32{0, 2, 4, 6}

	for _, total := range totals {
		for _, referenceSet := range references {
			for _, scale := range scales {
				inputs := make([]AllocationInput, len(referenceSet))
				for index, reference := range referenceSet {
					inputs[index] = alloc(string(rune('a'+index)), reference, int32(index))
				}

				results := Allocate(decimal.RequireFromString(total), inputs, scale)
				want := decimal.RequireFromString(total)
				if got := AllocationSum(results); !got.Equal(want) {
					t.Errorf("total=%s references=%v scale=%d: sum %s != %s",
						total, referenceSet, scale, got, want)
				}
			}
		}
	}
}
