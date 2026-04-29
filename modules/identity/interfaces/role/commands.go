package role

import (
	"github.com/sky-as-code/nikki-erp/common/datastructure"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	domain "github.com/sky-as-code/nikki-erp/modules/identity/domain/models"

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
	req = (*DeletePrivateRoleCommand)(nil)
	util.Unused(req)
}

var createRoleCommandType = cqrs.RequestType{Module: "identity", Submodule: "role", Action: "createRole"}

type CreateRoleCommand struct {
	domain.Role
}

func (CreateRoleCommand) CqrsRequestType() cqrs.RequestType { return createRoleCommandType }

func (CreateRoleCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.RoleSchemaName)
}

var createPrivateRoleCommandType = cqrs.RequestType{
	Module: "identity", Submodule: "role", Action: "createPrivateRole",
}

type CreatePrivateRoleCommand struct {
	OwnerId   model.Id `json:"owner_id" param:"owner_id"`
	OwnerType string   `json:"owner_type" param:"owner_type"`
}

func (CreatePrivateRoleCommand) CqrsRequestType() cqrs.RequestType {
	return createPrivateRoleCommandType
}

func (CreatePrivateRoleCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"identity.create_private_role_command",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				Field(dmodel.DefineField().
					Name("owner_id").
					DataType(dmodel.FieldDataTypeUlid()).
					RequiredAlways()).
				Field(dmodel.DefineField().
					Name("owner_type").
					DataType(dmodel.FieldDataTypeEnumString([]string{"group", "user"})).
					RequiredAlways())
		},
	)
}

type CreateRoleResult = dyn.OpResult[domain.Role]

var deleteRoleCommandType = cqrs.RequestType{Module: "identity", Submodule: "role", Action: "deleteRole"}

type DeleteRoleCommand dyn.DeleteOneCommand

func (DeleteRoleCommand) CqrsRequestType() cqrs.RequestType { return deleteRoleCommandType }

var deletePrivateRoleCommandType = cqrs.RequestType{
	Module: "identity", Submodule: "role", Action: "deletePrivateRole",
}

type DeletePrivateRoleCommand struct {
	OwnerId model.Id `json:"owner_id" param:"owner_id"`
}

func (DeletePrivateRoleCommand) CqrsRequestType() cqrs.RequestType {
	return deletePrivateRoleCommandType
}

func (DeletePrivateRoleCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"identity.delete_private_role_command",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				Field(dmodel.DefineField().
					Name("owner_id").
					DataType(dmodel.FieldDataTypeUlid()).
					RequiredAlways())
		},
	)
}

type DeleteRoleResult = dyn.OpResult[dyn.MutateResultData]

var getRoleQueryType = cqrs.RequestType{Module: "identity", Submodule: "role", Action: "getRole"}

type GetRoleQuery dyn.GetOneQuery

func (GetRoleQuery) CqrsRequestType() cqrs.RequestType { return getRoleQueryType }

type GetRoleResult = dyn.OpResult[dyn.SingleResultData[domain.Role]]

var roleExistsQueryType = cqrs.RequestType{Module: "identity", Submodule: "role", Action: "roleExists"}

type RoleExistsQuery dyn.ExistsQuery

func (RoleExistsQuery) CqrsRequestType() cqrs.RequestType { return roleExistsQueryType }

type RoleExistsResult = dyn.OpResult[dyn.ExistsResultData]

var searchRolesQueryType = cqrs.RequestType{Module: "identity", Submodule: "role", Action: "searchRoles"}

type SearchRolesQuery dyn.SearchQuery

func (SearchRolesQuery) CqrsRequestType() cqrs.RequestType { return searchRolesQueryType }

type SearchRolesResultData = dyn.PagedResultData[domain.Role]
type SearchRolesResult = dyn.OpResult[SearchRolesResultData]

var updateRoleCommandType = cqrs.RequestType{Module: "identity", Submodule: "role", Action: "updateRole"}

type UpdateRoleCommand struct {
	domain.Role
}

func (UpdateRoleCommand) CqrsRequestType() cqrs.RequestType { return updateRoleCommandType }

func (UpdateRoleCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.RoleSchemaName)
}

type UpdateRoleResult = dyn.OpResult[dyn.MutateResultData]

var setRoleIsArchivedCommandType = cqrs.RequestType{Module: "identity", Submodule: "role", Action: "setRoleIsArchived"}

type SetRoleIsArchivedCommand dyn.SetIsArchivedCommand

func (SetRoleIsArchivedCommand) CqrsRequestType() cqrs.RequestType {
	return setRoleIsArchivedCommandType
}

type SetRoleIsArchivedResult = dyn.OpResult[dyn.MutateResultData]

var manageRoleEntitlementsCommandType = cqrs.RequestType{
	Module: "identity", Submodule: "role", Action: "manageRoleEntitlements",
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

var manageRoleAssignmentsCommandType = cqrs.RequestType{
	Module: "identity", Submodule: "role", Action: "manageRoleAssignments",
}

type ManageRoleAssignmentsCommand struct {
	RoleId    model.Id                    `json:"role_id" param:"role_id"`
	Add       datastructure.Set[model.Id] `json:"add"`
	Remove    datastructure.Set[model.Id] `json:"remove"`
	OwnerType string                      `json:"owner_type" param:"owner_type"`
}

func (ManageRoleAssignmentsCommand) CqrsRequestType() cqrs.RequestType {
	return manageRoleAssignmentsCommandType
}

func (ManageRoleAssignmentsCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"identity.manage_role_assignments_command",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				Field(dmodel.DefineField().
					Name("owner_type").
					DataType(dmodel.FieldDataTypeEnumString([]string{"group", "user"})).
					RequiredAlways())
		},
	)
}

type ManageRoleAssignmentsResult = dyn.OpResult[dyn.MutateResultData]
