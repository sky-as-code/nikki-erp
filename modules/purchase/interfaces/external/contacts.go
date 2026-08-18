package external

import (
	itVendor "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/vendor"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itCurrency "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/currency"
)

// VendorExtService is Purchase's port onto Contacts' vendor capability.
//
// D3: "is a vendor" means "has a contacts_vendor_profile row in this organization", which is a
// checkable fact. Purchase holds a plain party id and asks Contacts whether that party may be
// ordered from, rather than deciding for itself from a field it would have to keep in step.
type VendorExtService interface {
	// GetVendor reads the vendor's terms, to default a new order's currency and lead time.
	GetVendor(ctx corectx.Context, query GetVendorQuery) (*GetVendorResult, error)

	// AssertOrderable refuses a vendor a new order may not name. It is narrower than GetVendor on
	// purpose: a suspended vendor stays readable so orders already placed against it still
	// resolve, while a new one cannot select it.
	AssertOrderable(ctx corectx.Context, query AssertOrderableQuery) (*AssertOrderableResult, error)
}

type GetVendorQuery = itVendor.GetVendorQuery
type GetVendorResult = itVendor.GetVendorResult
type GetVendorResultData = itVendor.GetVendorResultData
type AssertOrderableQuery = itVendor.AssertOrderableQuery
type AssertOrderableResult = itVendor.AssertOrderableResult

// CurrencyExtService is Purchase's port onto Essential's currency capability.
//
// Rounding goes through Round rather than being arithmetic Purchase writes, so that every module
// recording money resolves half-way cases identically — two modules rounding 0.125 differently
// produce totals that will not reconcile.
//
// There is no conversion here and there is meant to be none: no rate model exists, so a converted
// amount would be fiction. One order is denominated in one currency (D5).
type CurrencyExtService interface {
	// GetCurrency reads a currency, including the decimal_places that drive rounding.
	GetCurrency(ctx corectx.Context, query GetCurrencyQuery) (*GetCurrencyResult, error)

	// AssertUsable refuses a currency a new order may not be denominated in.
	AssertUsable(ctx corectx.Context, query AssertUsableQuery) (*AssertUsableResult, error)
}

type GetCurrencyQuery = itCurrency.GetCurrencyQuery
type GetCurrencyResult = itCurrency.GetCurrencyResult
type GetCurrencyResultData = itCurrency.GetCurrencyResultData
type AssertUsableQuery = itCurrency.AssertUsableQuery
type AssertUsableResult = itCurrency.AssertUsableResult
