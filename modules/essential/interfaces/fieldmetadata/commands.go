package fieldmetadata

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain"
)

func init() {
	var req cqrs.Request
	req = (*CreateFieldMetadataCommand)(nil)
	req = (*DeleteFieldMetadataCommand)(nil)
	req = (*GetFieldMetadataQuery)(nil)
	req = (*SearchFieldMetadataQuery)(nil)
	req = (*UpdateFieldMetadataCommand)(nil)
	req = (*FieldMetadataExistsQuery)(nil)
	util.Unused(req)
}

var createFieldMetadataCommandType = cqrs.RequestType{Module: "essential", Submodule: "field_metadata", Action: "create"}

type CreateFieldMetadataCommand struct{ domain.FieldMetadata }

func (CreateFieldMetadataCommand) CqrsRequestType() cqrs.RequestType {
	return createFieldMetadataCommandType
}
func (CreateFieldMetadataCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.FieldMetadataSchemaName)
}

type CreateFieldMetadataResult = dyn.OpResult[domain.FieldMetadata]

var updateFieldMetadataCommandType = cqrs.RequestType{Module: "essential", Submodule: "field_metadata", Action: "update"}

type UpdateFieldMetadataCommand struct{ domain.FieldMetadata }

func (UpdateFieldMetadataCommand) CqrsRequestType() cqrs.RequestType {
	return updateFieldMetadataCommandType
}
func (UpdateFieldMetadataCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.FieldMetadataSchemaName)
}

type UpdateFieldMetadataResult = dyn.OpResult[dyn.MutateResultData]

var deleteFieldMetadataCommandType = cqrs.RequestType{Module: "essential", Submodule: "field_metadata", Action: "delete"}

type DeleteFieldMetadataCommand dyn.DeleteOneCommand

func (DeleteFieldMetadataCommand) CqrsRequestType() cqrs.RequestType {
	return deleteFieldMetadataCommandType
}

type DeleteFieldMetadataResult = dyn.OpResult[dyn.MutateResultData]

var getFieldMetadataQueryType = cqrs.RequestType{Module: "essential", Submodule: "field_metadata", Action: "get"}

type GetFieldMetadataQuery dyn.GetOneQuery

func (GetFieldMetadataQuery) CqrsRequestType() cqrs.RequestType { return getFieldMetadataQueryType }

type GetFieldMetadataResult = dyn.OpResult[domain.FieldMetadata]

var searchFieldMetadataQueryType = cqrs.RequestType{Module: "essential", Submodule: "field_metadata", Action: "search"}

type SearchFieldMetadataQuery dyn.SearchQuery

func (SearchFieldMetadataQuery) CqrsRequestType() cqrs.RequestType {
	return searchFieldMetadataQueryType
}

type SearchFieldMetadataResultData = dyn.PagedResultData[domain.FieldMetadata]
type SearchFieldMetadataResult = dyn.OpResult[SearchFieldMetadataResultData]

var fieldMetadataExistsQueryType = cqrs.RequestType{Module: "essential", Submodule: "field_metadata", Action: "exists"}

type FieldMetadataExistsQuery dyn.ExistsQuery

func (FieldMetadataExistsQuery) CqrsRequestType() cqrs.RequestType {
	return fieldMetadataExistsQueryType
}

type FieldMetadataExistsResult = dyn.OpResult[dyn.ExistsResultData]
