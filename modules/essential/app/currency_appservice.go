package app

import (
	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itCurrency "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/currency"
)

// NewCurrencyApplicationService is handed the composable default by the currency onion. The
// domain service behind it is the module's own, so the assertion is a wiring check rather
// than a guess.
func NewCurrencyApplicationService(base composable.CrudApplicationService) itCurrency.CurrencyApplicationService {
	currencySvc, ok := base.DomainService().(itCurrency.CurrencyDomainService)
	if !ok {
		panic(errors.New("the currency onion must be built with NewCurrencyDomainService"))
	}
	return &CurrencyApplicationServiceImpl{CrudApplicationService: base, currencySvc: currencySvc}
}

// CurrencyApplicationServiceImpl is the authorized CRUD plus the capability boundary other
// modules bind to.
//
// The cross-module methods stay a thin delegation on purpose, exactly as the UoM one does: when
// Essential is split into its own service, this is the type a REST client replaces, and any
// logic living here would have to be duplicated. They assert no permission because the caller is
// another module acting on its own authority, not a user reading currencies.
type CurrencyApplicationServiceImpl struct {
	composable.CrudApplicationService
	currencySvc itCurrency.CurrencyDomainService
}

func (this *CurrencyApplicationServiceImpl) GetCurrency(
	ctx corectx.Context, query itCurrency.GetCurrencyQuery,
) (*itCurrency.GetCurrencyResult, error) {
	return this.currencySvc.GetCurrency(ctx, query)
}

func (this *CurrencyApplicationServiceImpl) GetCurrencyByCode(
	ctx corectx.Context, query itCurrency.GetCurrencyByCodeQuery,
) (*itCurrency.GetCurrencyByCodeResult, error) {
	return this.currencySvc.GetCurrencyByCode(ctx, query)
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
