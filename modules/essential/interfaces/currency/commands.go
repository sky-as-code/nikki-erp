package currency

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The CRUD commands, queries and results of the currency resource. They are the composable
// shapes under a resource-specific name, so that a service signature reads as this module's
// while the wire type stays the schema-agnostic field map.
type (
	CreateCurrencyCommand      = composable.CreateCommand
	UpdateCurrencyCommand      = composable.UpdateCommand
	DeleteCurrencyCommand      = composable.DeleteCommand
	SetCurrencyArchivedCommand = composable.SetArchivedCommand
	GetCurrencyByIdQuery       = composable.GetByIdQuery
	SearchCurrenciesQuery      = composable.SearchQuery
	CurrencyExistsQuery        = composable.ExistsQuery
)

type (
	CreateCurrencyResult      = composable.CreateResult
	UpdateCurrencyResult      = composable.MutateResult
	DeleteCurrencyResult      = composable.MutateResult
	SetCurrencyArchivedResult = composable.MutateResult
	GetCurrencyByIdResult     = composable.GetOneResult
	SearchCurrenciesResult    = composable.SearchResult
	CurrencyExistsResult      = composable.ExistsResult
)
