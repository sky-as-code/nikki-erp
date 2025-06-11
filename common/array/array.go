package array

import "github.com/thoas/go-funk"

func Contains[TSrc any](array []TSrc, item TSrc) bool {
	return funk.Contains(array, item)
}

func Map[TSrc any, TDest any](array []TSrc, mapper func(TSrc) TDest) []TDest {
	return funk.Map(array, mapper).([]TDest)
}

func Filter[TSrc any](array []TSrc, predicate func(TSrc) bool) []TSrc {
	return funk.Filter(array, predicate).([]TSrc)
}
