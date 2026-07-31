package role

import (
	"github.com/sky-as-code/nikki-erp/common/datastructure"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

func init() {
	var req cqrs.Request
	req = (*CreateRoleCommand)(nil)
	req = (*DeleteRoleCommand)(nil)
	req = (*GetRoleQuery)(nil)
	req = (*RoleExistsQuery)(nil)
	req = (*SearchRolesQuery)(nil)
	req = (*UpdateRoleCommand)(nil)
	req = (*SetRoleIsArchivedCommand)(nil)
	req = (*ManageRoleEntitlementsCommand)(nil)
	util.Unused(req)
}

var createRoleCommandType = cqrs.RequestType{Module: "iam", Submodule: "role", Action: "createRole"}

type CreateRoleCommand struct {
	domain.Role
}

func (CreateRoleCommand) CqrsRequestType() cqrs.RequestType { return createRoleCommandType }

func (CreateRoleCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.RoleSchemaName)
}

type CreateRoleResult = dyn.OpResult[domain.Role]

var deleteRoleCommandType = cqrs.RequestType{Module: "iam", Submodule: "role", Action: "deleteRole"}

type DeleteRoleCommand dyn.DeleteOneCommand

func (DeleteRoleCommand) CqrsRequestType() cqrs.RequestType { return deleteRoleCommandType }

type DeleteRoleResult = dyn.OpResult[dyn.MutateResultData]

var getRoleQueryType = cqrs.RequestType{Module: "iam", Submodule: "role", Action: "getRole"}

type GetRoleQuery dyn.GetOneQuery

func (GetRoleQuery) CqrsRequestType() cqrs.RequestType { return getRoleQueryType }

type GetRoleResult = dyn.OpResult[dyn.SingleResultData[domain.Role]]

var roleExistsQueryType = cqrs.RequestType{Module: "iam", Submodule: "role", Action: "roleExists"}

type RoleExistsQuery dyn.ExistsQuery

func (RoleExistsQuery) CqrsRequestType() cqrs.RequestType { return roleExistsQueryType }

type RoleExistsResult = dyn.OpResult[dyn.ExistsResultData]

var searchRolesQueryType = cqrs.RequestType{Module: "iam", Submodule: "role", Action: "searchRoles"}

type SearchRolesQuery dyn.SearchQuery

func (SearchRolesQuery) CqrsRequestType() cqrs.RequestType { return searchRolesQueryType }

type SearchRolesResultData = dyn.PagedResultData[domain.Role]
type SearchRolesResult = dyn.OpResult[SearchRolesResultData]

var updateRoleCommandType = cqrs.RequestType{Module: "iam", Submodule: "role", Action: "updateRole"}

type UpdateRoleCommand struct {
	domain.Role
}

func (UpdateRoleCommand) CqrsRequestType() cqrs.RequestType { return updateRoleCommandType }

func (UpdateRoleCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.RoleSchemaName)
}

type UpdateRoleResult = dyn.OpResult[dyn.MutateResultData]

var setRoleIsArchivedCommandType = cqrs.RequestType{Module: "iam", Submodule: "role", Action: "setRoleIsArchived"}

type SetRoleIsArchivedCommand dyn.SetIsArchivedCommand

func (SetRoleIsArchivedCommand) CqrsRequestType() cqrs.RequestType {
	return setRoleIsArchivedCommandType
}

type SetRoleIsArchivedResult = dyn.OpResult[dyn.MutateResultData]

var manageRoleEntitlementsCommandType = cqrs.RequestType{
	Module: "iam", Submodule: "role", Action: "manageRoleEntitlements",
}

type ManageRoleEntitlementsCommand struct {
	RoleId model.Id                    `json:"role_id" param:"role_id"`
	Add    datastructure.Set[model.Id] `json:"add"`
	Remove datastructure.Set[model.Id] `json:"remove"`
}

func (ManageRoleEntitlementsCommand) CqrsRequestType() cqrs.RequestType {
	return manageRoleEntitlementsCommandType
}

type ManageRoleEntitlementsResult = dyn.OpResult[dyn.MutateResultData]
