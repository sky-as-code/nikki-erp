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
	it "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/requestforquote"
)

type requestForQuoteRestParams struct {
	dig.In
	Svc it.RequestForQuoteService
}

func NewRequestForQuoteRest(params requestForQuoteRestParams) *RequestForQuoteRest {
	return &RequestForQuoteRest{svc: params.Svc}
}

type RequestForQuoteRest struct{ svc it.RequestForQuoteService }

func (this RequestForQuoteRest) CreateRequestForQuote(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create request for quote"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequestDynamic(echoCtx, this.svc.CreateRequestForQuote,
		func(fields dmodel.DynamicFields) it.CreateRequestForQuoteCommand {
			cmd := it.CreateRequestForQuoteCommand{RequestForQuote: *domain.NewRequestForQuote()}
			cmd.SetFieldData(fields)
			return cmd
		},
		func(data domain.RequestForQuote) CreateRequestForQuoteResponse {
			return *httpserver.NewRestCreateResponseDyn(data.GetFieldData())
		},
		httpserver.JsonCreated)
}
func (this RequestForQuoteRest) DeleteRequestForQuote(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete request for quote"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.DeleteRequestForQuote,
		func(request DeleteRequestForQuoteRequest) it.DeleteRequestForQuoteCommand {
			return it.DeleteRequestForQuoteCommand(request)
		},
		func(data dyn.MutateResultData) DeleteRequestForQuoteResponse {
			return httpserver.NewRestDeleteResponse2(data)
		}, httpserver.JsonOk)
}
func (this RequestForQuoteRest) GetRequestForQuote(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get request for quote"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.GetRequestForQuote,
		func(request GetRequestForQuoteRequest) it.GetRequestForQuoteQuery {
			return it.GetRequestForQuoteQuery(request)
		},
		func(data domain.RequestForQuote) GetRequestForQuoteResponse { return data.GetFieldData() }, httpserver.JsonOk)
}
func (this RequestForQuoteRest) RequestForQuoteExists(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST request for quote exists"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.RequestForQuoteExists,
		func(request RequestForQuoteExistsRequest) it.RequestForQuoteExistsQuery {
			return it.RequestForQuoteExistsQuery(request)
		},
		func(data dyn.ExistsResultData) RequestForQuoteExistsResponse {
			return RequestForQuoteExistsResponse(data)
		}, httpserver.JsonOk)
}
func (this RequestForQuoteRest) SearchRequestForQuotes(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search request for quotes"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.SearchRequestForQuotes,
		func(request SearchRequestForQuotesRequest) it.SearchRequestForQuotesQuery {
			return it.SearchRequestForQuotesQuery(request)
		},
		func(data it.SearchRequestForQuotesResultData) SearchRequestForQuotesResponse {
			return httpserver.NewSearchResponseDyn(data)
		},
		httpserver.JsonOk, true)
}
func (this RequestForQuoteRest) SetRequestForQuoteIsArchived(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST set request for quote archived"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.SetRequestForQuoteIsArchived,
		func(request SetRequestForQuoteIsArchivedRequest) it.SetRequestForQuoteIsArchivedCommand {
			return it.SetRequestForQuoteIsArchivedCommand(request)
		},
		httpserver.NewRestMutateResponse, httpserver.JsonOk)
}
func (this RequestForQuoteRest) UpdateRequestForQuote(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update request for quote"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.UpdateRequestForQuote,
		func(request UpdateRequestForQuoteRequest) it.UpdateRequestForQuoteCommand {
			cmd := it.UpdateRequestForQuoteCommand{RequestForQuote: *domain.NewRequestForQuote()}
			cmd.SetFieldData(request.DynamicFields)
			cmd.SetId(util.ToPtr(model.Id(request.RequestForQuoteId)))
			return cmd
		},
		httpserver.NewRestMutateResponse, httpserver.JsonOk)
}
