package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itCurrency "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/currency"
)

func NewCurrencyApplicationServiceImpl(
	currencySvc itCurrency.CurrencyDomainService,
) itCurrency.CurrencyAppService {
	return &CurrencyApplicationServiceImpl{currencySvc: currencySvc}
}

// CurrencyApplicationServiceImpl is the capability boundary other modules bind to.
//
// It stays a thin delegation on purpose, exactly as the UoM one does: when Essential is split into
// its own service, this is the type a REST client replaces, and any logic living here would have to
// be duplicated.
//
// Purchase is the first consumer ([PUR-018]) and the reason this exists. Until it did, the
// interface documented a binding that nothing provided — which is not a compile error, so it
// surfaced only when a consuming module failed to start.
type CurrencyApplicationServiceImpl struct {
	currencySvc itCurrency.CurrencyDomainService
}

func (this *CurrencyApplicationServiceImpl) GetCurrency(
	ctx corectx.Context, query itCurrency.GetCurrencyQuery,
) (*itCurrency.GetCurrencyResult, error) {
	return this.currencySvc.GetCurrency(ctx, query)
}

func (this *CurrencyApplicationServiceImpl) Round(
	ctx corectx.Context, query itCurrency.RoundQuery,
) (*itCurrency.RoundResult, error) {
	return this.currencySvc.Round(ctx, query)
}

func (this *CurrencyApplicationServiceImpl) AssertUsable(
	ctx corectx.Context, query itCurrency.AssertUsableQuery,
) (*itCurrency.AssertUsableResult, error) {
	return this.currencySvc.AssertUsable(ctx, query)
}
