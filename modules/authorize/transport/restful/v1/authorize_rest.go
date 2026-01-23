package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"

	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces"
)

type authorizeRestParams struct {
	dig.In

	AuthorizeSvc it.AuthorizeService
}

func NewAuthorizeRest(params authorizeRestParams) *AuthorizeRest {
	return &AuthorizeRest{
		AuthorizeSvc: params.AuthorizeSvc,
	}
}

type AuthorizeRest struct {
	httpserver.RestBase
	AuthorizeSvc it.AuthorizeService
}

func (this AuthorizeRest) IsAuthorized(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST is authorized"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx,
		this.AuthorizeSvc.IsAuthorized,
		func(request IsAuthorizedRequest) it.IsAuthorizedQuery {
			return it.IsAuthorizedQuery(request)
		},
		func(result it.IsAuthorizedResult) IsAuthorizedResponse {
			response := IsAuthorizedResponse{}
			response.FromResult(result)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}

func (this AuthorizeRest) PermissionSnapshot(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST permission snapshot"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx,
		this.AuthorizeSvc.PermissionSnapshot,
		func(request PermissionSnapshotRequest) it.PermissionSnapshotQuery {
			return it.PermissionSnapshotQuery(request)
		},
		func(result it.PermissionSnapshotResult) PermissionSnapshotResponse {
			response := PermissionSnapshotResponse{}
			response.FromResult(result)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}
