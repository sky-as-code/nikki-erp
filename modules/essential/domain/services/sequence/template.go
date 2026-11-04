package sequence

import (
	"fmt"
	"strings"
	"time"
)

const (
	FormatOwnerSequence = "sequence"
	FormatOwnerConsumer = "consumer"

	varPeriod     = "period"
	prefixRange   = "range_"
	prefixCurrent = "current_"
)

// TemplateVars are the values a template may read (BR §5.1). Unprefixed variables read EffectiveAt, `current_`
// ones AcceptedAt and `range_` ones RangeStart (the local start of the period; zero for reset "never").
// Times should already be in the definition timezone; when Location is set they are converted to it here.
type TemplateVars struct {
	EffectiveAt time.Time
	AcceptedAt  time.Time
	RangeStart  time.Time
	PeriodKey   string
	Location    *time.Location
}

var braceAliases = map[string]string{
	"YYYY": "year", "YY": "y", "MM": "month", "DD": "day", "PERIOD": varPeriod,
}

var dateTimeVars = map[string]func(time.Time) string{
	"year":    func(t time.Time) string { return fmt.Sprintf("%04d", t.Year()) },
	"y":       func(t time.Time) string { return fmt.Sprintf("%02d", t.Year()%100) },
	"month":   func(t time.Time) string { return fmt.Sprintf("%02d", int(t.Month())) },
	"day":     func(t time.Time) string { return fmt.Sprintf("%02d", t.Day()) },
	"doy":     func(t time.Time) string { return fmt.Sprintf("%03d", t.YearDay()) },
	"woy":     func(t time.Time) string { return fmt.Sprintf("%02d", mondayWeekOfYear(t)) },
	"weekday": func(t time.Time) string { return fmt.Sprintf("%d", int(t.Weekday())) },
	"h24":     func(t time.Time) string { return fmt.Sprintf("%02d", t.Hour()) },
	"h12":     func(t time.Time) string { return fmt.Sprintf("%02d", hour12(t)) },
	"min":     func(t time.Time) string { return fmt.Sprintf("%02d", t.Minute()) },
	"sec":     func(t time.Time) string { return fmt.Sprintf("%02d", t.Second()) },
	"isoyear": func(t time.Time) string { year, _ := t.ISOWeek(); return fmt.Sprintf("%04d", year) },
	"isoy":    func(t time.Time) string { year, _ := t.ISOWeek(); return fmt.Sprintf("%02d", year%100) },
	"isoweek": func(t time.Time) string { _, week := t.ISOWeek(); return fmt.Sprintf("%02d", week) },
}

// mondayWeekOfYear is week 00-53 with weeks starting on Monday; days before the first Monday are week 00.
func mondayWeekOfYear(t time.Time) int {
	mondayBased := (int(t.Weekday()) + 6) % 7
	return (t.YearDay() - 1 + 7 - mondayBased) / 7
}

func hour12(t time.Time) int {
	hour := t.Hour() % 12
	if hour == 0 {
		return 12
	}
	return hour
}

// templateToken is either a literal (variable == "") or a canonical variable name such as "range_year".
type templateToken struct {
	literal  string
	variable string
}

func parseTemplate(tpl string) ([]templateToken, error) {
	tokens := []templateToken{}
	var literal strings.Builder
	for pos := 0; pos < len(tpl); {
		escaped, variable, next, err := scanToken(tpl, pos)
		if err != nil {
			return nil, err
		}
		pos = next
		if variable == "" {
			literal.WriteString(escaped)
			continue
		}
		if literal.Len() > 0 {
			tokens = append(tokens, templateToken{literal: literal.String()})
			literal.Reset()
		}
		tokens = append(tokens, templateToken{variable: variable})
	}
	if literal.Len() > 0 {
		tokens = append(tokens, templateToken{literal: literal.String()})
	}
	return tokens, nil
}

