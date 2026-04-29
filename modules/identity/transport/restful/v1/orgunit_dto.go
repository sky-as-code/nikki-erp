package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/orgunit"
)

type CreateOrgUnitRequest = it.CreateOrgUnitCommand
type CreateOrgUnitResponse = httpserver.RestCreateResponse

type DeleteOrgUnitRequest = it.DeleteOrgUnitCommand
type DeleteOrgUnitResponse = httpserver.RestMutateResponse

type GetOrgUnitRequest = it.GetOrgUnitQuery
type GetOrgUnitResponse = httpserver.RestGetOneResponse[dmodel.DynamicFields]

type OrgUnitExistsRequest = it.OrgUnitExistsQuery
type OrgUnitExistsResponse = dyn.ExistsResultData

type ManageOrgUnitUsersRequest = it.ManageOrgUnitUsersCommand
type ManageOrgUnitUsersResponse = httpserver.RestMutateResponse

type SearchOrgUnitsRequest = it.SearchOrgUnitsQuery
type SearchOrgUnitsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]

type UpdateOrgUnitRequest = it.UpdateOrgUnitCommand
type UpdateOrgUnitResponse = httpserver.RestMutateResponse
