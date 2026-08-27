// Package cronexpr parses and evaluates 5-field cron expressions in UTC.
//
// It exists because the scheduler must not depend on a third-party scheduling library, and
// because the evaluation rules it needs are narrow enough to state exactly: five fields, UTC
// only, one-minute resolution, and a Next that is total rather than potentially unbounded.
//
// The supported syntax per field is '*', an exact value, a comma-separated list, an inclusive
// range, a step on '*', and a step on a range. Month names, weekday names, macros such as
// @daily, and the Quartz extensions '?', 'L', 'W' and '#' are deliberately not supported and
// are rejected by name.
package cronexpr

import (
	"strconv"
	"strings"
	"time"
)

// searchHorizonYears bounds how far Next will look before giving up.
//
// The bound is what makes Next total. An expression such as "0 0 30 2 *" - the 30th of
// February - is syntactically valid and never occurs, and an unbounded search for its next
// occurrence would spin a goroutine forever. Five years clears any leap-year cycle, so the
// only expressions that reach the horizon are the ones with no occurrence at all.
const searchHorizonYears = 5

const fieldCount = 5

// CronExpr is a parsed 5-field cron expression, evaluated in UTC only.
//
// The five fields are minute, hour, day-of-month, month and day-of-week. There is no seconds
// field and no year field: the minimum resolution is one minute.
type CronExpr struct {
	minute     uint64
	hour       uint64
	dayOfMonth uint64
	month      uint64
	dayOfWeek  uint64

	// domRestricted and dowRestricted record whether each day field was written as something
	// other than "*". The day-matching rule depends on it - see dayMatches.
	domRestricted bool
	dowRestricted bool

	canonical string
}

// ParseError names the field that could not be parsed and why.
//
// It is a distinct type, rather than a wrapped generic error, so that a REST layer can attach
// the failure to the offending field and answer with a message a caller can act on.
type ParseError struct {
	// Field is the cron field name ("minute", "day_of_week", ...), or empty when the problem
	// is with the expression as a whole rather than one of its fields.
	Field string
	Expr  string
	Msg   string
}

func (this *ParseError) Error() string {
	if this.Field == "" {
		return "invalid cron expression '" + this.Expr + "': " + this.Msg
	}
	return "invalid cron expression: field " + this.Field + " ('" + this.Expr + "'): " + this.Msg
}

func newParseError(field string, expr string, msg string) *ParseError {
	return &ParseError{Field: field, Expr: expr, Msg: msg}
}

// Parse compiles a 5-field cron expression.
//
// It returns a *ParseError on every failure, so a caller can inspect Field to report which
// part of the expression was wrong.
func Parse(expr string) (*CronExpr, error) {
	fields := strings.Fields(expr)
	if len(fields) != fieldCount {
		return nil, newParseError("", expr,
			"expected 5 fields (minute hour day-of-month month day-of-week), got "+strconv.Itoa(len(fields)))
	}

	parsed := &CronExpr{canonical: strings.Join(fields, " ")}

	var err error
	if parsed.minute, _, err = parseField(fields[0], minuteSpec); err != nil {
		return nil, err
	}
	if parsed.hour, _, err = parseField(fields[1], hourSpec); err != nil {
		return nil, err
	}
	if parsed.dayOfMonth, parsed.domRestricted, err = parseField(fields[2], dayOfMonthSpec); err != nil {
		return nil, err
	}
	if parsed.month, _, err = parseField(fields[3], monthSpec); err != nil {
		return nil, err
	}
	if parsed.dayOfWeek, parsed.dowRestricted, err = parseField(fields[4], dayOfWeekSpec); err != nil {
		return nil, err
	}

	return parsed, nil
}

// String returns the whitespace-normalized expression.
func (this *CronExpr) String() string {
	return this.canonical
}

// Next returns the first occurrence strictly after `after`, truncated to the minute and in UTC.
//
// The bool is false when the expression has no occurrence within the search horizon, which
// callers must treat as "no further occurrence" rather than as an error: an expression can be
// perfectly valid and still never occur. Returning a zero time with false, rather than a
// far-future sentinel, keeps that case impossible to use by accident.
//
// The result is always truncated to the minute, which is what guarantees the one-minute
// resolution and keeps occurrence timestamps canonical: two instances computing the same
// occurrence get byte-identical values.
func (this *CronExpr) Next(after time.Time) (time.Time, bool) {
	// Strictly after: truncate to the minute, then step one minute on. Without the strict
	// step, Next(t) for an expression matching t would return t itself, and a caller
	// advancing a schedule by calling Next in a loop would never move.
	candidate := after.UTC().Truncate(time.Minute).Add(time.Minute)
	limit := candidate.AddDate(searchHorizonYears, 0, 0)

	for candidate.Before(limit) {
		if !maskHas(this.month, int(candidate.Month())) {
			candidate = startOfNextMonth(candidate)
			continue
		}
		if !this.dayMatches(candidate) {
			candidate = startOfNextDay(candidate)
			continue
		}
		if !maskHas(this.hour, candidate.Hour()) {
			candidate = startOfNextHour(candidate)
			continue
		}
		if !maskHas(this.minute, candidate.Minute()) {
			candidate = candidate.Add(time.Minute)
			continue
		}
		return candidate, true
	}

	return time.Time{}, false
}

// dayMatches applies cron's day rule, which is the one place the two day fields interact.
//
// When both day-of-month and day-of-week are restricted the day matches if EITHER matches,
// not both: "0 0 13 * 5" fires on the 13th of every month and additionally on every Friday.
// When only one is restricted, only that one is consulted. When neither is, every day matches.
//
// This is why the parser records which fields were literally "*" instead of inferring it from
// the masks: an unrestricted field compiles to a mask with every bit set, which an explicit
// list of every value would also produce, and the two must behave differently here.
func (this *CronExpr) dayMatches(candidate time.Time) bool {
	domOk := maskHas(this.dayOfMonth, candidate.Day())
	dowOk := maskHas(this.dayOfWeek, int(candidate.Weekday()))

	switch {
	case this.domRestricted && this.dowRestricted:
		return domOk || dowOk
	case this.domRestricted:
		return domOk
	case this.dowRestricted:
		return dowOk
	default:
		return true
	}
}

func startOfNextMonth(from time.Time) time.Time {
	return time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0)
}

func startOfNextDay(from time.Time) time.Time {
	return time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, 1)
}

func startOfNextHour(from time.Time) time.Time {
	return time.Date(from.Year(), from.Month(), from.Day(), from.Hour(), 0, 0, 0, time.UTC).Add(time.Hour)
}
