package external

import (
	itAccCurrency "github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/currency"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// OrgCurrencyExtService is Sales' port onto the organization's book currency. It points at
// accounting, not essential: essential owns the currency catalogue, but which one an organization
// keeps its books in is accounting policy. Product's base_sales_price and cost carry no currency of
// their own and are denominated in the book currency by convention, so pricing needs this port to
// know what those numbers mean. A nil port fails CLOSED here (unlike the variant-sellability port):
// an amount of unknown denomination must never be used, or 100 USD silently becomes 100 VND.
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
