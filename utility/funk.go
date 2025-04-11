package utility

import (
	"github.com/thoas/go-funk"
)

// Example:
//
// var domainModels []Account
//
//	domainModels = funk.Map(dbModels, func (db AccountDb) Account {
//		return toAccount(db)
//	})
func Map[TOut any, TIn any](arr []TIn, mapFunc func(TIn) TOut) []TOut {
	result := funk.Map(arr, mapFunc)
	return result.([]TOut)
}
