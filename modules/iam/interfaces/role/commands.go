package role

import (
	"github.com/sky-as-code/nikki-erp/common/datastructure"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
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
	req = (*SearchUserRolesQuery)(nil)
	req = (*SearchGroupRolesQuery)(nil)
	req = (*DescribeRolesQuery)(nil)
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

var searchUserRolesQueryType = cqrs.RequestType{Module: "iam", Submodule: "role", Action: "searchUserRoles"}

// SearchUserRolesQuery searches roles assigned to a single user. The user id comes from the
// route, the caller may still narrow the result with its own graph, which is ANDed with the
// assignment condition.
type SearchUserRolesQuery struct {
	Fields []string            `json:"fields" query:"fields"`
	Graph  *dmodel.SearchGraph `json:"graph" query:"graph"`
	Page   int                 `json:"page" query:"page"`
	Size   int                 `json:"size" query:"size"`
	UserId model.Id            `json:"user_id" param:"user_id"`
}

func (SearchUserRolesQuery) CqrsRequestType() cqrs.RequestType { return searchUserRolesQueryType }

func (SearchUserRolesQuery) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"iam.search_user_roles_query",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				Field(dyn.DefineFieldSearchFields()).
				Field(dyn.DefineFieldSearchGraph()).
				Field(dyn.DefineFieldSearchPage()).
				Field(dyn.DefineFieldSearchSize()).
				Field(basemodel.DefineFieldId("user_id").
					RequiredAlways())
		},
	)
}

type SearchUserRolesResult = dyn.OpResult[SearchRolesResultData]

var searchGroupRolesQueryType = cqrs.RequestType{Module: "iam", Submodule: "role", Action: "searchGroupRoles"}

// SearchGroupRolesQuery is the group counterpart of SearchUserRolesQuery.
type SearchGroupRolesQuery struct {
	Fields  []string            `json:"fields" query:"fields"`
	Graph   *dmodel.SearchGraph `json:"graph" query:"graph"`
	Page    int                 `json:"page" query:"page"`
	Size    int                 `json:"size" query:"size"`
	GroupId model.Id            `json:"group_id" param:"group_id"`
}

func (SearchGroupRolesQuery) CqrsRequestType() cqrs.RequestType { return searchGroupRolesQueryType }

func (SearchGroupRolesQuery) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"iam.search_group_roles_query",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				Field(dyn.DefineFieldSearchFields()).
				Field(dyn.DefineFieldSearchGraph()).
				Field(dyn.DefineFieldSearchPage()).
				Field(dyn.DefineFieldSearchSize()).
				Field(basemodel.DefineFieldId("group_id").
					RequiredAlways())
		},
	)
}

type SearchGroupRolesResult = dyn.OpResult[SearchRolesResultData]

// DescribeRolesMaxIds caps how many roles a single describe call may resolve. Each role fans
// out into an entitlement search plus scope-name lookups, so an unbounded list would let one
// request do arbitrary work.
const DescribeRolesMaxIds = 20

var describeRolesQueryType = cqrs.RequestType{Module: "iam", Submodule: "role", Action: "describeRoles"}

// DescribeRolesQuery resolves roles into a human-presentable shape: every entitlement carries
// its resource name, action name and scope name rather than the raw expression triple.
type DescribeRolesQuery struct {
	RoleIds []model.Id `json:"role_ids" query:"role_id"`
}

func (DescribeRolesQuery) CqrsRequestType() cqrs.RequestType { return describeRolesQueryType }

func (DescribeRolesQuery) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"iam.describe_roles_query",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				Field(basemodel.DefineFieldIdArr("role_ids").
					Rule(dmodel.FieldRuleArrayLength(0, DescribeRolesMaxIds)))
		},
	)
}

// DescribedEntitlement is one entitlement with every id already resolved to a display name.
// ScopeId and ScopeName are only set for the org and orgunit scopes; domain and private
// entitlements carry no target, and the caller labels them from a static translation key.
type DescribedEntitlement struct {
	Id           model.Id  `json:"id"`
	ResourceId   *model.Id `json:"resource_id"`
	ResourceName *string   `json:"resource_name"`
	ActionId     *model.Id `json:"action_id"`
	ActionName   *string   `json:"action_name"`
	Scope        *string   `json:"scope"`
	ScopeId      *model.Id `json:"scope_id"`
	ScopeName    *string   `json:"scope_name"`
}

type DescribedRole struct {
	Id           model.Id               `json:"id"`
	Name         *string                `json:"name"`
	Entitlements []DescribedEntitlement `json:"entitlements"`
}

type DescribeRolesResultData struct {
	Items []DescribedRole `json:"items"`
}

type DescribeRolesResult = dyn.OpResult[DescribeRolesResultData]

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
