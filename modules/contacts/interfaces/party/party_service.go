package party

import (
	"context"

	tag "github.com/sky-as-code/nikki-erp/modules/core/tag/interfaces"
)

type PartyService interface {
	PartyTagService

	// CreateParty(ctx context.Context, cmd CreatePartyCommand) (*CreatePartyResult, error)
	// UpdateParty(ctx context.Context, cmd UpdatePartyCommand) (*UpdatePartyResult, error)
	// DeleteParty(ctx context.Context, cmd DeletePartyCommand) (*DeletePartyResult, error)
	// PartyExists(ctx context.Context, query PartyExistsQuery) (*PartyExistsResult, error)
	// PartyExistsMulti(ctx context.Context, query PartyExistsMultiQuery) (*PartyExistsMultiResult, error)
	// GetParty(ctx context.Context, query GetPartyQuery) (*GetPartyResult, error)
	// ListParty(ctx context.Context, query ListPartyQuery) (*ListPartyResult, error)
}

type PartyTagService interface {
	TagSvc() tag.TagService

	CreatePartyTag(ctx context.Context, cmd CreatePartyTagCommand) (*CreatePartyTagResult, error)
	UpdatePartyTag(ctx context.Context, cmd UpdatePartyTagCommand) (*UpdatePartyTagResult, error)
	DeletePartyTag(ctx context.Context, cmd DeletePartyTagCommand) (*DeletePartyTagResult, error)
	PartyTagExistsMulti(ctx context.Context, query PartyTagExistsMultiQuery) (*PartyTagExistsMultiResult, error)
	GetPartyTagById(ctx context.Context, query GetPartyByIdTagQuery) (*GetPartyTagByIdResult, error)
	ListPartyTags(ctx context.Context, query ListPartyTagsQuery) (*ListPartyTagsResult, error)
}
