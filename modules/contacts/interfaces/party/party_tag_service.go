package party

import (
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	tag "github.com/sky-as-code/nikki-erp/modules/core/tag/interfaces"
)

type PartyTagService interface {
	// Underlying generic Tag service (optional exposure)
	TagSvc() tag.TagService

	CreatePartyTag(ctx crud.Context, cmd CreatePartyTagCommand) (*CreatePartyTagResult, error)
	UpdatePartyTag(ctx crud.Context, cmd UpdatePartyTagCommand) (*UpdatePartyTagResult, error)
	DeletePartyTag(ctx crud.Context, cmd DeletePartyTagCommand) (*DeletePartyTagResult, error)

	PartyTagExistsMulti(ctx crud.Context, query PartyTagExistsMultiQuery) (*PartyTagExistsMultiResult, error)
	GetPartyTagById(ctx crud.Context, query GetPartyByIdTagQuery) (*GetPartyTagByIdResult, error)
	ListPartyTags(ctx crud.Context, query ListPartyTagsQuery) (*ListPartyTagsResult, error)
}
