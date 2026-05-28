package party

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type PartyService interface {
	CreateParty(ctx corectx.Context, cmd CreatePartyCommand) (*CreatePartyResult, error)
	DeleteParty(ctx corectx.Context, cmd DeletePartyCommand) (*DeletePartyResult, error)
	GetParty(ctx corectx.Context, query GetPartyQuery) (*GetPartyResult, error)
	SearchParties(ctx corectx.Context, query SearchPartiesQuery) (*SearchPartiesResult, error)
	PartyExists(ctx corectx.Context, query PartyExistsQuery) (*PartyExistsResult, error)
	UpdateParty(ctx corectx.Context, cmd UpdatePartyCommand) (*UpdatePartyResult, error)
}
