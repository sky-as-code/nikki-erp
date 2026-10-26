package services

import (
	"math/rand"
	"testing"

	"github.com/shopspring/decimal"
)

// The invariant is `Σ allocated == total`, EXACTLY, asserted by nearly every test below.

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

// References 30/20/10 over 48,000 divide evenly, so there is no residual.
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

// 100 across three equal shares at two places is 99.99; the hundredth must be placed.
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

// The residual goes to the LARGEST reference, not the last input and not the first.
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

// Input order must not change the answer, or the same order priced twice yields different numbers.
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

// Every reference zero: the total must still be allocated somewhere, or the money vanishes.
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

// A zero-priced component in a paid bundle takes no share.
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

// A discount allocated across lines is negative; the residual rule must still be deterministic.
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

	// The single share is the UNROUNDED total, because the residual is added back after rounding.
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

// The exactness invariant across a wide range of awkward inputs.
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
