package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/permission"
)

type permissionRestParam struct {
	dig.In

	PermissionSvc it.PermissionAppService
}

func NewPermissionRest(params permissionRestParam) *PermissionRest {
	return &PermissionRest{
		permissionSvc: params.PermissionSvc,
	}
}

type PermissionRest struct {
	httpserver.RestBase
	permissionSvc it.PermissionAppService
}

// TestMyPermissions answers whether the CALLER holds the entitlement expression
// in the request body, and when they do, which grant paths answer for it.
//
// Authentication only, no entitlement gate: the endpoint reveals nothing except
// the caller's own access, and gating it would make it useless for exactly the
// plain users who most need to know why they are being refused.
func (this PermissionRest) TestMyPermissions(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST test my permissions"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.permissionSvc.TestMyPermissions,
		func(request TestMyPermissionsRequest) it.TestMyPermissionsQuery {
			return it.TestMyPermissionsQuery(request)
		},
		NewTestMyPermissionsResponse,
		httpserver.JsonOk,
	)
}
