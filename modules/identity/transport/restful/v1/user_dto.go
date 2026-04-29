package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

type CreateUserRequest = it.CreateUserCommand
type CreateUserResponse = httpserver.RestCreateResponse

type UpdateUserRequest = it.UpdateUserCommand
type UpdateUserResponse = httpserver.RestMutateResponse

type DeleteUserRequest = it.DeleteUserCommand
type DeleteUserResponse = httpserver.RestMutateResponse

type GetUserRequest = it.GetUserQuery
type GetUserResponse = httpserver.RestGetOneResponse[dmodel.DynamicFields]

type GetUserContextResponse struct {
	Id           string   `json:"id"`
	AvatarUrl    string   `json:"avatar_url"`
	DisplayName  string   `json:"display_name"`
	Email        string   `json:"email"`
	Entitlements []string `json:"entitlements"`
	OrgIds       []string `json:"org_ids"`
}

type UserExistsRequest = it.UserExistsQuery
type UserExistsResponse = dyn.ExistsResultData

type SearchUsersRequest = it.SearchUsersQuery
type SearchUsersResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]

type SetUserIsArchivedRequest = it.SetUserIsArchivedCommand
type SetUserIsArchivedResponse = httpserver.RestMutateResponse
