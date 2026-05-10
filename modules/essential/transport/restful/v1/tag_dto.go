package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/tag"
)

type CreateTagRequest struct {
	dmodel.DynamicFields
}
type CreateTagResponse = httpserver.RestCreateResponse

type UpdateTagRequest struct {
	dmodel.DynamicFields
	TagId string `param:"id"`
}
type UpdateTagResponse = httpserver.RestMutateResponse

type DeleteTagRequest = it.DeleteTagCommand
type DeleteTagResponse = httpserver.RestDeleteResponse2

type GetTagRequest = it.GetTagQuery
type GetTagResponse = dmodel.DynamicFields

type TagExistsRequest = it.TagExistsQuery
type TagExistsResponse = dyn.ExistsResultData

type SearchTagsRequest = it.SearchTagsQuery
type SearchTagsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
