package v1

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	domain "github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/rolerequest"
)

type roleRequestRestParams struct {
	dig.In

	RoleRequestSvc it.RoleRequestAppService
}

func NewRoleRequestRest(params roleRequestRestParams) *RoleRequestRest {
	return &RoleRequestRest{RoleRequestSvc: params.RoleRequestSvc}
}

type RoleRequestRest struct {
	httpserver.RestBase
	RoleRequestSvc it.RoleRequestAppService
}

func (this RoleRequestRest) CreateRoleRequest(echoCtx *echo.Context) (err error) {
	return httpserver.ServeCreate[CreateRoleRequestRequest, CreateRoleRequestResponse, domain.RoleRequest](
		"create grant request",
		echoCtx,
		&it.CreateRoleRequestCommand{},
		this.RoleRequestSvc.CreateRoleRequest,
	)
}

func (this RoleRequestRest) DeleteRoleRequest(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate[DeleteRoleRequestRequest, DeleteRoleRequestResponse](
		"delete grant request",
		echoCtx,
		this.RoleRequestSvc.DeleteRoleRequest,
	)
}

func (this RoleRequestRest) GetRoleRequest(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGetOne2[GetRoleRequestRequest, GetRoleRequestResponse, domain.RoleRequest](
		"get grant request",
		echoCtx,
		this.RoleRequestSvc.GetRoleRequest,
	)
}

func (this RoleRequestRest) RoleRequestExists(echoCtx *echo.Context) (err error) {
	return httpserver.ServeExists[RoleRequestExistsRequest, RoleRequestExistsResponse](
		"grant request exists",
		echoCtx,
		this.RoleRequestSvc.RoleRequestExists,
	)
}

func (this RoleRequestRest) SearchRoleRequests(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch[SearchRoleRequestsRequest, SearchRoleRequestsResponse, domain.RoleRequest](
		"search grant requests",
		echoCtx,
		this.RoleRequestSvc.SearchRoleRequests,
	)
}

func (this RoleRequestRest) UpdateRoleRequest(echoCtx *echo.Context) (err error) {
	return httpserver.ServeUpdate[UpdateRoleRequestRequest, UpdateRoleRequestResponse](
		"update grant request",
		echoCtx,
		&it.UpdateRoleRequestCommand{},
		this.RoleRequestSvc.UpdateRoleRequest,
	)
}

/*
 * Non-CRUD APIs
 */

func (this RoleRequestRest) GetModelSchema(echoCtx *echo.Context) (err error) {
	schema := dmodel.MustGetSchema(domain.RoleRequestSchemaName)
	echoCtx.JSON(http.StatusOK, schema.ToSimplized())
	return nil
}
