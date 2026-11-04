package sequence

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderVars() TemplateVars {
	return TemplateVars{
		// Sunday 2026-01-04 15:07:09: the ISO year is 2026 week 01, Monday-based week 00.
		EffectiveAt: time.Date(2026, time.January, 4, 15, 7, 9, 0, time.UTC),
		AcceptedAt:  time.Date(2027, time.January, 1, 0, 30, 0, 0, time.UTC),
		RangeStart:  time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
		PeriodKey:   "2026",
	}
}

func TestRenderTemplate(t *testing.T) {
	cases := map[string]string{
		"":                                  "",
		"SO/":                               "SO/",
		"SO/{YYYY}/":                        "SO/2026/",
		"{YY}{MM}{DD}-":                     "260104-",
		"%(year)s-%(y)s":                    "2026-26",
		"%(month)s/%(day)s":                 "01/04",
		"%(doy)s %(woy)s":                   "004 00",
		"%(weekday)s":                       "0",
		"%(h24)s:%(min)s:%(sec)s":           "15:07:09",
		"%(h12)s":                           "03",
		"%(isoyear)s-%(isoy)s-W%(isoweek)s": "2026-26-W01",
		"%(range_year)s%(range_month)s%(range_day)s": "20260101",
		"%(current_year)s/%(current_h12)s":           "2027/12",
		"%(period)s|{PERIOD}":                        "2026|2026",
		"100%% {{ok}} {{YYYY}}":                      "100% {ok} {YYYY}",
		"é/{MM}/":                                    "é/01/",
	}
	for tpl, want := range cases {
		got, err := RenderTemplate(tpl, renderVars())
		require.NoError(t, err, tpl)
		assert.Equal(t, want, got, tpl)
	}
}

func TestRenderTemplateWeekOfYear(t *testing.T) {
	cases := map[string]string{"2026-01-05": "01", "2024-01-01": "01", "2023-01-01": "00", "2026-12-31": "52"}
	for day, want := range cases {
		parsed, err := time.Parse("2006-01-02", day)
		require.NoError(t, err)
		got, err := RenderTemplate("%(woy)s", TemplateVars{EffectiveAt: parsed})
		require.NoError(t, err)
		assert.Equal(t, want, got, day)
	}
}

func TestRenderTemplateLocation(t *testing.T) {
	location, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	require.NoError(t, err)
	vars := TemplateVars{EffectiveAt: time.Date(2026, 11, 3, 17, 0, 0, 0, time.UTC), Location: location}
	got, err := RenderTemplate("{YYYY}{MM}{DD}%(h24)s", vars)
	require.NoError(t, err)
	assert.Equal(t, "2026110400", got)
}

// AC08: unknown variables and bad escapes are client errors.
func TestRenderTemplateErrors(t *testing.T) {
	templates := []string{
		"SO/{WH}/", "{ORG}", "{yyyy}", "{YYYY", "}", "a}b", "%", "50%", "%(year)", "%(year)d", "%(unknown)s",
		"%(range_period)s", "%(current_)s", "%()s", "{}",
	}
	for _, tpl := range templates {
		_, err := RenderTemplate(tpl, renderVars())
		require.Error(t, err, tpl)
		assert.ErrorIs(t, err, ErrInvalidTemplate, tpl)
		assert.Contains(t, err.Error(), CodeInvalidTemplate, tpl)
	}
}

func TestRenderTemplateMissingValues(t *testing.T) {
	vars := renderVars()
	vars.RangeStart = time.Time{}
	_, err := RenderTemplate("%(range_year)s", vars)
	assert.ErrorIs(t, err, ErrInvalidTemplate)

	vars = renderVars()
	vars.PeriodKey = ""
	_, err = RenderTemplate("{PERIOD}", vars)
	assert.ErrorIs(t, err, ErrInvalidTemplate)

	_, err = RenderTemplate("%(current_year)s", TemplateVars{EffectiveAt: time.Now()})
	assert.ErrorIs(t, err, ErrInvalidTemplate)
}

