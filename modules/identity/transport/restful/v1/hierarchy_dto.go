package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/hierarchy"
)

type CreateHierarchyLevelRequest = it.CreateHierarchyLevelCommand
type CreateHierarchyLevelResponse = httpserver.RestCreateResponse

type DeleteHierarchyLevelRequest = it.DeleteHierarchyLevelCommand
type DeleteHierarchyLevelResponse = httpserver.RestDeleteResponse2

type GetHierarchyLevelRequest = it.GetHierarchyLevelQuery
type GetHierarchyLevelResponse = dmodel.DynamicFields

type HierarchyLevelExistsRequest = it.HierarchyLevelExistsQuery
type HierarchyLevelExistsResponse = dyn.ExistsResultData

type ManageHierarchyLevelUsersRequest = it.ManageHierarchyLevelUsersCommand
type ManageHierarchyLevelUsersResponse = httpserver.RestMutateResponse

type SearchHierarchyLevelsRequest = it.SearchHierarchyLevelsQuery
type SearchHierarchyLevelsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]

type UpdateHierarchyLevelRequest = it.UpdateHierarchyLevelCommand
type UpdateHierarchyLevelResponse = httpserver.RestUpdateResponse2
