package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/resource"
)

type CreateResourceRequest = it.CreateResourceCommand
type CreateResourceResponse = httpserver.RestCreateResponse

type DeleteResourceRequest = it.DeleteResourceCommand
type DeleteResourceResponse = httpserver.RestDeleteResponse2

type GetResourceRequest = it.GetResourceQuery
type GetResourceResponse = dmodel.DynamicFields

type ResourceExistsRequest = it.ResourceExistsQuery
type ResourceExistsResponse = dyn.ExistsResultData

type SearchResourcesRequest = it.SearchResourcesQuery
type SearchResourcesResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]

type UpdateResourceRequest = it.UpdateResourceCommand
type UpdateResourceResponse = httpserver.RestUpdateResponse2
