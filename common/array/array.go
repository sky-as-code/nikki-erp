package array

import "github.com/thoas/go-funk"

func Map[TSrc any, TDest any](array []TSrc, mapper func(TSrc) TDest) []TDest {
	return funk.Map(array, mapper).([]TDest)
}
