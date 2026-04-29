package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/action"
)

type CreateActionRequest struct {
	dmodel.DynamicFields
	ResourceId string `param:"resource_id"`
}
type CreateActionResponse = httpserver.RestCreateResponse

type DeleteActionRequest = it.DeleteActionCommand
type DeleteActionResponse = httpserver.RestMutateResponse

type GetActionRequest = it.GetActionQuery
type GetActionResponse = httpserver.RestGetOneResponse[dmodel.DynamicFields]

type ActionExistsRequest = it.ActionExistsQuery
type ActionExistsResponse = dyn.ExistsResultData

type SearchActionsRequest = it.SearchActionsQuery
type SearchActionsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]

type UpdateActionRequest struct {
	dmodel.DynamicFields
	ActionId   string `param:"action_id"`
	ResourceId string `param:"resource_id"`
}
type UpdateActionResponse = httpserver.RestMutateResponse
