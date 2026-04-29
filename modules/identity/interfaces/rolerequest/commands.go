package role_request

import (
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	domain "github.com/sky-as-code/nikki-erp/modules/identity/domain/models"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

func init() {
	var req cqrs.Request
	req = (*CreateRoleRequestCommand)(nil)
	req = (*DeleteRoleRequestCommand)(nil)
	req = (*GetRoleRequestQuery)(nil)
	req = (*RoleRequestExistsQuery)(nil)
	req = (*SearchRoleRequestsQuery)(nil)
	req = (*UpdateRoleRequestCommand)(nil)
	util.Unused(req)
}

var createRoleRequestCommandType = cqrs.RequestType{
	Module: "identity", Submodule: "role_request", Action: "createRoleRequest",
}

type CreateRoleRequestCommand struct {
	domain.RoleRequest
}

func (CreateRoleRequestCommand) CqrsRequestType() cqrs.RequestType {
	return createRoleRequestCommandType
}

func (CreateRoleRequestCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.RoleRequestSchemaName)
}

type CreateRoleRequestResult = dyn.OpResult[domain.RoleRequest]

var deleteRoleRequestCommandType = cqrs.RequestType{
	Module: "identity", Submodule: "role_request", Action: "deleteRoleRequest",
}

type DeleteRoleRequestCommand dyn.DeleteOneCommand

func (DeleteRoleRequestCommand) CqrsRequestType() cqrs.RequestType {
	return deleteRoleRequestCommandType
}

type DeleteRoleRequestResult = dyn.OpResult[dyn.MutateResultData]

var getRoleRequestQueryType = cqrs.RequestType{
	Module: "identity", Submodule: "role_request", Action: "getRoleRequest",
}

type GetRoleRequestQuery dyn.GetOneQuery

func (GetRoleRequestQuery) CqrsRequestType() cqrs.RequestType { return getRoleRequestQueryType }

type GetRoleRequestResult = dyn.OpResult[dyn.SingleResultData[domain.RoleRequest]]

var roleRequestExistsQueryType = cqrs.RequestType{
	Module: "identity", Submodule: "role_request", Action: "roleRequestExists",
}

type RoleRequestExistsQuery dyn.ExistsQuery

func (RoleRequestExistsQuery) CqrsRequestType() cqrs.RequestType { return roleRequestExistsQueryType }

type RoleRequestExistsResult = dyn.OpResult[dyn.ExistsResultData]

var searchRoleRequestsQueryType = cqrs.RequestType{
	Module: "identity", Submodule: "role_request", Action: "searchRoleRequests",
}

type SearchRoleRequestsQuery dyn.SearchQuery

func (SearchRoleRequestsQuery) CqrsRequestType() cqrs.RequestType { return searchRoleRequestsQueryType }

type SearchRoleRequestsResultData = dyn.PagedResultData[domain.RoleRequest]
type SearchRoleRequestsResult = dyn.OpResult[SearchRoleRequestsResultData]

var updateRoleRequestCommandType = cqrs.RequestType{
	Module: "identity", Submodule: "role_request", Action: "updateRoleRequest",
}

type UpdateRoleRequestCommand struct {
	domain.RoleRequest
}

func (UpdateRoleRequestCommand) CqrsRequestType() cqrs.RequestType {
	return updateRoleRequestCommandType
}

func (UpdateRoleRequestCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.RoleRequestSchemaName)
}

type UpdateRoleRequestResult = dyn.OpResult[dyn.MutateResultData]
