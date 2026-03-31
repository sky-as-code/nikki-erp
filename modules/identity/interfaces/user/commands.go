package user

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/crud"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

func init() {
	// Assert interface implementation
	var req cqrs.Request
	req = (*CreateUserCommand)(nil)
	req = (*DeleteUserCommand)(nil)
	req = (*GetUserQuery)(nil)
	req = (*SearchUsersQuery)(nil)
	req = (*UpdateUserCommand)(nil)
	req = (*UserExistsQuery)(nil)
	req = (*UserExistsMultiQuery)(nil)
	util.Unused(req)
}

var createUserCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "create",
}

type CreateUserCommand struct {
	domain.User
}

func (CreateUserCommand) CqrsRequestType() cqrs.RequestType {
	return createUserCommandType
}

func (this CreateUserCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.UserSchemaName)
}

type CreateUserResult = dyn.OpResult[domain.User]

var deleteUserCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "delete",
}

type DeleteUserCommand dyn.DeleteOneCommand

func (DeleteUserCommand) CqrsRequestType() cqrs.RequestType {
	return deleteUserCommandType
}

type DeleteUserResult = dyn.OpResult[dyn.MutateResultData]

var getUserByIdQueryType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "getUserById",
}

type GetUserQuery struct {
	Columns []string `json:"columns" query:"columns"`
	Id      *string  `json:"id" param:"id"`
	Email   *string  `json:"email"`
}

func (GetUserQuery) CqrsRequestType() cqrs.RequestType {
	return getUserByIdQueryType
}

type GetUserResult = dyn.OpResult[domain.User]

var getUserContextQueryType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "getUserContext",
}

type GetUserContextQuery struct {
	UserId model.Id `json:"id" param:"id"`
}

func (GetUserContextQuery) CqrsRequestType() cqrs.RequestType {
	return getUserContextQueryType
}

func (this GetUserContextQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.UserId, true),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

// type GetUserContextResult struct {
// 	User        *domain.User                                       `json:"user,omitempty"`
// 	Hierachies  []domain.HierarchyLevel                            `json:"hierarchies,omitempty"`
// 	Orgs        []domain.Organization                              `json:"orgs,omitempty"`
// 	Permissions *map[string][]itAuthorize.ResourceScopePermissions `json:"permissions,omitempty"`
// }

// type GetUserContextResultData = corecrud.OpResult[*GetUserContextResult]

var searchUsersQueryType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "search",
}

type SearchUsersQuery dyn.SearchQuery

func (SearchUsersQuery) CqrsRequestType() cqrs.RequestType {
	return searchUsersQueryType
}

type SearchUsersResultData = dyn.PagedResultData[domain.User]
type SearchUsersResult = dyn.OpResult[SearchUsersResultData]

var setUserIsArchivedCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "setUserIsArchived",
}

type SetUserIsArchivedCommand dyn.SetIsArchivedCommand

func (SetUserIsArchivedCommand) CqrsRequestType() cqrs.RequestType {
	return setUserIsArchivedCommandType
}

type SetUserIsArchivedResult = dyn.OpResult[dyn.MutateResultData]

var updateUserCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "update",
}

type UpdateUserCommand struct {
	domain.User
}

func (UpdateUserCommand) CqrsRequestType() cqrs.RequestType {
	return updateUserCommandType
}

func (this UpdateUserCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.UserSchemaName)
}

type UpdateUserResult = dyn.OpResult[dyn.MutateResultData]

var existsCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "exists",
}

type UserExistsQuery dyn.ExistsQuery

func (UserExistsQuery) CqrsRequestType() cqrs.RequestType {
	return existsCommandType
}

type UserExistsResult = dyn.OpResult[dyn.ExistsResultData]

var userExistsMultiCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "existsMulti",
}

type UserExistsMultiQuery struct {
	Ids   []model.Id `json:"ids"`
	OrgId *model.Id  `json:"org_id,omitempty"`
}

func (UserExistsMultiQuery) CqrsRequestType() cqrs.RequestType {
	return userExistsMultiCommandType
}

type UserExistsMultiResult = corecrud.OpResult[*dyn.ExistsResultData]
