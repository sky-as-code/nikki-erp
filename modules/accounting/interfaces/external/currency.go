package external

import (
	itCurrency "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/currency"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// CurrencyExtService is Accounting's port onto Essential's currency catalogue, narrowed to
// resolving a configured currency id into a code to render and a scale to round to.
//
// It uses GetCurrency, not AssertUsable, on purpose: AssertUsable refuses an archived or inactive
// currency, but an organization whose currency was later deactivated still has books denominated in
// it, and refusing to name it would make every existing amount unreadable.
type CurrencyExtService interface {
	GetCurrency(ctx corectx.Context, query GetCurrencyQuery) (*GetCurrencyResult, error)

	// GetCurrencyByCode resolves the code a new organization's defaults name its currency by.
	GetCurrencyByCode(ctx corectx.Context, query GetCurrencyByCodeQuery) (*GetCurrencyByCodeResult, error)
}

type GetCurrencyQuery = itCurrency.GetCurrencyQuery
type GetCurrencyByCodeQuery = itCurrency.GetCurrencyByCodeQuery
type GetCurrencyByCodeResult = itCurrency.GetCurrencyByCodeResult
type GetCurrencyResult = itCurrency.GetCurrencyResult
type GetCurrencyResultData = itCurrency.GetCurrencyResultData
