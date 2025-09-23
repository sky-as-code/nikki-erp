package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"

	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/grant_request"
)

type grantRequestRestParams struct {
	dig.In

	GrantRequestSvc it.GrantRequestService
}

func NewGrantRequestRest(params grantRequestRestParams) *GrantRequestRest {
	return &GrantRequestRest{
		GrantRequestSvc: params.GrantRequestSvc,
	}
}

type GrantRequestRest struct {
	httpserver.RestBase
	GrantRequestSvc it.GrantRequestService
}

func (this GrantRequestRest) CreateGrantRequest(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST create grant request"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.GrantRequestSvc.CreateGrantRequest,
		func(request CreateGrantRequestRequest) it.CreateGrantRequestCommand {
			return it.CreateGrantRequestCommand(request)
		},
		func(result it.CreateGrantRequestResult) CreateGrantRequestResponse {
			response := CreateGrantRequestResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonCreated,
	)

	return err
}

func (this GrantRequestRest) CancelGrantRequest(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST cancel grant request"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.GrantRequestSvc.CancelGrantRequest,
		func(request CancelGrantRequestRequest) it.CancelGrantRequestCommand {
			return it.CancelGrantRequestCommand(request)
		},
		func(result it.CancelGrantRequestResult) CancelGrantRequestResponse {
			response := CancelGrantRequestResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}

func (this GrantRequestRest) RespondToGrantRequest(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST respond to grant request"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.GrantRequestSvc.RespondToGrantRequest,
		func(request RespondToGrantRequestRequest) it.RespondToGrantRequestCommand {
			return it.RespondToGrantRequestCommand(request)
		},
		func(result it.RespondToGrantRequestResult) RespondToGrantRequestResponse {
			response := RespondToGrantRequestResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}
