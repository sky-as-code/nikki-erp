package external

import (
	itAccCurrency "github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/currency"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// OrgCurrencyExtService is Sales' port onto the organization's book currency.
//
// It points at ACCOUNTING, not Essential, and the distinction matters. Essential owns the
// catalogue — what a currency is, how it rounds — while which one an organization keeps its books
// in is an accounting policy. Sales needs the latter.
//
// Why Sales needs it at all: a pricelist carries its own currency_id, and a sales order carries a
// currency_code, but Product's base_sales_price and cost carry neither (BR-PRICE-CUR-003). They are
// denominated in the organization's book currency by convention. So when pricing falls back to the
// base sales price, or a FORMULA rule reads cost, Sales is holding a number whose currency is not
// written down anywhere near it — and this is what says what that number means.
//
// A nil port FAILS CLOSED for that comparison, deliberately, and the opposite way round from the
// variant-sellability port. An unresolved currency means an amount of unknown denomination; using
// it anyway is how 100 USD silently becomes 100 VND, which BR-PRICE-CUR-004 names as the thing that
// must not happen. Refusing is recoverable and loud; guessing is neither.
type OrgCurrencyExtService interface {
	// GetOrgDefaultCurrency resolves the organization's book currency. HasData false means none is
	// configured — a fresh deployment, not a failure.
	GetOrgDefaultCurrency(
		ctx corectx.Context, query GetOrgDefaultCurrencyQuery,
	) (*GetOrgDefaultCurrencyResult, error)
}

type GetOrgDefaultCurrencyQuery = itAccCurrency.GetOrgDefaultCurrencyQuery
type GetOrgDefaultCurrencyResult = itAccCurrency.GetOrgDefaultCurrencyResult
type GetOrgDefaultCurrencyResultData = itAccCurrency.GetOrgDefaultCurrencyResultData
