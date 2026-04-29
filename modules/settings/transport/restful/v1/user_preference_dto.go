package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/userpref"
)

type CreateUserPreferenceRequest = it.CreateUserPreferenceCommand
type CreateUserPreferenceResponse = httpserver.RestCreateResponse

type DeleteUserPreferenceRequest = it.DeleteUserPreferenceCommand
type DeleteUserPreferenceResponse = httpserver.RestDeleteResponse2

type GetUserPreferenceRequest = it.GetUserPreferenceQuery
type GetUserPreferenceResponse = dmodel.DynamicFields

type UserPreferenceExistsRequest = it.UserPreferenceExistsQuery
type UserPreferenceExistsResponse = dyn.ExistsResultData

type SearchUserPreferencesRequest = it.SearchUserPreferencesQuery
type SearchUserPreferencesResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]

type UpdateUserPreferenceRequest = it.UpdateUserPreferenceCommand
type UpdateUserPreferenceResponse = httpserver.RestMutateResponse
