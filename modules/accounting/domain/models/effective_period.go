package models

import (
	"github.com/sky-as-code/nikki-erp/common/model"
)

// Effective-period arithmetic, shared by every versioned tax configuration.
//
// Both bounds are inclusive (a closed interval), and a nil upper bound means open-ended. Dates are
// compared as YYYY-MM-DD strings, not time.Time: the format sorts lexicographically, so a string
// comparison is exactly a calendar comparison and cannot pick up a time-of-day or zone offset. The
// bounds are calendar dates because a rate changes at midnight in a jurisdiction, not at an instant
// on the server clock.

// periodContains reports whether taxDate falls inside the closed interval [from, to]. A nil from
// means not matching, not "since forever": effective_from is required, so nil means a malformed row,
// and treating that as universally applicable is how a wrong rate reaches an invoice.
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

// PeriodsOverlap reports whether two closed-or-open periods share at least one day. Publish-time
// validation uses it: two published rate versions of the same tax must never overlap, so a tax date
// resolves to exactly one rate. A nil lower bound yields false, as in periodContains, so malformed
// configuration cannot silently widen.
func PeriodsOverlap(
	fromA *model.ModelDate, toA *model.ModelDate,
	fromB *model.ModelDate, toB *model.ModelDate,
) bool {
	if fromA == nil || fromB == nil {
		return false
	}

	// Two periods overlap unless one ends strictly before the other begins; the negated form needs no
	// separate branch for open-ended bounds.
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

// IsWellFormedDate reports whether a string is a calendar date of the form YYYY-MM-DD. The check is
// structural rather than a time.Parse because parsing accepts forms that do not sort correctly
// against the stored bounds ("2026-1-5"), which would make a rate lookup silently miss. The day is
// deliberately not validated against the month's length; the database refuses those.
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
