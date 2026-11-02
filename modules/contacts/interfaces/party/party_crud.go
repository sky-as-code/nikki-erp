package party

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// PartyRepository reads and writes party rows. It is the composable default today; the type
// exists so a bespoke query has a home when one is needed.
type PartyRepository interface {
	dynamicresource.CrudRepository
}

// PartyApplicationService is the whole application layer of the resource: the authorized CRUD
// the REST handler serves, plus the cross-module capability.
type PartyApplicationService interface {
	dynamicresource.CrudApplicationService
	PartyAppService
}

type (
	CreatePartyCommand      = dynamicresource.CreateCommand
	UpdatePartyCommand      = dynamicresource.UpdateCommand
	DeletePartyCommand      = dynamicresource.DeleteCommand
	SetPartyArchivedCommand = dynamicresource.SetArchivedCommand
	GetPartyByIdQuery       = dynamicresource.GetByIdQuery
	SearchPartiesQuery      = dynamicresource.SearchQuery
	PartyExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreatePartyResult      = dynamicresource.CreateResult
	UpdatePartyResult      = dynamicresource.MutateResult
	DeletePartyResult      = dynamicresource.MutateResult
	SetPartyArchivedResult = dynamicresource.MutateResult
	GetPartyByIdResult     = dynamicresource.GetOneResult
	SearchPartiesResult    = dynamicresource.SearchResult
	PartyExistsResult      = dynamicresource.ExistsResult
)
