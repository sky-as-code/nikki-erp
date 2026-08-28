// Package currency publishes the one question other modules ask Accounting about money that is
// not tax: which currency is this organization's books kept in?
//
// It exists because the alternative is worse. Product stores a base sales price and a cost with no
// currency column of their own (BR-PRICE-CUR-003), which is deliberate — a currency on Product,
// another on Inventory and a third on Sales would be three answers to one question, free to
// disagree the moment one is edited. So the answer lives in exactly one place, and everyone else
// asks for it.
//
// Accounting owns it rather than Essential because Essential owns the *catalogue* of currencies —
// what a currency is, how many decimal places it has, how to round in it — while which one a
// particular organization keeps its books in is an accounting policy. Essential would have to
// grow a notion of organizational policy to hold it, and it has none.
//
// What this package deliberately does NOT do is convert between currencies. There is no FX engine
// and no rate table anywhere in the system, and §35.1 of the pricing change request is explicit
// that one must not be improvised here.
package currency

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

type GetOrgDefaultCurrencyResult = dyn.OpResult[GetOrgDefaultCurrencyResultData]

// OrgCurrencyService answers what an organization's default currency is.
//
// It is one method, and narrow on purpose. A consumer needs the currency an unqualified amount is
// denominated in; it does not need to read, still less write, the organization's other accounting
// settings. Publishing the settings contract instead would hand every consumer the power to change
// the rounding policy on the way past.
type OrgCurrencyService interface {
	// GetOrgDefaultCurrency resolves the organization's book currency.
	//
	// HasData false means the organization has not configured one. That is NOT an error: a fresh
	// deployment has no setting yet, and a caller that merely wants to label an amount can say so.
	// A caller that must *interpret* an amount has to refuse instead, because assuming a currency
	// is exactly the failure BR-PRICE-CUR-004 names — and it is silent, which is what makes it
	// expensive.
	GetOrgDefaultCurrency(
		ctx corectx.Context, query GetOrgDefaultCurrencyQuery,
	) (*GetOrgDefaultCurrencyResult, error)
}

// GetOrgDefaultCurrencyQuery names the organization whose currency is wanted.
type GetOrgDefaultCurrencyQuery struct {
	// OrgId is optional. Empty means "the acting caller's organization", which the settings
	// module already derives from the request context — the same way every other effective
	// settings read works.
	OrgId string
}

// GetOrgDefaultCurrencyResultData is the resolved currency.
//
// Both the id and the code are returned because callers want different halves and neither should
// have to make a second call to get the other: a stored reference wants the id, an error message
// or a rendered amount wants the code. DecimalPlaces comes along because a caller that has the
// currency almost always needs to know how many fractional digits it is quoted to.
type GetOrgDefaultCurrencyResultData struct {
	CurrencyId    string
	Code          string
	Symbol        string
	DecimalPlaces int32
}
