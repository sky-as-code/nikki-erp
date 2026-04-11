package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/role"
)

type CreateRoleRequest = it.CreateRoleCommand
type CreateRoleResponse = httpserver.RestCreateResponse

type DeleteRoleRequest = it.DeleteRoleCommand
type DeleteRoleResponse = httpserver.RestDeleteResponse2

type GetRoleRequest = it.GetRoleQuery
type GetRoleResponse = dmodel.DynamicFields

type ManageRoleEntitlementsRequest = it.ManageRoleEntitlementsCommand
type ManageRoleEntitlementsResponse = httpserver.RestMutateResponse

type RoleExistsRequest = it.RoleExistsQuery
type RoleExistsResponse = dyn.ExistsResultData

type SearchRolesRequest = it.SearchRolesQuery
type SearchRolesResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]

type SetRoleIsArchivedRequest = it.SetRoleIsArchivedCommand
type SetRoleIsArchivedResponse = httpserver.RestMutateResponse

type UpdateRoleRequest = it.UpdateRoleCommand
type UpdateRoleResponse = httpserver.RestMutateResponse
