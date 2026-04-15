package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/purchaserequest"
)

type CreatePurchaseRequestRequest struct{ dmodel.DynamicFields }
type CreatePurchaseRequestResponse = httpserver.RestCreateResponse

type DeletePurchaseRequestRequest = it.DeletePurchaseRequestCommand
type DeletePurchaseRequestResponse = httpserver.RestDeleteResponse2

type GetPurchaseRequestRequest = it.GetPurchaseRequestQuery
type GetPurchaseRequestResponse = dmodel.DynamicFields

type PurchaseRequestExistsRequest = it.PurchaseRequestExistsQuery
type PurchaseRequestExistsResponse = dyn.ExistsResultData

type SearchPurchaseRequestsRequest = it.SearchPurchaseRequestsQuery
type SearchPurchaseRequestsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]

type SetPurchaseRequestIsArchivedRequest = it.SetPurchaseRequestIsArchivedCommand
type SetPurchaseRequestIsArchivedResponse = httpserver.RestMutateResponse

type UpdatePurchaseRequestRequest struct {
	dmodel.DynamicFields
	PurchaseRequestId string `param:"id"`
}

type UpdatePurchaseRequestResponse = httpserver.RestMutateResponse

type SubmitPurchaseRequestForApprovalRequest = it.SubmitPurchaseRequestForApprovalCommand
type SubmitPurchaseRequestForApprovalResponse = httpserver.RestMutateResponse

type ApprovePurchaseRequestRequest = it.ApprovePurchaseRequestCommand
type ApprovePurchaseRequestResponse = httpserver.RestMutateResponse

type RejectPurchaseRequestRequest = it.RejectPurchaseRequestCommand
type RejectPurchaseRequestResponse = httpserver.RestMutateResponse

type CancelPurchaseRequestRequest = it.CancelPurchaseRequestCommand
type CancelPurchaseRequestResponse = httpserver.RestMutateResponse

type MarkPurchaseRequestPriorityRequest = it.MarkPurchaseRequestPriorityCommand
type MarkPurchaseRequestPriorityResponse = httpserver.RestMutateResponse

type ConvertPurchaseRequestToRfqRequest = it.ConvertPurchaseRequestToRfqCommand
type ConvertPurchaseRequestToRfqResponse = httpserver.RestMutateResponse

type ConvertPurchaseRequestToPoRequest = it.ConvertPurchaseRequestToPoCommand
type ConvertPurchaseRequestToPoResponse = httpserver.RestMutateResponse

type ConsolidatePurchaseRequestsRequest = it.ConsolidatePurchaseRequestsCommand
type ConsolidatePurchaseRequestsResponse = httpserver.RestMutateResponse
