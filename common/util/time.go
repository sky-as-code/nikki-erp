package util

import "time"

func FloorHour(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
}

func FloorDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func FloorMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

func FloorYear(t time.Time) time.Time {
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
}

func EndOfHour(t time.Time) time.Time {
	return FloorHour(t).Add(time.Hour).Add(-time.Nanosecond)
}

func EndOfDay(t time.Time) time.Time {
	return FloorDay(t).AddDate(0, 0, 1).Add(-time.Nanosecond)
}

func EndOfMonth(t time.Time) time.Time {
	return FloorMonth(t).AddDate(0, 1, 0).Add(-time.Nanosecond)
}

func EndOfYear(t time.Time) time.Time {
	return FloorYear(t).AddDate(1, 0, 0).Add(-time.Nanosecond)
}
