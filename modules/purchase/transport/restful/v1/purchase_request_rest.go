package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain"
	it "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/purchaserequest"
)

type purchaseRequestRestParams struct {
	dig.In
	PurchaseRequestSvc it.PurchaseRequestService
}

func NewPurchaseRequestRest(params purchaseRequestRestParams) *PurchaseRequestRest {
	return &PurchaseRequestRest{svc: params.PurchaseRequestSvc}
}

type PurchaseRequestRest struct {
	svc it.PurchaseRequestService
}

func (this PurchaseRequestRest) CreatePurchaseRequest(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create purchase request"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequestDynamic(
		echoCtx,
		this.svc.CreatePurchaseRequest,
		func(requestFields dmodel.DynamicFields) it.CreatePurchaseRequestCommand {
			cmd := it.CreatePurchaseRequestCommand{PurchaseRequest: *domain.NewPurchaseRequest()}
			cmd.SetFieldData(requestFields)
			return cmd
		},
		func(data domain.PurchaseRequest) CreatePurchaseRequestResponse {
			return *httpserver.NewRestCreateResponseDyn(data.GetFieldData())
		},
		httpserver.JsonCreated,
	)
}

func (this PurchaseRequestRest) DeletePurchaseRequest(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete purchase request"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.DeletePurchaseRequest,
		func(request DeletePurchaseRequestRequest) it.DeletePurchaseRequestCommand {
			return it.DeletePurchaseRequestCommand(request)
		},
		func(data dyn.MutateResultData) DeletePurchaseRequestResponse {
			return httpserver.NewRestDeleteResponse2(data)
		},
		httpserver.JsonOk,
	)
}

func (this PurchaseRequestRest) GetPurchaseRequest(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get purchase request"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.GetPurchaseRequest,
		func(request GetPurchaseRequestRequest) it.GetPurchaseRequestQuery {
			return it.GetPurchaseRequestQuery(request)
		},
		func(data domain.PurchaseRequest) GetPurchaseRequestResponse { return data.GetFieldData() },
		httpserver.JsonOk,
	)
}

func (this PurchaseRequestRest) PurchaseRequestExists(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST purchase request exists"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.PurchaseRequestExists,
		func(request PurchaseRequestExistsRequest) it.PurchaseRequestExistsQuery {
			return it.PurchaseRequestExistsQuery(request)
		},
		func(data dyn.ExistsResultData) PurchaseRequestExistsResponse {
			return PurchaseRequestExistsResponse(data)
		},
		httpserver.JsonOk,
	)
}

func (this PurchaseRequestRest) SearchPurchaseRequests(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search purchase requests"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.SearchPurchaseRequests,
		func(request SearchPurchaseRequestsRequest) it.SearchPurchaseRequestsQuery {
			return it.SearchPurchaseRequestsQuery(request)
		},
		func(data it.SearchPurchaseRequestsResultData) SearchPurchaseRequestsResponse {
			return httpserver.NewSearchResponseDyn(data)
		},
		httpserver.JsonOk,
		true,
	)
}

func (this PurchaseRequestRest) SetPurchaseRequestIsArchived(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST set purchase request archived"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.SetPurchaseRequestIsArchived,
		func(request SetPurchaseRequestIsArchivedRequest) it.SetPurchaseRequestIsArchivedCommand {
			return it.SetPurchaseRequestIsArchivedCommand(request)
		},
		httpserver.NewRestMutateResponse,
		httpserver.JsonOk,
	)
}

func (this PurchaseRequestRest) UpdatePurchaseRequest(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update purchase request"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.UpdatePurchaseRequest,
		func(request UpdatePurchaseRequestRequest) it.UpdatePurchaseRequestCommand {
			cmd := it.UpdatePurchaseRequestCommand{PurchaseRequest: *domain.NewPurchaseRequest()}
			cmd.SetFieldData(request.DynamicFields)
			cmd.SetId(util.ToPtr(model.Id(request.PurchaseRequestId)))
			return cmd
		},
		httpserver.NewRestMutateResponse,
		httpserver.JsonOk,
	)
}

func (this PurchaseRequestRest) SubmitPurchaseRequestForApproval(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST submit purchase request"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.SubmitPurchaseRequestForApproval,
		func(request SubmitPurchaseRequestForApprovalRequest) it.SubmitPurchaseRequestForApprovalCommand {
			return request
		},
		httpserver.NewRestMutateResponse, httpserver.JsonOk)
}

func (this PurchaseRequestRest) ApprovePurchaseRequest(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST approve purchase request"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.ApprovePurchaseRequest,
		func(request ApprovePurchaseRequestRequest) it.ApprovePurchaseRequestCommand { return request },
		httpserver.NewRestMutateResponse, httpserver.JsonOk)
}

func (this PurchaseRequestRest) RejectPurchaseRequest(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST reject purchase request"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.RejectPurchaseRequest,
		func(request RejectPurchaseRequestRequest) it.RejectPurchaseRequestCommand { return request },
		httpserver.NewRestMutateResponse, httpserver.JsonOk)
}

func (this PurchaseRequestRest) CancelPurchaseRequest(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST cancel purchase request"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.CancelPurchaseRequest,
		func(request CancelPurchaseRequestRequest) it.CancelPurchaseRequestCommand { return request },
		httpserver.NewRestMutateResponse, httpserver.JsonOk)
}

func (this PurchaseRequestRest) MarkPurchaseRequestPriority(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST mark purchase request priority"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.MarkPurchaseRequestPriority,
		func(request MarkPurchaseRequestPriorityRequest) it.MarkPurchaseRequestPriorityCommand { return request },
		httpserver.NewRestMutateResponse, httpserver.JsonOk)
}

func (this PurchaseRequestRest) ConvertPurchaseRequestToRfq(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST convert purchase request to rfq"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.ConvertPurchaseRequestToRfq,
		func(request ConvertPurchaseRequestToRfqRequest) it.ConvertPurchaseRequestToRfqCommand { return request },
		httpserver.NewRestMutateResponse, httpserver.JsonOk)
}

func (this PurchaseRequestRest) ConvertPurchaseRequestToPo(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST convert purchase request to po"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.ConvertPurchaseRequestToPo,
		func(request ConvertPurchaseRequestToPoRequest) it.ConvertPurchaseRequestToPoCommand { return request },
		httpserver.NewRestMutateResponse, httpserver.JsonOk)
}

func (this PurchaseRequestRest) ConsolidatePurchaseRequests(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST consolidate purchase requests"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.ConsolidatePurchaseRequests,
		func(request ConsolidatePurchaseRequestsRequest) it.ConsolidatePurchaseRequestsCommand { return request },
		httpserver.NewRestMutateResponse, httpserver.JsonOk)
}
