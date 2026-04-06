package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/entitlement"
)

type CreateEntitlementRequest = it.CreateEntitlementCommand
type CreateEntitlementResponse = httpserver.RestCreateResponse

type DeleteEntitlementRequest = it.DeleteEntitlementCommand
type DeleteEntitlementResponse = httpserver.RestDeleteResponse2

type GetEntitlementRequest = it.GetEntitlementQuery
type GetEntitlementResponse = dmodel.DynamicFields

type EntitlementExistsRequest = it.EntitlementExistsQuery
type EntitlementExistsResponse = dyn.ExistsResultData

type ManageEntitlementRolesRequest = it.ManageEntitlementRolesCommand
type ManageEntitlementRolesResponse = httpserver.RestMutateResponse

type SearchEntitlementsRequest = it.SearchEntitlementsQuery
type SearchEntitlementsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]

type SetEntitlementIsArchivedRequest = it.SetEntitlementIsArchivedCommand
type SetEntitlementIsArchivedResponse = httpserver.RestUpdateResponse2

type UpdateEntitlementRequest = it.UpdateEntitlementCommand
type UpdateEntitlementResponse = httpserver.RestUpdateResponse2
