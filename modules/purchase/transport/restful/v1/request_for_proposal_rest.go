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
	it "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/requestforproposal"
)

type requestForProposalRestParams struct {
	dig.In
	Svc it.RequestForProposalService
}

func NewRequestForProposalRest(params requestForProposalRestParams) *RequestForProposalRest {
	return &RequestForProposalRest{svc: params.Svc}
}

type RequestForProposalRest struct{ svc it.RequestForProposalService }

func (this RequestForProposalRest) CreateRequestForProposal(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create request for proposal"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequestDynamic(echoCtx, this.svc.CreateRequestForProposal,
		func(fields dmodel.DynamicFields) it.CreateRequestForProposalCommand {
			cmd := it.CreateRequestForProposalCommand{RequestForProposal: *domain.NewRequestForProposal()}
			cmd.SetFieldData(fields)
			return cmd
		},
		func(data domain.RequestForProposal) CreateRequestForProposalResponse {
			return *httpserver.NewRestCreateResponseDyn(data.GetFieldData())
		},
		httpserver.JsonCreated)
}
func (this RequestForProposalRest) DeleteRequestForProposal(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete request for proposal"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.DeleteRequestForProposal,
		func(request DeleteRequestForProposalRequest) it.DeleteRequestForProposalCommand {
			return it.DeleteRequestForProposalCommand(request)
		},
		func(data dyn.MutateResultData) DeleteRequestForProposalResponse {
			return httpserver.NewRestDeleteResponse2(data)
		}, httpserver.JsonOk)
}
func (this RequestForProposalRest) GetRequestForProposal(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get request for proposal"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.GetRequestForProposal,
		func(request GetRequestForProposalRequest) it.GetRequestForProposalQuery {
			return it.GetRequestForProposalQuery(request)
		},
		func(data domain.RequestForProposal) GetRequestForProposalResponse { return data.GetFieldData() }, httpserver.JsonOk)
}
func (this RequestForProposalRest) RequestForProposalExists(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST request for proposal exists"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.RequestForProposalExists,
		func(request RequestForProposalExistsRequest) it.RequestForProposalExistsQuery {
			return it.RequestForProposalExistsQuery(request)
		},
		func(data dyn.ExistsResultData) RequestForProposalExistsResponse {
			return RequestForProposalExistsResponse(data)
		}, httpserver.JsonOk)
}
func (this RequestForProposalRest) SearchRequestForProposals(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search request for proposals"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.SearchRequestForProposals,
		func(request SearchRequestForProposalsRequest) it.SearchRequestForProposalsQuery {
			return it.SearchRequestForProposalsQuery(request)
		},
		func(data it.SearchRequestForProposalsResultData) SearchRequestForProposalsResponse {
			return httpserver.NewSearchResponseDyn(data)
		},
		httpserver.JsonOk, true)
}
func (this RequestForProposalRest) SetRequestForProposalIsArchived(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST set request for proposal archived"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.SetRequestForProposalIsArchived,
		func(request SetRequestForProposalIsArchivedRequest) it.SetRequestForProposalIsArchivedCommand {
			return it.SetRequestForProposalIsArchivedCommand(request)
		}, httpserver.NewRestMutateResponse, httpserver.JsonOk)
}
func (this RequestForProposalRest) UpdateRequestForProposal(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update request for proposal"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(echoCtx, this.svc.UpdateRequestForProposal,
		func(request UpdateRequestForProposalRequest) it.UpdateRequestForProposalCommand {
			cmd := it.UpdateRequestForProposalCommand{RequestForProposal: *domain.NewRequestForProposal()}
			cmd.SetFieldData(request.DynamicFields)
			cmd.SetId(util.ToPtr(model.Id(request.RequestForProposalId)))
			return cmd
		}, httpserver.NewRestMutateResponse, httpserver.JsonOk)
}
