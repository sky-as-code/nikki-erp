package modelmetadata

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain"
)

func init() {
	var req cqrs.Request
	req = (*CreateModelMetadataCommand)(nil)
	req = (*DeleteModelMetadataCommand)(nil)
	req = (*GetModelMetadataQuery)(nil)
	req = (*SearchModelMetadataQuery)(nil)
	req = (*UpdateModelMetadataCommand)(nil)
	req = (*ModelMetadataExistsQuery)(nil)
	util.Unused(req)
}

var createModelMetadataCommandType = cqrs.RequestType{Module: "essential", Submodule: "model_metadata", Action: "create"}

type CreateModelMetadataCommand struct{ domain.ModelMetadata }

func (CreateModelMetadataCommand) CqrsRequestType() cqrs.RequestType {
	return createModelMetadataCommandType
}
func (CreateModelMetadataCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.ModelMetadataSchemaName)
}

type CreateModelMetadataResult = dyn.OpResult[domain.ModelMetadata]

var updateModelMetadataCommandType = cqrs.RequestType{Module: "essential", Submodule: "model_metadata", Action: "update"}

type UpdateModelMetadataCommand struct{ domain.ModelMetadata }

func (UpdateModelMetadataCommand) CqrsRequestType() cqrs.RequestType {
	return updateModelMetadataCommandType
}
func (UpdateModelMetadataCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.ModelMetadataSchemaName)
}

type UpdateModelMetadataResult = dyn.OpResult[dyn.MutateResultData]

var deleteModelMetadataCommandType = cqrs.RequestType{Module: "essential", Submodule: "model_metadata", Action: "delete"}

type DeleteModelMetadataCommand dyn.DeleteOneCommand

func (DeleteModelMetadataCommand) CqrsRequestType() cqrs.RequestType {
	return deleteModelMetadataCommandType
}

type DeleteModelMetadataResult = dyn.OpResult[dyn.MutateResultData]

var getModelMetadataQueryType = cqrs.RequestType{Module: "essential", Submodule: "model_metadata", Action: "get"}

type GetModelMetadataQuery dyn.GetOneQuery

func (GetModelMetadataQuery) CqrsRequestType() cqrs.RequestType { return getModelMetadataQueryType }

type GetModelMetadataResult = dyn.OpResult[domain.ModelMetadata]

var searchModelMetadataQueryType = cqrs.RequestType{Module: "essential", Submodule: "model_metadata", Action: "search"}

type SearchModelMetadataQuery dyn.SearchQuery

func (SearchModelMetadataQuery) CqrsRequestType() cqrs.RequestType {
	return searchModelMetadataQueryType
}

type SearchModelMetadataResultData = dyn.PagedResultData[domain.ModelMetadata]
type SearchModelMetadataResult = dyn.OpResult[SearchModelMetadataResultData]

var modelMetadataExistsQueryType = cqrs.RequestType{Module: "essential", Submodule: "model_metadata", Action: "exists"}

type ModelMetadataExistsQuery dyn.ExistsQuery

func (ModelMetadataExistsQuery) CqrsRequestType() cqrs.RequestType {
	return modelMetadataExistsQueryType
}

type ModelMetadataExistsResult = dyn.OpResult[dyn.ExistsResultData]
