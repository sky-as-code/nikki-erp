package sequence

import (
	"fmt"
	"time"

	"go.bryk.io/pkg/errors"
)

const (
	ResetNever   = "never"
	ResetYearly  = "yearly"
	ResetMonthly = "monthly"
	ResetDaily   = "daily"

	PeriodKeyNone = "NONE"
)

// PeriodResult is the reset period an instant falls in. Start and NextStart are local midnights in the
// definition timezone and are zero for "never".
type PeriodResult struct {
	Key       string
	Start     time.Time
	NextStart time.Time
}

// PeriodOf returns the period of t in timezone (IANA name; empty means UTC). Boundaries come from calendar
// arithmetic (time.Date), never from adding 24 hours, so DST days are 23 or 25 hours long as they should be.
func PeriodOf(t time.Time, timezone string, resetPeriod string) (PeriodResult, error) {
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return PeriodResult{}, errors.Wrapf(err, "sequence: invalid timezone %q", timezone)
	}
	local := t.In(location)
	year, month, day := local.Date()
	switch resetPeriod {
	case ResetNever:
		return PeriodResult{Key: PeriodKeyNone}, nil
	case ResetYearly:
		return PeriodResult{
			Key:       fmt.Sprintf("%04d", year),
			Start:     time.Date(year, time.January, 1, 0, 0, 0, 0, location),
			NextStart: time.Date(year+1, time.January, 1, 0, 0, 0, 0, location),
		}, nil
	case ResetMonthly:
		return PeriodResult{
			Key:       fmt.Sprintf("%04d-%02d", year, int(month)),
			Start:     time.Date(year, month, 1, 0, 0, 0, 0, location),
			NextStart: time.Date(year, month+1, 1, 0, 0, 0, 0, location),
		}, nil
	case ResetDaily:
		return PeriodResult{
			Key:       fmt.Sprintf("%04d-%02d-%02d", year, int(month), day),
			Start:     time.Date(year, month, day, 0, 0, 0, 0, location),
			NextStart: time.Date(year, month, day+1, 0, 0, 0, 0, location),
		}, nil
	default:
		return PeriodResult{}, errors.Errorf("sequence: invalid reset period %q", resetPeriod)
	}
}
