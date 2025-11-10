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

func IndexOf[TSrc any](array []TSrc, item TSrc) int {
	return funk.IndexOf(array, item)
}

func RemoveString(array []string, str string) ([]string, bool) {
	found := false
	for i, v := range array {
		if v == str {
			found = true
			return append(array[:i], array[i+1:]...), true
		}
	}
	return array, found
}

func Prepend[TSrc any](array []TSrc, value TSrc) []TSrc {
	return append([]TSrc{value}, array...)
}
