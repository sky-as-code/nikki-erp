package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/group"
)

type CreateGroupRequest = it.CreateGroupCommand
type CreateGroupResponse = httpserver.RestCreateResponse

type DeleteGroupRequest = it.DeleteGroupCommand
type DeleteGroupResponse = httpserver.RestDeleteResponse2

type GetGroupRequest = it.GetGroupQuery
type GetGroupResponse = dmodel.DynamicFields

type GroupExistsRequest = it.GroupExistsQuery
type GroupExistsResponse = dyn.ExistsResultData

type ManageGroupUsersRequest = it.ManageGroupUsersCommand
type ManageGroupUsersResponse = httpserver.RestMutateResponse

type SearchGroupsRequest = it.SearchGroupsQuery
type SearchGroupsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]

type UpdateGroupRequest = it.UpdateGroupCommand
type UpdateGroupResponse = httpserver.RestUpdateResponse2
