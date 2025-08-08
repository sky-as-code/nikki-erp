package v1

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
	"github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/party"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
)

type PartyDto struct {
	Id           string  `json:"id"`
	AvatarUrl    *string `json:"avatarUrl,omitempty"`
	CreatedAt    int64   `json:"createdAt"`
	DisplayName  string  `json:"displayName"`
	LegalName    *string `json:"legalName,omitempty"`
	LegalAddress *string `json:"legalAddress,omitempty"`
	TaxId        *string `json:"taxId,omitempty"`
	JobPosition  *string `json:"jobPosition,omitempty"`
	Title        *string `json:"title,omitempty"`
	Type         string  `json:"type"`
	Note         *string `json:"note,omitempty"`
	Nationality  *string `json:"nationality,omitempty"`
	Website      *string `json:"website,omitempty"`
	Etag         string  `json:"etag"`
	UpdatedAt    *int64  `json:"updatedAt,omitempty"`

	CommChannels  []CommChannelDto  `json:"commChannels,omitempty"`
	Relationships []RelationshipDto `json:"relationships,omitempty"`
	Tags          []PartyTagDto     `json:"tags,omitempty"`
}

func (this *PartyDto) FromParty(partyEntity domain.Party) {
	model.MustCopy(partyEntity.AuditableBase, this)
	model.MustCopy(partyEntity.ModelBase, this)
	model.MustCopy(partyEntity, this)

	// Handle Title enum
	if partyEntity.Title != nil && partyEntity.Title.Value != nil {
		this.Title = partyEntity.Title.Value
	}

	// Handle related entities
	this.CommChannels = array.Map(partyEntity.CommChannels, func(commChannel domain.CommChannel) CommChannelDto {
		commChannelDto := CommChannelDto{}
		commChannelDto.FromCommChannel(commChannel)
		return commChannelDto
	})

	this.Relationships = array.Map(partyEntity.Relationships, func(relationship domain.Relationship) RelationshipDto {
		relationshipDto := RelationshipDto{}
		relationshipDto.FromRelationship(relationship)
		return relationshipDto
	})

	this.Tags = array.Map(partyEntity.Tags, func(tag domain.PartyTag) PartyTagDto {
		tagDto := PartyTagDto{}
		tagDto.FromTag(tag)
		return tagDto
	})
}

type PartyTagDto struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

func (this *PartyTagDto) FromTag(tag domain.PartyTag) {
	model.MustCopy(tag, this)
}

type CreatePartyRequest = party.CreatePartyCommand
type CreatePartyResponse = httpserver.RestCreateResponse

type UpdatePartyRequest = party.UpdatePartyCommand
type UpdatePartyResponse = httpserver.RestUpdateResponse

type DeletePartyRequest = party.DeletePartyCommand
type DeletePartyResponse = httpserver.RestDeleteResponse

type GetPartyByIdRequest = party.GetPartyByIdQuery
type GetPartyByIdResponse = PartyDto

type SearchPartiesRequest = party.SearchPartiesQuery

type SearchPartiesResponse httpserver.RestSearchResponse[PartyDto]

func (this *SearchPartiesResponse) FromResult(result *party.SearchPartiesResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = array.Map(result.Items, func(partyEntity domain.Party) PartyDto {
		item := PartyDto{}
		item.FromParty(partyEntity)
		return item
	})
}
