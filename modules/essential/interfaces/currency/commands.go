package currency

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// The CRUD commands, queries and results of the currency resource. They are the composable
// shapes under a resource-specific name, so that a service signature reads as this module's
// while the wire type stays the schema-agnostic field map.
type (
	CreateCurrencyCommand      = dynamicresource.CreateCommand
	UpdateCurrencyCommand      = dynamicresource.UpdateCommand
	DeleteCurrencyCommand      = dynamicresource.DeleteCommand
	SetCurrencyArchivedCommand = dynamicresource.SetArchivedCommand
	GetCurrencyByIdQuery       = dynamicresource.GetByIdQuery
	SearchCurrenciesQuery      = dynamicresource.SearchQuery
	CurrencyExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateCurrencyResult      = dynamicresource.CreateResult
	UpdateCurrencyResult      = dynamicresource.MutateResult
	DeleteCurrencyResult      = dynamicresource.MutateResult
	SetCurrencyArchivedResult = dynamicresource.MutateResult
	GetCurrencyByIdResult     = dynamicresource.GetOneResult
	SearchCurrenciesResult    = dynamicresource.SearchResult
	CurrencyExistsResult      = dynamicresource.ExistsResult
)
