// Package currency declares the currency capability that Essential offers to other modules.
//
// CRUD on the currency resource itself goes through the dynamic resource engine and needs no types
// here. What lives in this package is what a module recording money must be able to do without
// reaching into Essential's repositories: validate that a currency_id it holds refers to a currency
// that may be used, and round an amount to the number of fractional digits that currency is quoted
// to.
//
// A consuming module never imports this package from its domain or application layer. It declares a
// local port in its own interfaces/external/ and binds it once in infra/external/index.go — see
// docs/wiki/01 "Microservice-ready Monolith".
//
// Deliberately absent: exchange rates, a base currency, and conversion between currencies. No rate
// model exists, so a converted amount would be fiction. One document is denominated in one currency.
package currency

import (
	"github.com/shopspring/decimal"

	"github.com/sky-as-code/nikki-erp/common/model"
)

// GetCurrencyQuery fetches a single currency, for consumers that must validate a currency
// reference they hold. HasData false on the result means the currency does not exist.
type GetCurrencyQuery struct {
	Id model.Id
}

// GetCurrencyResultData exposes the parts of a currency a consuming module may legitimately
// depend on: enough to validate a reference, render an amount, and round one correctly.
//
// IsActive and IsArchived are both exposed because they answer different questions and a consumer
// that collapses them gets one of the two wrong. See AssertUsable.
type GetCurrencyResultData struct {
	Id     model.Id
	Code   string
	Symbol string

	// DecimalPlaces is how many fractional digits this currency is quoted to: 0 for VND, 2 for
	// USD, 3 for KWD. A consumer rounds with Round rather than reading this and rounding itself,
	// so that half-way cases are resolved the same way everywhere.
	DecimalPlaces int32

	// IsActive is whether new amounts may be denominated in this currency.
	IsActive bool

	// IsArchived is whether the record has been archived out of the working set.
	IsArchived bool
}

// RoundQuery asks for an amount to be rounded to the fractional precision of a currency.
type RoundQuery struct {
	Amount     decimal.Decimal
	CurrencyId model.Id
}

// RoundResultData carries the rounded amount.
type RoundResultData struct {
	Amount decimal.Decimal
}

// AssertUsableQuery asks whether a currency may be chosen for a NEW amount.
type AssertUsableQuery struct {
	Id model.Id

	// Field is the name the caller knows this reference by — "currency_id" on most documents.
	// It is echoed back on any violation so the error points at the caller's own field rather
	// than at Essential's.
	Field string
}
