package restful

import (
	stdErr "errors"

	"github.com/labstack/echo/v5"
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	v1 "github.com/sky-as-code/nikki-erp/modules/purchase/transport/restful/v1"
)

func InitRestfulHandlers() error {
	err := deps.Register(
		v1.NewPurchaseOrderRest,
		v1.NewPurchaseRequestRest,
		v1.NewRequestForProposalRest,
		v1.NewRequestForQuoteRest,
		v1.NewVendorRest,
	)
	if err != nil {
		return err
	}

	return stdErr.Join(
		initPurchaseOrderV1(),
		initPurchaseRequestV1(),
		initRequestForProposalV1(),
		initRequestForQuoteV1(),
		initVendorV1(),
	)

}

func initPurchaseOrderV1() error {
	return deps.Invoke(func(route *echo.Group, rest *v1.PurchaseOrderRest) {
		routeV1 := route.Group("/v1/purchase")
		routeV1.DELETE("/purchase-orders/:id", rest.DeletePurchaseOrder)
		routeV1.GET("/purchase-orders/:id", rest.GetPurchaseOrder)
		routeV1.GET("/purchase-orders", rest.SearchPurchaseOrders)
		routeV1.POST("/purchase-orders/exists", rest.PurchaseOrderExists)
		routeV1.POST("/purchase-orders/:id/archived", rest.SetPurchaseOrderIsArchived)
		routeV1.POST("/purchase-orders", rest.CreatePurchaseOrder)
		routeV1.PUT("/purchase-orders/:id", rest.UpdatePurchaseOrder)
	})
}

func initPurchaseRequestV1() error {
	return deps.Invoke(func(route *echo.Group, rest *v1.PurchaseRequestRest) {
		routeV1 := route.Group("/v1/purchase")
		routeV1.DELETE("/purchase-requests/:id", rest.DeletePurchaseRequest)
		routeV1.GET("/purchase-requests/:id", rest.GetPurchaseRequest)
		routeV1.GET("/purchase-requests", rest.SearchPurchaseRequests)
		routeV1.POST("/purchase-requests/exists", rest.PurchaseRequestExists)
		routeV1.POST("/purchase-requests/:id/archived", rest.SetPurchaseRequestIsArchived)
		routeV1.POST("/purchase-requests/:id/submit", rest.SubmitPurchaseRequestForApproval)
		routeV1.POST("/purchase-requests/:id/approve", rest.ApprovePurchaseRequest)
		routeV1.POST("/purchase-requests/:id/reject", rest.RejectPurchaseRequest)
		routeV1.POST("/purchase-requests/:id/cancel", rest.CancelPurchaseRequest)
		routeV1.POST("/purchase-requests/:id/priority", rest.MarkPurchaseRequestPriority)
		routeV1.POST("/purchase-requests/:id/convert-rfq", rest.ConvertPurchaseRequestToRfq)
		routeV1.POST("/purchase-requests/:id/convert-po", rest.ConvertPurchaseRequestToPo)
		routeV1.POST("/purchase-requests/consolidate", rest.ConsolidatePurchaseRequests)
		routeV1.POST("/purchase-requests", rest.CreatePurchaseRequest)
		routeV1.PUT("/purchase-requests/:id", rest.UpdatePurchaseRequest)
	})
}

func initRequestForProposalV1() error {
	return deps.Invoke(func(route *echo.Group, rest *v1.RequestForProposalRest) {
		routeV1 := route.Group("/v1/purchase")
		routeV1.DELETE("/request-for-proposals/:id", rest.DeleteRequestForProposal)
		routeV1.GET("/request-for-proposals/:id", rest.GetRequestForProposal)
		routeV1.GET("/request-for-proposals", rest.SearchRequestForProposals)
		routeV1.POST("/request-for-proposals/exists", rest.RequestForProposalExists)
		routeV1.POST("/request-for-proposals/:id/archived", rest.SetRequestForProposalIsArchived)
		routeV1.POST("/request-for-proposals", rest.CreateRequestForProposal)
		routeV1.PUT("/request-for-proposals/:id", rest.UpdateRequestForProposal)
	})
}

func initRequestForQuoteV1() error {
	return deps.Invoke(func(route *echo.Group, rest *v1.RequestForQuoteRest) {
		routeV1 := route.Group("/v1/purchase")
		routeV1.DELETE("/request-for-quotes/:id", rest.DeleteRequestForQuote)
		routeV1.GET("/request-for-quotes/:id", rest.GetRequestForQuote)
		routeV1.GET("/request-for-quotes", rest.SearchRequestForQuotes)
		routeV1.POST("/request-for-quotes/exists", rest.RequestForQuoteExists)
		routeV1.POST("/request-for-quotes/:id/archived", rest.SetRequestForQuoteIsArchived)
		routeV1.POST("/request-for-quotes", rest.CreateRequestForQuote)
		routeV1.PUT("/request-for-quotes/:id", rest.UpdateRequestForQuote)
	})
}

func initVendorV1() error {
	return deps.Invoke(func(route *echo.Group, rest *v1.VendorRest) {
		routeV1 := route.Group("/v1/purchase")
		routeV1.DELETE("/vendors/:id", rest.DeleteVendor)
		routeV1.GET("/vendors/:id", rest.GetVendor)
		routeV1.GET("/vendors", rest.SearchVendors)
		routeV1.POST("/vendors/exists", rest.VendorExists)
		routeV1.POST("/vendors/:id/archived", rest.SetVendorIsArchived)
		routeV1.POST("/vendors", rest.CreateVendor)
		routeV1.PUT("/vendors/:id", rest.UpdateVendor)
	})
}
