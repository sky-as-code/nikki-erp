package party

import (
	"context"

	tag "github.com/sky-as-code/nikki-erp/modules/core/tag/interfaces"
)

type PartyTagService interface {
	// Underlying generic Tag service (optional exposure)
	TagSvc() tag.TagService

	CreatePartyTag(ctx context.Context, cmd CreatePartyTagCommand) (*CreatePartyTagResult, error)
	UpdatePartyTag(ctx context.Context, cmd UpdatePartyTagCommand) (*UpdatePartyTagResult, error)
	DeletePartyTag(ctx context.Context, cmd DeletePartyTagCommand) (*DeletePartyTagResult, error)

	PartyTagExistsMulti(ctx context.Context, query PartyTagExistsMultiQuery) (*PartyTagExistsMultiResult, error)
	GetPartyTagById(ctx context.Context, query GetPartyByIdTagQuery) (*GetPartyTagByIdResult, error)
	ListPartyTags(ctx context.Context, query ListPartyTagsQuery) (*ListPartyTagsResult, error)
}
