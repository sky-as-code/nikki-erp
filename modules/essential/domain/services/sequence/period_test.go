package sequence

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func utc(text string) time.Time {
	parsed, err := time.Parse(time.RFC3339, text)
	if err != nil {
		panic(err)
	}
	return parsed
}

func TestPeriodOfKeys(t *testing.T) {
	cases := []struct {
		instant, timezone, reset, wantKey, wantStart, wantNext string
	}{
		{"2026-11-03T10:00:00Z", "UTC", ResetNever, PeriodKeyNone, "", ""},
		{"2026-11-03T10:00:00Z", "UTC", ResetYearly, "2026", "2026-01-01T00:00:00Z", "2027-01-01T00:00:00Z"},
		{"2026-11-03T10:00:00Z", "UTC", ResetMonthly, "2026-11", "2026-11-01T00:00:00Z", "2026-12-01T00:00:00Z"},
		{"2026-12-15T10:00:00Z", "UTC", ResetMonthly, "2026-12", "2026-12-01T00:00:00Z", "2027-01-01T00:00:00Z"},
		{"2026-11-03T10:00:00Z", "UTC", ResetDaily, "2026-11-03", "2026-11-03T00:00:00Z", "2026-11-04T00:00:00Z"},
		{"2026-02-28T10:00:00Z", "", ResetDaily, "2026-02-28", "2026-02-28T00:00:00Z", "2026-03-01T00:00:00Z"},
		// Asia/Ho_Chi_Minh is UTC+7: 17:00 UTC is local midnight.
		{"2026-11-03T16:59:59Z", "Asia/Ho_Chi_Minh", ResetDaily, "2026-11-03",
			"2026-11-02T17:00:00Z", "2026-11-03T17:00:00Z"},
		{"2026-11-03T17:00:00Z", "Asia/Ho_Chi_Minh", ResetDaily, "2026-11-04",
			"2026-11-03T17:00:00Z", "2026-11-04T17:00:00Z"},
		{"2026-11-30T17:00:00Z", "Asia/Ho_Chi_Minh", ResetMonthly, "2026-12",
			"2026-11-30T17:00:00Z", "2026-12-31T17:00:00Z"},
		{"2026-12-31T17:00:00Z", "Asia/Ho_Chi_Minh", ResetYearly, "2027",
			"2026-12-31T17:00:00Z", "2027-12-31T17:00:00Z"},
		// New York spring forward (2026-03-08, a 23 h day) and fall back (2026-11-01, a 25 h day).
		{"2026-03-08T12:00:00Z", "America/New_York", ResetDaily, "2026-03-08",
			"2026-03-08T05:00:00Z", "2026-03-09T04:00:00Z"},
		{"2026-03-09T04:30:00Z", "America/New_York", ResetDaily, "2026-03-09",
			"2026-03-09T04:00:00Z", "2026-03-10T04:00:00Z"},
		{"2026-11-01T12:00:00Z", "America/New_York", ResetDaily, "2026-11-01",
			"2026-11-01T04:00:00Z", "2026-11-02T05:00:00Z"},
		{"2026-11-02T04:30:00Z", "America/New_York", ResetDaily, "2026-11-01",
			"2026-11-01T04:00:00Z", "2026-11-02T05:00:00Z"},
		{"2026-03-15T12:00:00Z", "America/New_York", ResetMonthly, "2026-03",
			"2026-03-01T05:00:00Z", "2026-04-01T04:00:00Z"},
	}
	for _, tc := range cases {
		result, err := PeriodOf(utc(tc.instant), tc.timezone, tc.reset)
		require.NoError(t, err, tc.instant)
		assert.Equal(t, tc.wantKey, result.Key, tc.instant)
		if tc.wantStart == "" {
			assert.True(t, result.Start.IsZero())
			assert.True(t, result.NextStart.IsZero())
			continue
		}
		assert.True(t, utc(tc.wantStart).Equal(result.Start), "%s start %s", tc.instant, result.Start)
		assert.True(t, utc(tc.wantNext).Equal(result.NextStart), "%s next %s", tc.instant, result.NextStart)
	}
}

func TestPeriodOfDstDayLengths(t *testing.T) {
	spring, err := PeriodOf(utc("2026-03-08T12:00:00Z"), "America/New_York", ResetDaily)
	require.NoError(t, err)
	assert.Equal(t, 23*time.Hour, spring.NextStart.Sub(spring.Start))
	fall, err := PeriodOf(utc("2026-11-01T12:00:00Z"), "America/New_York", ResetDaily)
	require.NoError(t, err)
	assert.Equal(t, 25*time.Hour, fall.NextStart.Sub(fall.Start))
}

func TestPeriodOfErrors(t *testing.T) {
	_, err := PeriodOf(time.Now(), "Mars/Olympus", ResetDaily)
	assert.Error(t, err)
	_, err = PeriodOf(time.Now(), "UTC", "weekly")
	assert.Error(t, err)
}
