// Package currency publishes which currency an organization keeps its books in. Product prices and
// costs deliberately have no currency column of their own, so this is the single answer everyone
// asks for. It does not convert between currencies: there is no FX engine or rate table in the
// system, and one must not be improvised here.
package currency

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

type GetOrgDefaultCurrencyResult = dyn.OpResult[GetOrgDefaultCurrencyResultData]

// OrgCurrencyService answers what an organization's default currency is. It is narrow on purpose:
// publishing the whole settings contract would let every consumer change the rounding policy too.
type OrgCurrencyService interface {
	// GetOrgDefaultCurrency resolves the organization's book currency. HasData false means none is
	// configured, which is not an error; a caller that must interpret an amount has to refuse rather
	// than assume a currency.
	GetOrgDefaultCurrency(
		ctx corectx.Context, query GetOrgDefaultCurrencyQuery,
	) (*GetOrgDefaultCurrencyResult, error)
}

// GetOrgDefaultCurrencyQuery names the organization whose currency is wanted.
type GetOrgDefaultCurrencyQuery struct {
	// OrgId is optional; empty means the acting caller's organization, derived from the request
	// context.
	OrgId string
}

// GetOrgDefaultCurrencyResultData is the resolved currency. DecimalPlaces is how many fractional
// digits amounts in this currency are quoted to.
type GetOrgDefaultCurrencyResultData struct {
	CurrencyId    string
	Code          string
	Symbol        string
	DecimalPlaces int32
}