// AC44 / SEQ-UNIQ-02: the number must have a non-digit boundary on both sides.
func TestValidateTemplateBoundary(t *testing.T) {
	cases := []struct {
		tpl, owner, reset string
		isPrefix, valid   bool
	}{
		{"", FormatOwnerSequence, ResetNever, true, true},
		{"", FormatOwnerSequence, ResetNever, false, true},
		{"SO/{YYYY}/", FormatOwnerSequence, ResetYearly, true, true},
		{"SO{YYYY}", FormatOwnerSequence, ResetYearly, true, false},
		{"{YYYY}", FormatOwnerSequence, ResetYearly, true, false},
		{"INV2", FormatOwnerSequence, ResetNever, true, false},
		{"{PERIOD}-", FormatOwnerSequence, ResetDaily, true, true},
		{"{PERIOD}", FormatOwnerSequence, ResetNever, true, true},
		{"{PERIOD}", FormatOwnerSequence, ResetDaily, true, false},
		{"/{YYYY}", FormatOwnerSequence, ResetNever, false, true},
		{"{YYYY}", FormatOwnerSequence, ResetNever, false, false},
		{"7X", FormatOwnerSequence, ResetNever, false, false},
		{"{YYYY}", FormatOwnerConsumer, ResetYearly, true, true},
		{"%(range_year)s/", FormatOwnerSequence, ResetYearly, true, true},
		{"%(range_year)s/", FormatOwnerSequence, ResetNever, true, false},
		{"%(range_year)s/", FormatOwnerConsumer, ResetNever, true, false},
		{"SO/{WH}/", FormatOwnerSequence, ResetNever, true, false},
		{"SO/{WH}/", FormatOwnerConsumer, ResetNever, true, false},
	}
	for _, tc := range cases {
		err := ValidateTemplate(tc.tpl, tc.owner, tc.reset, tc.isPrefix)
		if tc.valid {
			assert.NoError(t, err, "%+v", tc)
		} else {
			assert.ErrorIs(t, err, ErrInvalidTemplate, "%+v", tc)
		}
	}
}

func TestValidatePeriodMarker(t *testing.T) {
	cases := []struct {
		prefix, suffix, reset string
		valid                 bool
	}{
		{"SO/", "", ResetNever, true},
		{"SO/{YYYY}/", "", ResetYearly, true},
		{"SO/", "/{YYYY}", ResetYearly, true},
		{"SO/{YY}/", "", ResetYearly, false},
		{"SO/%(isoyear)s/", "", ResetYearly, false},
		{"SO/%(current_year)s/", "", ResetYearly, false},
		{"SO/%(range_year)s/", "", ResetYearly, true},
		{"SO/{YYYY}/", "", ResetMonthly, false},
		{"SO/{YYYY}{MM}/", "", ResetMonthly, true},
		{"SO/%(range_year)s/", "/%(range_month)s", ResetMonthly, true},
		{"SO/{YYYY}{MM}/", "", ResetDaily, false},
		{"SO/{YYYY}{MM}{DD}/", "", ResetDaily, true},
		{"SO/{PERIOD}/", "", ResetDaily, true},
		{"", "/%(period)s", ResetMonthly, true},
		{"SO/{WH}/", "", ResetDaily, false},
		{"SO/", "", "weekly", false},
	}
	for _, tc := range cases {
		err := ValidatePeriodMarker(tc.prefix, tc.suffix, tc.reset)
		if tc.valid {
			assert.NoError(t, err, "%+v", tc)
		} else {
			assert.ErrorIs(t, err, ErrInvalidTemplate, "%+v", tc)
		}
	}
}

// A variable carries at most one source prefix. Stacked or bare prefixes are rejected by validation
// (for every format owner) and by rendering, never resolved to a missing formatter.
func TestTemplateRejectsStackedOrBarePrefixes(t *testing.T) {
	cases := []string{
		"%(range_current_year)s", "%(current_range_year)s", "%(range_range_year)s",
		"%(current_current_year)s", "%(range_)s", "%(current_)s", "%(range_period)s", "%(current_period)s",
	}
	for _, tpl := range cases {
		assert.False(t, isKnownVariable(tpl[2:len(tpl)-2]), tpl)
		for _, owner := range []string{FormatOwnerSequence, FormatOwnerConsumer} {
			err := ValidateTemplate(tpl, owner, ResetYearly, true)
			assert.ErrorIs(t, err, ErrInvalidTemplate, "%s owner=%s", tpl, owner)
		}
		assert.NotPanics(t, func() {
			_, err := RenderTemplate(tpl, renderVars())
			assert.ErrorIs(t, err, ErrInvalidTemplate, tpl)
		}, tpl)
	}
}

func TestResolveVariableNeverPanicsOnUnknownBase(t *testing.T) {
	for _, name := range []string{"range_current_year", "current_range_year", "range_", "current_", "nope"} {
		assert.NotPanics(t, func() {
			_, err := resolveVariable(name, renderVars())
			assert.ErrorIs(t, err, ErrInvalidTemplate, name)
		}, name)
	}
}
