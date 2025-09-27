package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"

	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/revoke_request"
)

type revokeRequestRestParams struct {
	dig.In

	RevokeRequestSvc it.RevokeRequestService
}

func NewRevokeRequestRest(params revokeRequestRestParams) *RevokeRequestRest {
	return &RevokeRequestRest{
		RevokeRequestSvc: params.RevokeRequestSvc,
	}
}

type RevokeRequestRest struct {
	httpserver.RestBase
	RevokeRequestSvc it.RevokeRequestService
}

func (this RevokeRequestRest) Create(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST create revoke request"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.RevokeRequestSvc.Create,
		func(request CreateRevokeRequestRequest) it.CreateRevokeRequestCommand {
			return it.CreateRevokeRequestCommand(request)
		},
		func(result it.CreateRevokeRequestResult) CreateRevokeRequestResponse {
			response := CreateRevokeRequestResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonCreated,
	)

	return err
}
