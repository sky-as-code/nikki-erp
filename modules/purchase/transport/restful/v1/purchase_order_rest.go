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
	it "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/purchaseorder"
)

type purchaseOrderRestParams struct {
	dig.In
	Svc it.PurchaseOrderService
}

func NewPurchaseOrderRest(params purchaseOrderRestParams) *PurchaseOrderRest {
	return &PurchaseOrderRest{svc: params.Svc}
}

type PurchaseOrderRest struct{ svc it.PurchaseOrderService }

func (this PurchaseOrderRest) CreatePurchaseOrder(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create purchase order"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequestDynamic(echoCtx, this.svc.CreatePurchaseOrder,
		func(fields dmodel.DynamicFields) it.CreatePurchaseOrderCommand {
			cmd := it.CreatePurchaseOrderCommand{PurchaseOrder: *domain.NewPurchaseOrder()}
			cmd.SetFieldData(fields)
			return cmd
		},
		func(data domain.PurchaseOrder) CreatePurchaseOrderResponse {
			return *httpserver.NewRestCreateResponseDyn(data.GetFieldData())
		},
		httpserver.JsonCreated)
}
func (this PurchaseOrderRest) DeletePurchaseOrder(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete purchase order"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.DeletePurchaseOrder,
		func(request DeletePurchaseOrderRequest) it.DeletePurchaseOrderCommand {
			return it.DeletePurchaseOrderCommand(request)
		},
		func(data dyn.MutateResultData) DeletePurchaseOrderResponse {
			return httpserver.NewRestDeleteResponse2(data)
		}, httpserver.JsonOk)
}
func (this PurchaseOrderRest) GetPurchaseOrder(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get purchase order"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.GetPurchaseOrder,
		func(request GetPurchaseOrderRequest) it.GetPurchaseOrderQuery {
			return it.GetPurchaseOrderQuery(request)
		},
		func(data domain.PurchaseOrder) GetPurchaseOrderResponse { return data.GetFieldData() }, httpserver.JsonOk)
}
func (this PurchaseOrderRest) PurchaseOrderExists(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST purchase order exists"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.PurchaseOrderExists,
		func(request PurchaseOrderExistsRequest) it.PurchaseOrderExistsQuery {
			return it.PurchaseOrderExistsQuery(request)
		},
		func(data dyn.ExistsResultData) PurchaseOrderExistsResponse { return PurchaseOrderExistsResponse(data) }, httpserver.JsonOk)
}
func (this PurchaseOrderRest) SearchPurchaseOrders(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search purchase orders"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.SearchPurchaseOrders,
		func(request SearchPurchaseOrdersRequest) it.SearchPurchaseOrdersQuery {
			return it.SearchPurchaseOrdersQuery(request)
		},
		func(data it.SearchPurchaseOrdersResultData) SearchPurchaseOrdersResponse {
			return httpserver.NewSearchResponseDyn(data)
		},
		httpserver.JsonOk, true)
}
func (this PurchaseOrderRest) SetPurchaseOrderIsArchived(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST set purchase order archived"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.SetPurchaseOrderIsArchived,
		func(request SetPurchaseOrderIsArchivedRequest) it.SetPurchaseOrderIsArchivedCommand {
			return it.SetPurchaseOrderIsArchivedCommand(request)
		}, httpserver.NewRestMutateResponse, httpserver.JsonOk)
}
func (this PurchaseOrderRest) UpdatePurchaseOrder(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update purchase order"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.UpdatePurchaseOrder,
		func(request UpdatePurchaseOrderRequest) it.UpdatePurchaseOrderCommand {
			cmd := it.UpdatePurchaseOrderCommand{PurchaseOrder: *domain.NewPurchaseOrder()}
			cmd.SetFieldData(request.DynamicFields)
			cmd.SetId(util.ToPtr(model.Id(request.PurchaseOrderId)))
			return cmd
		}, httpserver.NewRestMutateResponse, httpserver.JsonOk)
}
