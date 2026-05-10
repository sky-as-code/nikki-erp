package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/relationship"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
)

type CreateRelationshipRequest = relationship.CreateRelationshipCommand
type CreateRelationshipResponse = httpserver.RestCreateResponse

type UpdateRelationshipRequest = relationship.UpdateRelationshipCommand
type UpdateRelationshipResponse = httpserver.RestMutateResponse

type DeleteRelationshipRequest = relationship.DeleteRelationshipCommand
type DeleteRelationshipResponse = httpserver.RestMutateResponse

type GetRelationshipRequest = relationship.GetRelationshipQuery
type GetRelationshipResponse = dmodel.DynamicFields

type RelationshipExistsRequest = relationship.RelationshipExistsQuery
type RelationshipExistsResponse = dynamicmodel.ExistsResultData

type SearchRelationshipsRequest = relationship.SearchRelationshipsQuery
type SearchRelationshipsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
