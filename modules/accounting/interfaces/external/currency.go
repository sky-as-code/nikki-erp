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
}

type GetCurrencyQuery = itCurrency.GetCurrencyQuery
type GetCurrencyResult = itCurrency.GetCurrencyResult
type GetCurrencyResultData = itCurrency.GetCurrencyResultData
