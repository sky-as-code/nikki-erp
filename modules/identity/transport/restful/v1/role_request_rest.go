package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/rolerequest"
)

type roleRequestRestParams struct {
	dig.In

	RoleRequestSvc it.RoleRequestService
}

func NewRoleRequestRest(params roleRequestRestParams) *RoleRequestRest {
	return &RoleRequestRest{RoleRequestSvc: params.RoleRequestSvc}
}

type RoleRequestRest struct {
	httpserver.RestBase
	RoleRequestSvc it.RoleRequestService
}

func (this RoleRequestRest) CreateRoleRequest(echoCtx *echo.Context) (err error) {
	return httpserver.ServeCreate(
		"create grant request",
		echoCtx,
		&it.CreateRoleRequestCommand{},
		this.RoleRequestSvc.CreateRoleRequest,
	)
}

func (this RoleRequestRest) DeleteRoleRequest(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"delete grant request",
		echoCtx,
		this.RoleRequestSvc.DeleteRoleRequest,
	)
}

func (this RoleRequestRest) GetRoleRequest(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGetOne(
		"get grant request",
		echoCtx,
		this.RoleRequestSvc.GetRoleRequest,
	)
}

func (this RoleRequestRest) RoleRequestExists(echoCtx *echo.Context) (err error) {
	return httpserver.ServeExists(
		"grant request exists",
		echoCtx,
		this.RoleRequestSvc.RoleRequestExists,
	)
}

func (this RoleRequestRest) SearchRoleRequests(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch(
		"search grant requests",
		echoCtx,
		this.RoleRequestSvc.SearchRoleRequests,
		true,
	)
}

func (this RoleRequestRest) UpdateRoleRequest(echoCtx *echo.Context) (err error) {
	return httpserver.ServeUpdate(
		"update grant request",
		echoCtx,
		&it.UpdateRoleRequestCommand{},
		this.RoleRequestSvc.UpdateRoleRequest,
	)
}
