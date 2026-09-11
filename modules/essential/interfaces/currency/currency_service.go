package currency

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type GetCurrencyResult = dyn.OpResult[GetCurrencyResultData]
type RoundResult = dyn.OpResult[RoundResultData]
type AssertUsableResult = dyn.OpResult[struct{}]

// CurrencyRepository reads and writes currency rows. It is the composable default and nothing
// more today; it exists as its own type so a bespoke query has a home when one is needed.
type CurrencyRepository interface {
	composable.CrudRepository
}

// CurrencyLookupService answers questions about a single currency. It exists because other
// modules must be able to validate a currency reference without reaching into Essential's
// repositories.
type CurrencyLookupService interface {
	GetCurrency(ctx corectx.Context, query GetCurrencyQuery) (*GetCurrencyResult, error)
}

// CurrencyDomainService is the full capability, implemented inside Essential: the CRUD of the
// resource plus the money rules other modules depend on.
type CurrencyDomainService interface {
	composable.CrudDomainService
	CurrencyLookupService

	// Round rounds an amount to the currency's decimal_places.
	//
	// It is a service call rather than arithmetic each caller writes, because every module
	// recording money must resolve half-way cases identically — two modules rounding 0.125
	// differently produce totals that will not reconcile.
	Round(ctx corectx.Context, query RoundQuery) (*RoundResult, error)

	// AssertUsable reports whether a currency may be chosen for a NEW amount: it must exist, be
	// active, and not be archived.
	//
	// This is deliberately narrower than GetCurrency. An inactive or archived currency must stay
	// *readable*, because amounts already recorded in it still refer to it — so a historical
	// document keeps resolving its currency while a new one cannot select it. A consumer that
	// checked only existence would let a withdrawn currency back into new records.
	AssertUsable(ctx corectx.Context, query AssertUsableQuery) (*AssertUsableResult, error)
}

// CurrencyAppService is the capability other modules consume. It is the type a consuming module's
// infra/external/index.go binds to its own local port, and deliberately excludes the CRUD.
type CurrencyAppService interface {
	CurrencyLookupService

	Round(ctx corectx.Context, query RoundQuery) (*RoundResult, error)
	AssertUsable(ctx corectx.Context, query AssertUsableQuery) (*AssertUsableResult, error)
}

// CurrencyApplicationService is the whole application layer of the resource: the authorized CRUD
// the REST handler serves, plus the cross-module capability.
type CurrencyApplicationService interface {
	composable.CrudApplicationService
	CurrencyAppService
}
