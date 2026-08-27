package cronexpr

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustParse(t *testing.T, expr string) *CronExpr {
	t.Helper()
	parsed, err := Parse(expr)
	require.NoError(t, err, "expression %q should parse", expr)
	return parsed
}

func utc(year int, month time.Month, day, hour, minute int) time.Time {
	return time.Date(year, month, day, hour, minute, 0, 0, time.UTC)
}

func TestParseRejectsMalformedExpressions(t *testing.T) {
	tests := []struct {
		name      string
		expr      string
		wantField string
	}{
		{"six fields", "*/5 * * * * *", ""},
		{"four fields", "* * * *", ""},
		{"empty", "", ""},
		{"macro", "@daily", ""},
		{"minute out of range", "60 * * * *", "minute"},
		{"hour out of range", "* 24 * * *", "hour"},
		{"day of month zero", "* * 0 * *", "day_of_month"},
		{"month out of range", "* * * 13 *", "month"},
		{"day of week out of range", "* * * * 8", "day_of_week"},
		{"inverted range", "10-5 * * * *", "minute"},
		{"non numeric", "abc * * * *", "minute"},
		{"zero step", "*/0 * * * *", "minute"},
		{"negative step", "*/-1 * * * *", "minute"},
		{"step on bare value", "5/10 * * * *", "minute"},
		{"quartz question mark", "0 0 * * ?", "day_of_week"},
		{"last day token", "0 0 L * *", "day_of_month"},
		{"weekday token", "0 0 W * *", "day_of_month"},
		{"nth weekday token", "0 0 * * 5#2", "day_of_week"},
		{"month name", "0 0 * JAN *", "month"},
		{"weekday name", "0 0 * * MON", "day_of_week"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse(tc.expr)
			require.Error(t, err, "expression %q must be rejected", tc.expr)

			parseErr, ok := err.(*ParseError)
			require.True(t, ok, "error must be a *ParseError so REST can scope it to a field")
			assert.Equal(t, tc.wantField, parseErr.Field)
		})
	}
}

func TestParseAcceptsEverySupportedSyntaxForm(t *testing.T) {
	tests := []string{
		"* * * * *",
		"5 * * * *",
		"1,5,9 * * * *",
		"1-5 * * * *",
		"*/5 * * * *",
		"10-40/5 * * * *",
		"0,10-20/5,59 * * * *",
		"0 0 1 1 0",
		"59 23 31 12 6",
	}

	for _, expr := range tests {
		t.Run(expr, func(t *testing.T) {
			_, err := Parse(expr)
			assert.NoError(t, err)
		})
	}
}

func TestNextTruncatesToTheMinuteAndIsStrictlyAfter(t *testing.T) {
	expr := mustParse(t, "*/5 * * * *")

	next, ok := expr.Next(time.Date(2026, time.August, 20, 10, 0, 30, 0, time.UTC))

	require.True(t, ok)
	assert.Equal(t, utc(2026, time.August, 20, 10, 5), next)
	assert.Equal(t, 0, next.Second(), "an occurrence must never carry sub-minute precision")
	assert.Equal(t, 0, next.Nanosecond())
}

// This is the case the retry window depends on. next_occurrence_at is computed as
// Next(scheduled_for); if Next returned the instant it was handed, the window would be
// zero-width and every retry would be cancelled.
func TestNextOnAnExactOccurrenceReturnsTheFollowingOne(t *testing.T) {
	expr := mustParse(t, "*/5 * * * *")

	next, ok := expr.Next(utc(2026, time.August, 20, 10, 5))

	require.True(t, ok)
	assert.Equal(t, utc(2026, time.August, 20, 10, 10), next,
		"Next must be strictly after its argument, never equal to it")
}

func TestNextAdvancesAcrossHourDayMonthAndYearBoundaries(t *testing.T) {
	tests := []struct {
		name string
		expr string
		from time.Time
		want time.Time
	}{
		{
			name: "next hour",
			expr: "0 * * * *",
			from: utc(2026, time.August, 20, 10, 30),
			want: utc(2026, time.August, 20, 11, 0),
		},
		{
			name: "next day",
			expr: "0 3 * * *",
			from: utc(2026, time.August, 20, 10, 0),
			want: utc(2026, time.August, 21, 3, 0),
		},
		{
			name: "next month",
			expr: "0 0 1 * *",
			from: utc(2026, time.August, 20, 10, 0),
			want: utc(2026, time.September, 1, 0, 0),
		},
		{
			name: "next year",
			expr: "0 0 1 1 *",
			from: utc(2026, time.August, 20, 10, 0),
			want: utc(2027, time.January, 1, 0, 0),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			next, ok := mustParse(t, tc.expr).Next(tc.from)
			require.True(t, ok)
			assert.Equal(t, tc.want, next)
		})
	}
}

