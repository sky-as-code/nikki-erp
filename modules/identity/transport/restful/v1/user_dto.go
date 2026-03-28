package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

type SetUserIsArchivedRequest = it.SetUserIsArchived
type SetUserIsArchivedResponse = httpserver.RestUpdateResponse2

type CreateUserRequest = it.CreateUserCommand
type CreateUserResponse = httpserver.RestCreateResponse

type UpdateUserRequest = it.UpdateUserCommand
type UpdateUserResponse = httpserver.RestUpdateResponse2

type DeleteUserRequest = it.DeleteUserCommand
type DeleteUserResponse = httpserver.RestDeleteResponse2

type GetUserRequest = it.GetUserQuery
type GetUserResponse = dmodel.DynamicFields

type GetUserContextRequest = it.GetUserContextQuery
type GetUserContextResponse = it.GetUserContextResult

type UserExistsMultiRequest = it.UserExistsMultiQuery
type UserExistsMultiResponse = it.ExistsMultiResultData

type SearchUsersResponse2 = httpserver.RestSearchResponse[dmodel.DynamicFields]

type SearchUsers2Request = it.SearchUsersQuery2
