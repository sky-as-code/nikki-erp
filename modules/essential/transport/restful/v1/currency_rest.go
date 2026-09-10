package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// CurrencyEngineName is the container name of the currency onion. A struct tag cannot call
// composable.EngineDependencyName, so the literal is declared once here and checked against it
// by a test.
const CurrencyEngineName = "dynengine_essential_currency"

type currencyRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_essential_currency"`
}

func NewCurrencyRest(params currencyRestParams) *CurrencyRest {
	rest := &CurrencyRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

// CurrencyRest serves the built-in CRUD of the currency resource. Currency is reference data
// with no lifecycle beyond archiving, so it carries no route of its own: withdrawing a currency
// from use is is_active, a plain field, because amounts already recorded in it must stay
// readable either way.
type CurrencyRest struct {
	composable.CrudRestBase
}
