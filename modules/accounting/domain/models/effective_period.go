package models

import (
	"github.com/sky-as-code/nikki-erp/common/model"
)

// Effective-period arithmetic, shared by every versioned tax configuration.
//
// Dates are compared as YYYY-MM-DD strings rather than as time.Time. That is deliberate: the
// format sorts lexicographically, so a string comparison is exactly a calendar comparison, and it
// cannot pick up a time-of-day or a zone offset along the way. BR-TAX-ESS-004 requires these
// bounds to be calendar dates precisely because a tax rate changes at midnight in a jurisdiction,
// not at an instant on the server's clock.
//
// Both bounds are inclusive, and a nil upper bound means open-ended.

// periodContains reports whether taxDate falls inside [from, to].
//
// A nil from is treated as not matching rather than as "since forever": every versioned resource
// declares effective_from as required, so a nil there means the row is malformed, and silently
// treating malformed configuration as universally applicable is how a wrong rate reaches a
// customer invoice.
func periodContains(from *model.ModelDate, to *model.ModelDate, taxDate string) bool {
	if from == nil || taxDate == "" {
		return false
	}
	if from.String() > taxDate {
		return false
	}
	if to != nil && to.String() < taxDate {
		return false
	}
	return true
}

// PeriodContains is the exported form, for callers outside this package.
func PeriodContains(from *model.ModelDate, to *model.ModelDate, taxDate string) bool {
	return periodContains(from, to, taxDate)
}

// PeriodsOverlap reports whether two closed-or-open periods share at least one day.
//
// This is what publish-time validation asks: two published rate versions of the same tax must
// never overlap, so that a tax_date resolves to exactly one rate (TAX-SUP-INV-06). Answering it
// with a search query is awkward because either upper bound may be null; answering it here keeps
// the null handling in one readable place.
//
// A nil lower bound makes the period non-comparable and the answer false, for the same reason
// periodContains rejects one: malformed configuration must not silently widen.
func PeriodsOverlap(
	fromA *model.ModelDate, toA *model.ModelDate,
	fromB *model.ModelDate, toB *model.ModelDate,
) bool {
	if fromA == nil || fromB == nil {
		return false
	}

	// Two periods overlap unless one ends strictly before the other begins. Written as the
	// negation of the two disjoint cases, which is shorter than enumerating the overlapping ones
	// and does not need a separate branch for the open-ended bounds.
	aEndsBeforeB := toA != nil && toA.String() < fromB.String()
	bEndsBeforeA := toB != nil && toB.String() < fromA.String()
	return !aEndsBeforeB && !bEndsBeforeA
}

// PeriodIsWellFormed reports whether the upper bound is on or after the lower bound.
func PeriodIsWellFormed(from *model.ModelDate, to *model.ModelDate) bool {
	if from == nil {
		return false
	}
	if to == nil {
		return true
	}
	return to.String() >= from.String()
}

// IsWellFormedDate reports whether a string is a calendar date of the form YYYY-MM-DD.
//
// The check is structural rather than a time.Parse, for the same reason the comparisons above are
// string comparisons: parsing would accept forms that do not sort correctly against the stored
// bounds ("2026-1-5"), and accepting one would make a rate lookup silently miss. The day is not
// validated against the month's length — a 31st of February is refused by the database, and
// duplicating a calendar here would be a second definition of what a date is.
func IsWellFormedDate(value string) bool {
	if len(value) != 10 || value[4] != '-' || value[7] != '-' {
		return false
	}
	for index, char := range value {
		if index == 4 || index == 7 {
			continue
		}
		if char < '0' || char > '9' {
			return false
		}
	}

	month := (int(value[5]-'0') * 10) + int(value[6]-'0')
	day := (int(value[8]-'0') * 10) + int(value[9]-'0')
	return month >= 1 && month <= 12 && day >= 1 && day <= 31
}