// Both day fields restricted means OR, not AND. Getting this backwards is the classic cron
// bug: the schedule would fire far less often than the operator intended, and only on the
// rare days where both happen to coincide.
func TestDayOfMonthAndDayOfWeekCombineWithOr(t *testing.T) {
	// The 13th of any month, and additionally every Friday.
	expr := mustParse(t, "0 0 13 * 5")

	// 2026-08-13 is a Thursday: matches on day-of-month alone.
	next, ok := expr.Next(utc(2026, time.August, 12, 0, 0))
	require.True(t, ok)
	assert.Equal(t, utc(2026, time.August, 13, 0, 0), next)

	// 2026-08-14 is a Friday: matches on day-of-week alone.
	next, ok = expr.Next(utc(2026, time.August, 13, 0, 1))
	require.True(t, ok)
	assert.Equal(t, utc(2026, time.August, 14, 0, 0), next)
}

func TestOnlyOneRestrictedDayFieldIsConsultedAlone(t *testing.T) {
	// Day-of-week restricted, day-of-month is "*": every Monday only.
	expr := mustParse(t, "0 0 * * 1")

	next, ok := expr.Next(utc(2026, time.August, 20, 0, 0)) // Thursday
	require.True(t, ok)
	assert.Equal(t, time.Monday, next.Weekday())
	assert.Equal(t, utc(2026, time.August, 24, 0, 0), next)
}

func TestDayOfWeekSevenIsSunday(t *testing.T) {
	withSeven := mustParse(t, "0 0 * * 7")
	withZero := mustParse(t, "0 0 * * 0")

	from := utc(2026, time.August, 20, 0, 0)
	nextSeven, okSeven := withSeven.Next(from)
	nextZero, okZero := withZero.Next(from)

	require.True(t, okSeven, "day-of-week 7 must be accepted, not silently never fire")
	require.True(t, okZero)
	assert.Equal(t, nextZero, nextSeven)
	assert.Equal(t, time.Sunday, nextSeven.Weekday())
}

func TestNextHandlesLeapDay(t *testing.T) {
	expr := mustParse(t, "0 0 29 2 *")

	next, ok := expr.Next(utc(2026, time.March, 1, 0, 0))

	require.True(t, ok)
	assert.Equal(t, utc(2028, time.February, 29, 0, 0), next,
		"the next 29 February after March 2026 is in the 2028 leap year")
}

// A valid expression that can never occur must terminate, not spin. This is what the search
// horizon exists for.
func TestNextReportsNoOccurrenceForAnImpossibleExpression(t *testing.T) {
	expr := mustParse(t, "0 0 30 2 *") // 30 February

	next, ok := expr.Next(utc(2026, time.August, 20, 10, 0))

	assert.False(t, ok, "an impossible expression has no next occurrence")
	assert.True(t, next.IsZero(), "the zero time signals absence; a sentinel would be usable by accident")
}

// AC-2: cron is evaluated in UTC. Across a boundary where a local timezone would shift, the
// occurrences stay evenly spaced, which is the observable consequence of never consulting a
// local zone.
func TestOccurrencesAreEvenlySpacedAcrossDstBoundaries(t *testing.T) {
	expr := mustParse(t, "0 * * * *")

	// 2026-03-29 01:00 UTC is inside the window where EU local time jumps forward.
	current := utc(2026, time.March, 29, 0, 0)
	for i := 0; i < 5; i++ {
		next, ok := expr.Next(current)
		require.True(t, ok)
		assert.Equal(t, time.Hour, next.Sub(current),
			"UTC evaluation must not gain or lose an hour at a DST boundary")
		assert.Equal(t, time.UTC, next.Location())
		current = next
	}
}

func TestNextAlwaysReturnsUtc(t *testing.T) {
	expr := mustParse(t, "*/5 * * * *")
	jakarta := time.FixedZone("WIB", 7*60*60)

	next, ok := expr.Next(time.Date(2026, time.August, 20, 17, 3, 0, 0, jakarta))

	require.True(t, ok)
	assert.Equal(t, time.UTC, next.Location(),
		"an input in another zone must not leak that zone into the result")
	assert.Equal(t, utc(2026, time.August, 20, 10, 5), next)
}

func TestStringReturnsNormalizedExpression(t *testing.T) {
	expr := mustParse(t, "  */5   *  * *    * ")
	assert.Equal(t, "*/5 * * * *", expr.String())
}
