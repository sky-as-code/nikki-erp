package resource

import (
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	domain "github.com/sky-as-code/nikki-erp/modules/identity/domain/models"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

func init() {
	var req cqrs.Request
	req = (*CreateResourceCommand)(nil)
	req = (*DeleteResourceCommand)(nil)
	req = (*GetResourceQuery)(nil)
	req = (*ResourceExistsQuery)(nil)
	req = (*SearchResourcesQuery)(nil)
	req = (*UpdateResourceCommand)(nil)
	util.Unused(req)
}

var createResourceCommandType = cqrs.RequestType{Module: "identity", Submodule: "resource", Action: "createResource"}

type CreateResourceCommand struct {
	domain.Resource
}

func (CreateResourceCommand) CqrsRequestType() cqrs.RequestType { return createResourceCommandType }

func (CreateResourceCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.ResourceSchemaName)
}

type CreateResourceResult = dyn.OpResult[domain.Resource]

var deleteResourceCommandType = cqrs.RequestType{Module: "identity", Submodule: "resource", Action: "deleteResource"}

type DeleteResourceCommand dyn.DeleteOneCommand

func (DeleteResourceCommand) CqrsRequestType() cqrs.RequestType { return deleteResourceCommandType }

type DeleteResourceResult = dyn.OpResult[dyn.MutateResultData]

var getResourceQueryType = cqrs.RequestType{Module: "identity", Submodule: "resource", Action: "getResource"}

type GetResourceQuery dyn.GetOneQuery

func (GetResourceQuery) CqrsRequestType() cqrs.RequestType { return getResourceQueryType }

type GetResourceResult = dyn.OpResult[dyn.SingleResultData[domain.Resource]]

var resourceExistsQueryType = cqrs.RequestType{Module: "identity", Submodule: "resource", Action: "resourceExists"}

type ResourceExistsQuery dyn.ExistsQuery

func (ResourceExistsQuery) CqrsRequestType() cqrs.RequestType { return resourceExistsQueryType }

type ResourceExistsResult = dyn.OpResult[dyn.ExistsResultData]

var searchResourcesQueryType = cqrs.RequestType{Module: "identity", Submodule: "resource", Action: "searchResources"}

type SearchResourcesQuery dyn.SearchQuery

func (SearchResourcesQuery) CqrsRequestType() cqrs.RequestType { return searchResourcesQueryType }

type SearchResourcesResultData = dyn.PagedResultData[domain.Resource]
type SearchResourcesResult = dyn.OpResult[SearchResourcesResultData]

var updateResourceCommandType = cqrs.RequestType{Module: "identity", Submodule: "resource", Action: "updateResource"}

type UpdateResourceCommand struct {
	domain.Resource
}

func (UpdateResourceCommand) CqrsRequestType() cqrs.RequestType { return updateResourceCommandType }

func (UpdateResourceCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.ResourceSchemaName)
}

type UpdateResourceResult = dyn.OpResult[dyn.MutateResultData]
