package sequence

import (
	"strings"
	"time"

	"go.bryk.io/pkg/errors"
)

// ValidateTemplate checks one prefix (isPrefix) or suffix template: syntax, known variables, `range_` only with a
// reset period, and, when Sequence formats the whole value, the digit boundary of SEQ-UNIQ-02 (a rendered prefix
// must not end with a digit, a rendered suffix must not start with one). The period-marker rule (SEQ-TIME-05)
// spans prefix and suffix together and lives in ValidatePeriodMarker.
func ValidateTemplate(tpl string, formatOwner string, resetPeriod string, isPrefix bool) error {
	tokens, err := parseTemplate(tpl)
	if err != nil {
		return err
	}
	if resetPeriod == ResetNever {
		for _, token := range tokens {
			if strings.HasPrefix(token.variable, prefixRange) {
				return templateErrorf("variable %q needs a reset period other than never", token.variable)
			}
		}
	}
	if formatOwner == FormatOwnerConsumer {
		return nil
	}
	vars, err := sampleVars(resetPeriod)
	if err != nil {
		return err
	}
	rendered, err := RenderTemplate(tpl, vars)
	if err != nil {
		return err
	}
	if rendered == "" {
		return nil
	}
	if isPrefix && isDigit(rendered[len(rendered)-1]) {
		return templateErrorf("prefix must be empty or end with a non-digit character")
	}
	if !isPrefix && isDigit(rendered[0]) {
		return templateErrorf("suffix must be empty or start with a non-digit character")
	}
	return nil
}

func isDigit(char byte) bool {
	return char >= '0' && char <= '9'
}

// sampleVars gives every variable a value so a template can be rendered without a real issuance.
func sampleVars(resetPeriod string) (TemplateVars, error) {
	sample := time.Date(2026, time.November, 3, 14, 5, 9, 0, time.UTC)
	period, err := PeriodOf(sample, "UTC", resetPeriod)
	if err != nil {
		return TemplateVars{}, errors.Wrap(err, "sequence: validate template")
	}
	rangeStart := period.Start
	if rangeStart.IsZero() {
		rangeStart = sample
	}
	return TemplateVars{EffectiveAt: sample, AcceptedAt: sample, RangeStart: rangeStart, PeriodKey: period.Key}, nil
}

// periodMarkers lists, per reset period, the date parts a format must show to tell its periods apart (SEQ-TIME-05).
var periodMarkers = map[string][]string{
	ResetYearly:  {"year"},
	ResetMonthly: {"year", "month"},
	ResetDaily:   {"year", "month", "day"},
}

// ValidatePeriodMarker enforces SEQ-TIME-05 on the combined prefix and suffix of a definition whose value Sequence
// formats (format_owner=sequence): with a reset period, the format must show `period` or every date part of the
// period at full width (4-digit year, then month, then day). Parts may come from effective_at (unprefixed, e.g.
// {YYYY}) or the period start (`range_`), both of which lie inside the period; `current_` parts do not count, nor
// do 2-digit years or ISO weeks.
func ValidatePeriodMarker(prefix, suffix, resetPeriod string) error {
	if resetPeriod == ResetNever {
		return nil
	}
	required, found := periodMarkers[resetPeriod]
	if !found {
		return templateErrorf("invalid reset period %q", resetPeriod)
	}
	present := map[string]bool{}
	for _, tpl := range []string{prefix, suffix} {
		tokens, err := parseTemplate(tpl)
		if err != nil {
			return err
		}
		for _, token := range tokens {
			present[strings.TrimPrefix(token.variable, prefixRange)] = true
		}
	}
	if present[varPeriod] {
		return nil
	}
	for _, part := range required {
		if !present[part] {
			return templateErrorf("reset %s needs %%(period)s / {PERIOD} or the full date parts in the format",
				resetPeriod)
		}
	}
	return nil
}
