package cronexpr

import (
	"strconv"
	"strings"
)

// fieldSpec describes one of the five cron fields: its name for error messages and the
// inclusive range of values it accepts.
type fieldSpec struct {
	name string
	min  int
	max  int
}

var (
	minuteSpec     = fieldSpec{name: "minute", min: 0, max: 59}
	hourSpec       = fieldSpec{name: "hour", min: 0, max: 23}
	dayOfMonthSpec = fieldSpec{name: "day_of_month", min: 1, max: 31}
	monthSpec      = fieldSpec{name: "month", min: 1, max: 12}
	dayOfWeekSpec  = fieldSpec{name: "day_of_week", min: 0, max: 6}
)

// parseField compiles one cron field into a bitmask, where bit N is set when value N matches.
//
// The second return reports whether the field was restricted, meaning it was anything other
// than "*". Day-of-month and day-of-week need that fact preserved: their combination rule
// depends on which of them was literally "*", and a mask with every bit set is
// indistinguishable from "*" once it has been built.
func parseField(expr string, spec fieldSpec) (uint64, bool, error) {
	if expr == "" {
		return 0, false, newParseError(spec.name, expr, "field is empty")
	}
	if err := rejectUnsupported(expr, spec); err != nil {
		return 0, false, err
	}

	if expr == "*" {
		return fullMask(spec), false, nil
	}

	var mask uint64
	for _, part := range strings.Split(expr, ",") {
		partMask, err := parseFieldPart(part, spec)
		if err != nil {
			return 0, false, err
		}
		mask |= partMask
	}
	return mask, true, nil
}

// rejectUnsupported names the syntaxes this parser deliberately does not implement, so a
// caller who types a Quartz or extended-cron expression is told which token is the problem
// instead of getting a generic "invalid value" for a numeric parse that was never going to
// succeed.
func rejectUnsupported(expr string, spec fieldSpec) error {
	unsupported := map[string]string{
		"?": "'?' is not supported; use '*'",
		"L": "'L' (last) is not supported",
		"W": "'W' (weekday) is not supported",
		"#": "'#' (nth weekday) is not supported",
		"@": "macros such as @daily are not supported",
	}
	for token, msg := range unsupported {
		if strings.Contains(expr, token) {
			return newParseError(spec.name, expr, msg)
		}
	}
	return nil
}

// parseFieldPart compiles a single comma-separated element: a value, a range, or either of
// those with a step.
func parseFieldPart(part string, spec fieldSpec) (uint64, error) {
	rangeExpr, stepExpr, hasStep := strings.Cut(part, "/")

	step := 1
	if hasStep {
		parsed, err := strconv.Atoi(stepExpr)
		if err != nil {
			return 0, newParseError(spec.name, part, "step '"+stepExpr+"' is not a number")
		}
		if parsed <= 0 {
			return 0, newParseError(spec.name, part, "step must be greater than zero")
		}
		step = parsed
	}

	low, high, err := parseRange(rangeExpr, part, spec, hasStep)
	if err != nil {
		return 0, err
	}

	var mask uint64
	for val := low; val <= high; val += step {
		mask |= 1 << uint(val)
	}
	return mask, nil
}

// parseRange resolves the value-or-range half of a field part into inclusive bounds.
//
// A bare value with a step ("5/10") is rejected rather than guessed at: cron implementations
// disagree on whether it means "5 only" or "from 5 to the end of the range", and §5 does not
// require the form, so accepting it would bake in a guess the caller cannot see.
func parseRange(rangeExpr string, part string, spec fieldSpec, hasStep bool) (int, int, error) {
	if rangeExpr == "*" {
		return spec.min, spec.max, nil
	}

	lowExpr, highExpr, isRange := strings.Cut(rangeExpr, "-")

	low, err := parseValue(lowExpr, part, spec)
	if err != nil {
		return 0, 0, err
	}

	if !isRange {
		if hasStep {
			return 0, 0, newParseError(spec.name, part,
				"a step needs a range or '*' on its left, for example '"+lowExpr+"-"+strconv.Itoa(spec.max)+"/n'")
		}
		return low, low, nil
	}

	high, err := parseValue(highExpr, part, spec)
	if err != nil {
		return 0, 0, err
	}
	if low > high {
		return 0, 0, newParseError(spec.name, part, "range start is greater than its end")
	}
	return low, high, nil
}

// parseValue converts one number and bounds-checks it against the field.
//
// Day-of-week 7 is accepted and folded onto 0. Both denote Sunday across cron
// implementations, and rejecting 7 would turn a widely written expression into a schedule
// that silently never fires.
func parseValue(expr string, part string, spec fieldSpec) (int, error) {
	val, err := strconv.Atoi(strings.TrimSpace(expr))
	if err != nil {
		return 0, newParseError(spec.name, part, "'"+expr+"' is not a number")
	}

	if spec.name == dayOfWeekSpec.name && val == 7 {
		val = 0
	}

	if val < spec.min || val > spec.max {
		return 0, newParseError(spec.name, part,
			"value "+strconv.Itoa(val)+" is outside "+strconv.Itoa(spec.min)+"-"+strconv.Itoa(spec.max))
	}
	return val, nil
}

func fullMask(spec fieldSpec) uint64 {
	var mask uint64
	for val := spec.min; val <= spec.max; val++ {
		mask |= 1 << uint(val)
	}
	return mask
}

func maskHas(mask uint64, val int) bool {
	if val < 0 || val > 63 {
		return false
	}
	return mask&(1<<uint(val)) != 0
}
