package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/purchaseorder"
)

type CreatePurchaseOrderRequest struct{ dmodel.DynamicFields }
type CreatePurchaseOrderResponse = httpserver.RestCreateResponse
type DeletePurchaseOrderRequest = it.DeletePurchaseOrderCommand
type DeletePurchaseOrderResponse = httpserver.RestDeleteResponse2
type GetPurchaseOrderRequest = it.GetPurchaseOrderQuery
type GetPurchaseOrderResponse = dmodel.DynamicFields
type PurchaseOrderExistsRequest = it.PurchaseOrderExistsQuery
type PurchaseOrderExistsResponse = dyn.ExistsResultData
type SearchPurchaseOrdersRequest = it.SearchPurchaseOrdersQuery
type SearchPurchaseOrdersResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
type SetPurchaseOrderIsArchivedRequest = it.SetPurchaseOrderIsArchivedCommand
type SetPurchaseOrderIsArchivedResponse = httpserver.RestMutateResponse
type UpdatePurchaseOrderRequest struct {
	dmodel.DynamicFields
	PurchaseOrderId string `param:"id"`
}
type UpdatePurchaseOrderResponse = httpserver.RestMutateResponse
