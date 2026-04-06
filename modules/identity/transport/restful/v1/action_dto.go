package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/action"
)

type CreateActionRequest = it.CreateActionCommand
type CreateActionResponse = httpserver.RestCreateResponse

type DeleteActionRequest = it.DeleteActionCommand
type DeleteActionResponse = httpserver.RestDeleteResponse2

type GetActionRequest = it.GetActionQuery
type GetActionResponse = dmodel.DynamicFields

type ActionExistsRequest = it.ActionExistsQuery
type ActionExistsResponse = dyn.ExistsResultData

type SearchActionsRequest = it.SearchActionsQuery
type SearchActionsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]

type UpdateActionRequest = it.UpdateActionCommand
type UpdateActionResponse = httpserver.RestUpdateResponse2
