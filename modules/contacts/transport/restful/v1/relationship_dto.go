package v1

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
	"github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/relationship"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
)

type RelationshipDto struct {
	Id            string                 `json:"id"`
	PartyFromId   string                 `json:"partyFromId"`
	PartyToId     string                 `json:"partyToId"`
	RelationType  string                 `json:"relationType"`
	ValidFrom     *int64                 `json:"validFrom,omitempty"`
	ValidTo       *int64                 `json:"validTo,omitempty"`
	IsActive      bool                   `json:"isActive"`
	ExtraMetadata map[string]interface{} `json:"extraMetadata,omitempty"`
	CreatedAt     int64                  `json:"createdAt"`
	UpdatedAt     *int64                 `json:"updatedAt,omitempty"`
	Etag          string                 `json:"etag"`

	PartyFrom *PartyDto `json:"partyFrom,omitempty"`
	PartyTo   *PartyDto `json:"partyTo,omitempty"`
}

func (this *RelationshipDto) FromRelationship(relationshipEntity domain.Relationship) {
	model.MustCopy(relationshipEntity.AuditableBase, this)
	model.MustCopy(relationshipEntity.ModelBase, this)
	model.MustCopy(relationshipEntity, this)
}

type CreateRelationshipRequest = relationship.CreateRelationshipCommand
type CreateRelationshipResponse = httpserver.RestCreateResponse

type UpdateRelationshipRequest = relationship.UpdateRelationshipCommand
type UpdateRelationshipResponse = httpserver.RestUpdateResponse

type DeleteRelationshipRequest = relationship.DeleteRelationshipCommand
type DeleteRelationshipResponse = httpserver.RestDeleteResponse

type GetRelationshipByIdRequest = relationship.GetRelationshipByIdQuery
type GetRelationshipByIdResponse = RelationshipDto

type SearchRelationshipsRequest = relationship.SearchRelationshipsQuery

type SearchRelationshipsResponse httpserver.RestSearchResponse[RelationshipDto]

func (this *SearchRelationshipsResponse) FromResult(result *relationship.SearchRelationshipsResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = array.Map(result.Items, func(relationshipEntity domain.Relationship) RelationshipDto {
		item := RelationshipDto{}
		item.FromRelationship(relationshipEntity)
		return item
	})
}
