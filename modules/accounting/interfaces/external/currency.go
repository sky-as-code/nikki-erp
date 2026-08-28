package external

import (
	itCurrency "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/currency"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// CurrencyExtService is Accounting's port onto Essential's currency catalogue.
//
// Narrowed to a single read. Accounting resolves the organization's configured currency id into
// something a caller can use — a code to render, a scale to round to — and needs nothing else from
// the catalogue. In particular it does not take Round: rounding an amount is the caller's business
// and Essential already publishes that to whoever needs it, so routing it through here would put
// Accounting in the middle of a conversation it has no part in.
//
// GetCurrency rather than AssertUsable, deliberately. AssertUsable asks whether a currency may be
// chosen for a NEW amount, and refuses an archived or inactive one. That is the right question when
// a user picks a currency, and the wrong one here: an organization whose currency was later
// deactivated still has books denominated in it, and refusing to name it would make every existing
// amount unreadable rather than merely un-extendable.
type CurrencyExtService interface {
	GetCurrency(ctx corectx.Context, query GetCurrencyQuery) (*GetCurrencyResult, error)
}

type GetCurrencyQuery = itCurrency.GetCurrencyQuery
type GetCurrencyResult = itCurrency.GetCurrencyResult
type GetCurrencyResultData = itCurrency.GetCurrencyResultData