// scanToken reads one element at pos: an escape or plain byte (returned as literal text) or a placeholder.
func scanToken(tpl string, pos int) (string, string, int, error) {
	rest := tpl[pos:]
	switch {
	case strings.HasPrefix(rest, "%%"):
		return "%", "", pos + 2, nil
	case strings.HasPrefix(rest, "{{"):
		return "{", "", pos + 2, nil
	case strings.HasPrefix(rest, "}}"):
		return "}", "", pos + 2, nil
	case rest[0] == '%':
		return scanPercent(tpl, pos)
	case rest[0] == '{':
		return scanBrace(tpl, pos)
	case rest[0] == '}':
		return "", "", 0, templateErrorf("unescaped '}' at position %d (use '}}')", pos)
	default:
		return rest[:1], "", pos + 1, nil
	}
}

func scanPercent(tpl string, pos int) (string, string, int, error) {
	rest := tpl[pos:]
	closing := strings.Index(rest, ")s")
	if !strings.HasPrefix(rest, "%(") || closing < 0 {
		return "", "", 0, templateErrorf("bad placeholder at position %d (expected %%(name)s or %%%%)", pos)
	}
	name := rest[2:closing]
	if !isKnownVariable(name) {
		return "", "", 0, templateErrorf("unknown variable %q", name)
	}
	return "", name, pos + closing + 2, nil
}

func scanBrace(tpl string, pos int) (string, string, int, error) {
	rest := tpl[pos:]
	closing := strings.IndexByte(rest, '}')
	if closing < 0 {
		return "", "", 0, templateErrorf("unclosed '{' at position %d (use '{{')", pos)
	}
	alias := rest[1:closing]
	name, found := braceAliases[alias]
	if !found {
		return "", "", 0, templateErrorf("unknown alias {%s}", alias)
	}
	return "", name, pos + closing + 1, nil
}

func isKnownVariable(name string) bool {
	if name == varPeriod {
		return true
	}
	_, base := splitVariable(name)
	_, found := dateTimeVars[base]
	return found
}

type variableSource int

const (
	sourceEffectiveAt variableSource = iota
	sourceRangeStart
	sourceAcceptedAt
)

// splitVariable strips at most ONE source prefix, so validation and rendering agree on the base:
// range_current_year is base "current_year", which is no date/time variable and is rejected.
func splitVariable(name string) (variableSource, string) {
	if base, found := strings.CutPrefix(name, prefixRange); found {
		return sourceRangeStart, base
	}
	if base, found := strings.CutPrefix(name, prefixCurrent); found {
		return sourceAcceptedAt, base
	}
	return sourceEffectiveAt, name
}

// RenderTemplate renders a prefix or suffix template. Unknown variables, bad syntax and missing values fail with
// a *TemplateError (INVALID_TEMPLATE).
func RenderTemplate(tpl string, vars TemplateVars) (string, error) {
	tokens, err := parseTemplate(tpl)
	if err != nil {
		return "", err
	}
	var output strings.Builder
	for _, token := range tokens {
		if token.variable == "" {
			output.WriteString(token.literal)
			continue
		}
		value, err := resolveVariable(token.variable, vars)
		if err != nil {
			return "", err
		}
		output.WriteString(value)
	}
	return output.String(), nil
}

func resolveVariable(name string, vars TemplateVars) (string, error) {
	if name == varPeriod {
		if vars.PeriodKey == "" {
			return "", templateErrorf("period key is missing")
		}
		return vars.PeriodKey, nil
	}
	kind, base := splitVariable(name)
	format, found := dateTimeVars[base]
	if !found {
		return "", templateErrorf("unknown variable %q", name)
	}
	source, label := vars.EffectiveAt, "effective_at"
	switch kind {
	case sourceRangeStart:
		source, label = vars.RangeStart, "range start"
	case sourceAcceptedAt:
		source, label = vars.AcceptedAt, "accepted_at"
	}
	if source.IsZero() {
		return "", templateErrorf("variable %q needs %s, which is not set", name, label)
	}
	if vars.Location != nil {
		source = source.In(vars.Location)
	}
	return format(source), nil
}
