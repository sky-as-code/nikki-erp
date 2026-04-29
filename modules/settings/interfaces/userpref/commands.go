package userpreference

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	domain "github.com/sky-as-code/nikki-erp/modules/settings/domain/models"
)

func init() {
	var req cqrs.Request
	req = (*CreateUserPreferenceCommand)(nil)
	req = (*DeleteUserPreferenceCommand)(nil)
	req = (*GetUserPreferenceQuery)(nil)
	req = (*UserPreferenceExistsQuery)(nil)
	req = (*SearchUserPreferencesQuery)(nil)
	req = (*UpdateUserPreferenceCommand)(nil)
	util.Unused(req)
}

var createUserPreferenceCommandType = cqrs.RequestType{Module: "settings", Submodule: "userPreference", Action: "create"}

type CreateUserPreferenceCommand struct {
	domain.UserPreference
}

func (CreateUserPreferenceCommand) CqrsRequestType() cqrs.RequestType {
	return createUserPreferenceCommandType
}

func (CreateUserPreferenceCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.UserPreferenceSchemaName)
}

type CreateUserPreferenceResult = dyn.OpResult[domain.UserPreference]

var deleteUserPreferenceCommandType = cqrs.RequestType{Module: "settings", Submodule: "userPreference", Action: "delete"}

type DeleteUserPreferenceCommand dyn.DeleteOneCommand

func (DeleteUserPreferenceCommand) CqrsRequestType() cqrs.RequestType {
	return deleteUserPreferenceCommandType
}

type DeleteUserPreferenceResult = dyn.OpResult[dyn.MutateResultData]

var getUserPreferenceQueryType = cqrs.RequestType{Module: "settings", Submodule: "userPreference", Action: "getOne"}

type GetUserPreferenceQuery dyn.GetOneQuery

func (GetUserPreferenceQuery) CqrsRequestType() cqrs.RequestType { return getUserPreferenceQueryType }

type GetUserPreferenceResult = dyn.OpResult[domain.UserPreference]

type GetUiSavedSearchQuery struct {
	SearchName string   `json:"search_name"`
	UserId     model.Id `json:"user_id" query:"user_id"`
}

type GetUiSavedSearchResultData struct {
	Fields []string            `json:"fields"`
	Graph  *dmodel.SearchGraph `json:"graph"`
}

type GetUiSavedSearchResult = dyn.OpResult[GetUiSavedSearchResultData]

var userPreferenceExistsQueryType = cqrs.RequestType{Module: "settings", Submodule: "userPreference", Action: "exists"}

type UserPreferenceExistsQuery dyn.ExistsQuery

func (UserPreferenceExistsQuery) CqrsRequestType() cqrs.RequestType {
	return userPreferenceExistsQueryType
}

type UserPreferenceExistsResult = dyn.OpResult[dyn.ExistsResultData]

var searchUserPreferencesQueryType = cqrs.RequestType{Module: "settings", Submodule: "userPreference", Action: "search"}

type SearchUserPreferencesQuery dyn.SearchQuery

func (SearchUserPreferencesQuery) CqrsRequestType() cqrs.RequestType {
	return searchUserPreferencesQueryType
}

type SearchUserPreferencesResultData = dyn.PagedResultData[domain.UserPreference]
type SearchUserPreferencesResult = dyn.OpResult[SearchUserPreferencesResultData]

var updateUserPreferenceCommandType = cqrs.RequestType{Module: "settings", Submodule: "userPreference", Action: "update"}

type UpdateUserPreferenceCommand struct {
	domain.UserPreference
}

func (UpdateUserPreferenceCommand) CqrsRequestType() cqrs.RequestType {
	return updateUserPreferenceCommandType
}

func (UpdateUserPreferenceCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.UserPreferenceSchemaName)
}

type UpdateUserPreferenceResult = dyn.OpResult[dyn.MutateResultData]
