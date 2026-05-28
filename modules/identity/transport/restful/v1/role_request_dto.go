package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/rolerequest"
)

type CreateRoleRequestRequest = it.CreateRoleRequestCommand
type CreateRoleRequestResponse = httpserver.RestCreateResponse

type DeleteRoleRequestRequest = it.DeleteRoleRequestCommand
type DeleteRoleRequestResponse = httpserver.RestMutateResponse

type GetRoleRequestRequest = it.GetRoleRequestQuery
type GetRoleRequestResponse = httpserver.RestGetOneResponse[dmodel.DynamicFields]

type RoleRequestExistsRequest = it.RoleRequestExistsQuery
type RoleRequestExistsResponse = dyn.ExistsResultData

type SearchRoleRequestsRequest = it.SearchRoleRequestsQuery
type SearchRoleRequestsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]

type UpdateRoleRequestRequest = it.UpdateRoleRequestCommand
type UpdateRoleRequestResponse = httpserver.RestMutateResponse
