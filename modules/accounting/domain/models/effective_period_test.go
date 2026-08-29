package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/sky-as-code/nikki-erp/common/model"
)

func date(t *testing.T, value string) *model.ModelDate {
	t.Helper()
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		t.Fatalf("bad test date %q: %v", value, err)
	}
	result := model.ModelDate(parsed)
	return &result
}

// The engine picks the rate whose period contains tax_date.
func TestPeriodContains(t *testing.T) {
	cases := []struct {
		name    string
		from    string
		to      string
		taxDate string
		want    bool
	}{
		{"inside a closed period", "2025-07-01", "2026-12-31", "2025-09-15", true},
		{"on the first day", "2025-07-01", "2026-12-31", "2025-07-01", true},
		{"on the last day", "2025-07-01", "2026-12-31", "2026-12-31", true},
		{"the day before it starts", "2025-07-01", "2026-12-31", "2025-06-30", false},
		{"the day after it ends", "2025-07-01", "2026-12-31", "2027-01-01", false},
		{"inside an open-ended period", "2025-07-01", "", "2099-01-01", true},
		{"before an open-ended period", "2025-07-01", "", "2025-06-30", false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var to *model.ModelDate
			if testCase.to != "" {
				to = date(t, testCase.to)
			}
			assert.Equal(t, testCase.want,
				PeriodContains(date(t, testCase.from), to, testCase.taxDate))
		})
	}
}

// A malformed row must not apply to every date: treating a missing lower bound as "since forever"
// would make a half-entered rate apply to historical transactions.
func TestPeriodContainsRejectsMissingBounds(t *testing.T) {
	assert.False(t, PeriodContains(nil, nil, "2025-09-15"))
	assert.False(t, PeriodContains(date(t, "2025-07-01"), nil, ""))
}

// Two published versions may not overlap, so a tax date resolves to one rate.
func TestPeriodsOverlap(t *testing.T) {
	cases := []struct {
		name  string
		fromA string
		toA   string
		fromB string
		toB   string
		want  bool
	}{
		{"disjoint, A before B", "2025-01-01", "2025-06-30", "2025-07-01", "2025-12-31", false},
		{"disjoint, B before A", "2025-07-01", "2025-12-31", "2025-01-01", "2025-06-30", false},
		{"abutting by one day", "2025-01-01", "2025-07-01", "2025-07-01", "2025-12-31", true},
		{"fully contained", "2025-01-01", "2025-12-31", "2025-06-01", "2025-06-30", true},
		{"partial overlap", "2025-01-01", "2025-08-31", "2025-06-01", "2025-12-31", true},
		{"identical", "2025-01-01", "2025-12-31", "2025-01-01", "2025-12-31", true},
		{"two open-ended always overlap", "2025-01-01", "", "2030-01-01", "", true},
		{"open-ended A reaches a later B", "2025-01-01", "", "2030-01-01", "2030-12-31", true},
		{"closed A ends before open-ended B", "2025-01-01", "2025-12-31", "2026-01-01", "", false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var toA, toB *model.ModelDate
			if testCase.toA != "" {
				toA = date(t, testCase.toA)
			}
			if testCase.toB != "" {
				toB = date(t, testCase.toB)
			}
			assert.Equal(t, testCase.want, PeriodsOverlap(
				date(t, testCase.fromA), toA, date(t, testCase.fromB), toB))

			// Overlap is symmetric; a rule holding only in the caller's argument order would let
			// one of the two orderings through.
			assert.Equal(t, testCase.want, PeriodsOverlap(
				date(t, testCase.fromB), toB, date(t, testCase.fromA), toA))
		})
	}
}

func TestPeriodIsWellFormed(t *testing.T) {
	assert.True(t, PeriodIsWellFormed(date(t, "2025-01-01"), date(t, "2025-12-31")))
	assert.True(t, PeriodIsWellFormed(date(t, "2025-01-01"), date(t, "2025-01-01")))
	assert.True(t, PeriodIsWellFormed(date(t, "2025-01-01"), nil))
	assert.False(t, PeriodIsWellFormed(date(t, "2025-12-31"), date(t, "2025-01-01")))
	assert.False(t, PeriodIsWellFormed(nil, date(t, "2025-12-31")))
}

// Vietnam's 2% VAT reduction must stop applying after 2026-12-31 with no deployment.
func TestVietnamVatReductionExpiresOnItsOwn(t *testing.T) {
	from, to := date(t, "2025-07-01"), date(t, "2026-12-31")

	assert.True(t, PeriodContains(from, to, "2026-12-31"), "still in force on the last day")
	assert.False(t, PeriodContains(from, to, "2027-01-01"), "lapsed the next day, with no code change")
}
