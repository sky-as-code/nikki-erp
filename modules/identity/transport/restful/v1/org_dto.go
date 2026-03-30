package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
)

type CreateOrgRequest = it.CreateOrgCommand
type CreateOrgResponse = httpserver.RestCreateResponse

type DeleteOrgRequest = it.DeleteOrgCommand
type DeleteOrgResponse = httpserver.RestDeleteResponse2

type GetOrgRequest = it.GetOrgQuery
type GetOrgResponse = dmodel.DynamicFields

type OrgExistsRequest = it.OrgExistsQuery
type OrgExistsResponse = dyn.ExistsResultData

type ManageOrgUsersRequest = it.ManageOrgUsersCommand
type ManageOrgsResponse = httpserver.RestMutateResponse

type SearchOrgsRequest = it.SearchOrgsQuery
type SearchOrgsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]

type SetOrgIsArchivedRequest = it.SetOrgIsArchivedCommand
type SetOrgIsArchivedResponse = httpserver.RestUpdateResponse2

type UpdateOrgRequest = it.UpdateOrgCommand
type UpdateOrgResponse = httpserver.RestUpdateResponse2
