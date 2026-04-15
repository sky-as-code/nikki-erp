package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/requestforproposal"
)

type CreateRequestForProposalRequest struct{ dmodel.DynamicFields }
type CreateRequestForProposalResponse = httpserver.RestCreateResponse
type DeleteRequestForProposalRequest = it.DeleteRequestForProposalCommand
type DeleteRequestForProposalResponse = httpserver.RestDeleteResponse2
type GetRequestForProposalRequest = it.GetRequestForProposalQuery
type GetRequestForProposalResponse = dmodel.DynamicFields
type RequestForProposalExistsRequest = it.RequestForProposalExistsQuery
type RequestForProposalExistsResponse = dyn.ExistsResultData
type SearchRequestForProposalsRequest = it.SearchRequestForProposalsQuery
type SearchRequestForProposalsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
type SetRequestForProposalIsArchivedRequest = it.SetRequestForProposalIsArchivedCommand
type SetRequestForProposalIsArchivedResponse = httpserver.RestMutateResponse
type UpdateRequestForProposalRequest struct {
	dmodel.DynamicFields
	RequestForProposalId string `param:"id"`
}
type UpdateRequestForProposalResponse = httpserver.RestMutateResponse
