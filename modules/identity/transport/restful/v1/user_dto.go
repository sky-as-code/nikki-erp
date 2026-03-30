package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

type SetUserIsArchivedRequest = it.SetUserIsArchivedCommand
type SetUserIsArchivedResponse = httpserver.RestUpdateResponse2

type CreateUserRequest = it.CreateUserCommand
type CreateUserResponse = httpserver.RestCreateResponse

type UpdateUserRequest = it.UpdateUserCommand
type UpdateUserResponse = httpserver.RestUpdateResponse2

type DeleteUserRequest = it.DeleteUserCommand
type DeleteUserResponse = httpserver.RestDeleteResponse2

type GetUserRequest = it.GetUserQuery
type GetUserResponse = dmodel.DynamicFields

// type GetUserContextRequest = it.GetUserContextQuery
// type GetUserContextResponse = it.GetUserContextResult

type UserExistsRequest = it.UserExistsQuery
type UserExistsResponse = dyn.ExistsResultData

type SearchUsers2Request = it.SearchUsersQuery
type SearchUsersResponse2 = httpserver.RestSearchResponse[dmodel.DynamicFields]
