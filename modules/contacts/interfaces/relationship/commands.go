package relationship

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

func init() {
	// Assert interface implementation
	var req cqrs.Request
	req = (*CreateRelationshipCommand)(nil)
	req = (*UpdateRelationshipCommand)(nil)
	req = (*DeleteRelationshipCommand)(nil)
	req = (*GetRelationshipQuery)(nil)
	req = (*SearchRelationshipsQuery)(nil)
	req = (*RelationshipExistsQuery)(nil)
	util.Unused(req)
}

var createRelationshipCommandType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "relationship",
	Action:    "create",
}

type CreateRelationshipCommand struct {
	domain.Relationship
}

func (CreateRelationshipCommand) CqrsRequestType() cqrs.RequestType {
	return createRelationshipCommandType
}

func (this CreateRelationshipCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.RelationshipSchemaName)
}

type CreateRelationshipResult = dyn.OpResult[domain.Relationship]

var updateRelationshipCommandType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "relationship",
	Action:    "update",
}

type UpdateRelationshipCommand struct {
	domain.Relationship
}

func (UpdateRelationshipCommand) CqrsRequestType() cqrs.RequestType {
	return updateRelationshipCommandType
}

func (this UpdateRelationshipCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.RelationshipSchemaName)
}

type UpdateRelationshipResult = dyn.OpResult[dyn.MutateResultData]

var deleteRelationshipCommandType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "relationship",
	Action:    "delete",
}

type DeleteRelationshipCommand dyn.DeleteOneCommand

func (DeleteRelationshipCommand) CqrsRequestType() cqrs.RequestType {
	return deleteRelationshipCommandType
}

type DeleteRelationshipResult = dyn.OpResult[dyn.MutateResultData]

var getRelationshipQueryType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "relationship",
	Action:    "getRelationship",
}

type GetRelationshipQuery struct {
	Columns []string `json:"columns" query:"columns"`
	Id      *string  `json:"id" param:"id"`
}

func (GetRelationshipQuery) CqrsRequestType() cqrs.RequestType {
	return getRelationshipQueryType
}

type GetRelationshipResult = dyn.OpResult[domain.Relationship]

var searchRelationshipsQueryType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "relationship",
	Action:    "search",
}

type SearchRelationshipsQuery struct {
	Fields          []string            `json:"fields" query:"fields"`
	Graph           *dmodel.SearchGraph `json:"graph" query:"graph"`
	Page            int                 `json:"page" query:"page"`
	Size            int                 `json:"size" query:"size"`
	IncludeArchived *bool               `json:"include_archived" query:"include_archived"`
	PartyId         model.Id            `json:"party_id" param:"party_id"`
}

func (SearchRelationshipsQuery) CqrsRequestType() cqrs.RequestType {
	return searchRelationshipsQueryType
}

func (SearchRelationshipsQuery) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"contacts.search_relationships_query",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				Field(dyn.DefineFieldSearchFields()).
				Field(dyn.DefineFieldSearchGraph()).
				Field(dyn.DefineFieldSearchPage()).
				Field(dyn.DefineFieldSearchSize()).
				Field(dyn.DefineFieldIncludeArchived()).
				Field(dmodel.DefineField().
					Name("party_id").
					DataType(dmodel.FieldDataTypeUlid()))
		},
	)
}

type SearchRelationshipsResultData = dyn.PagedResultData[domain.Relationship]
type SearchRelationshipsResult = dyn.OpResult[SearchRelationshipsResultData]

var relationshipExistsQueryType = cqrs.RequestType{
	Module:    "contacts",
	Submodule: "relationship",
	Action:    "exists",
}

type RelationshipExistsQuery dyn.ExistsQuery

func (RelationshipExistsQuery) CqrsRequestType() cqrs.RequestType {
	return relationshipExistsQueryType
}

type RelationshipExistsResult = dyn.OpResult[dyn.ExistsResultData]
