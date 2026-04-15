package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/requestforquote"
)

type CreateRequestForQuoteRequest struct{ dmodel.DynamicFields }
type CreateRequestForQuoteResponse = httpserver.RestCreateResponse
type DeleteRequestForQuoteRequest = it.DeleteRequestForQuoteCommand
type DeleteRequestForQuoteResponse = httpserver.RestDeleteResponse2
type GetRequestForQuoteRequest = it.GetRequestForQuoteQuery
type GetRequestForQuoteResponse = dmodel.DynamicFields
type RequestForQuoteExistsRequest = it.RequestForQuoteExistsQuery
type RequestForQuoteExistsResponse = dyn.ExistsResultData
type SearchRequestForQuotesRequest = it.SearchRequestForQuotesQuery
type SearchRequestForQuotesResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
type SetRequestForQuoteIsArchivedRequest = it.SetRequestForQuoteIsArchivedCommand
type SetRequestForQuoteIsArchivedResponse = httpserver.RestMutateResponse
type UpdateRequestForQuoteRequest struct {
	dmodel.DynamicFields
	RequestForQuoteId string `param:"id"`
}
type UpdateRequestForQuoteResponse = httpserver.RestMutateResponse
