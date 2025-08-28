package party

import (
	"context"
)

type PartyService interface {
	CreateParty(ctx context.Context, cmd CreatePartyCommand) (*CreatePartyResult, error)
	DeleteParty(ctx context.Context, cmd DeletePartyCommand) (*DeletePartyResult, error)
	GetPartyById(ctx context.Context, query GetPartyByIdQuery) (*GetPartyByIdResult, error)
	SearchParties(ctx context.Context, query SearchPartiesQuery) (*SearchPartiesResult, error)
	UpdateParty(ctx context.Context, cmd UpdatePartyCommand) (*UpdatePartyResult, error)
}
