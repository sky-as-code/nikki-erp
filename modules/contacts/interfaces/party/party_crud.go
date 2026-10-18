package party

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// PartyRepository reads and writes party rows. It is the composable default today; the type
// exists so a bespoke query has a home when one is needed.
type PartyRepository interface {
	composable.CrudRepository
}

// PartyApplicationService is the whole application layer of the resource: the authorized CRUD
// the REST handler serves, plus the cross-module capability.
type PartyApplicationService interface {
	composable.CrudApplicationService
	PartyAppService
}

type (
	CreatePartyCommand      = composable.CreateCommand
	UpdatePartyCommand      = composable.UpdateCommand
	DeletePartyCommand      = composable.DeleteCommand
	SetPartyArchivedCommand = composable.SetArchivedCommand
	GetPartyByIdQuery       = composable.GetByIdQuery
	SearchPartiesQuery      = composable.SearchQuery
	PartyExistsQuery        = composable.ExistsQuery
)

type (
	CreatePartyResult      = composable.CreateResult
	UpdatePartyResult      = composable.MutateResult
	DeletePartyResult      = composable.MutateResult
	SetPartyArchivedResult = composable.MutateResult
	GetPartyByIdResult     = composable.GetOneResult
	SearchPartiesResult    = composable.SearchResult
	PartyExistsResult      = composable.ExistsResult
)
