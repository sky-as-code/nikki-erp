package party

import (
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type PartyService interface {
	CreateParty(ctx crud.Context, cmd CreatePartyCommand) (*CreatePartyResult, error)
	DeleteParty(ctx crud.Context, cmd DeletePartyCommand) (*DeletePartyResult, error)
	GetPartyById(ctx crud.Context, query GetPartyByIdQuery) (*GetPartyByIdResult, error)
	SearchParties(ctx crud.Context, query SearchPartiesQuery) (*SearchPartiesResult, error)
	UpdateParty(ctx crud.Context, cmd UpdatePartyCommand) (*UpdatePartyResult, error)
}
