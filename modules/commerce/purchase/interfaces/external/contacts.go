package external

import (
	itVendor "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/vendor"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itCurrency "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/currency"
)

// VendorExtService is Purchase's port onto Contacts' vendor capability. "Is a vendor" means "has a
// contacts_vendor_profile row in this organization"; Purchase holds a plain party id and asks
// Contacts rather than keeping a field of its own in step.
type VendorExtService interface {
	// GetVendor reads the vendor's terms, to default a new order's currency and lead time.
	GetVendor(ctx corectx.Context, query GetVendorQuery) (*GetVendorResult, error)

	// AssertOrderable refuses a vendor a new order may not name. It is narrower than GetVendor on
	// purpose: a suspended vendor stays readable so existing orders resolve, but cannot be selected.
	AssertOrderable(ctx corectx.Context, query AssertOrderableQuery) (*AssertOrderableResult, error)
}

type GetVendorQuery = itVendor.GetVendorQuery
type GetVendorResult = itVendor.GetVendorResult
type GetVendorResultData = itVendor.GetVendorResultData
type AssertOrderableQuery = itVendor.AssertOrderableQuery
type AssertOrderableResult = itVendor.AssertOrderableResult

// CurrencyExtService is Purchase's port onto Essential's currency capability. Rounding goes through
// Round rather than arithmetic Purchase writes, so every module resolves half-way cases identically
// and totals reconcile. There is deliberately no conversion: no rate model exists, and one order is
// denominated in one currency.
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
