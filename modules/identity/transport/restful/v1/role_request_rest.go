package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
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

func (this RoleRequestRest) CreateRoleRequest(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create grant request"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequestDynamic(
		echoCtx,
		this.RoleRequestSvc.CreateRoleRequest,
		func(requestFields dmodel.DynamicFields) it.CreateRoleRequestCommand {
			cmd := it.CreateRoleRequestCommand{}
			cmd.SetFieldData(requestFields)
			return cmd
		},
		func(data domain.RoleRequest) CreateRoleRequestResponse {
			response := httpserver.NewRestCreateResponseDyn(data.GetFieldData())
			return *response
		},
		httpserver.JsonCreated,
	)
}

func (this RoleRequestRest) DeleteRoleRequest(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete grant request"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.RoleRequestSvc.DeleteRoleRequest,
		func(request DeleteRoleRequestRequest) it.DeleteRoleRequestCommand {
			return it.DeleteRoleRequestCommand(request)
		},
		func(data dyn.MutateResultData) DeleteRoleRequestResponse {
			return httpserver.NewRestDeleteResponse2(data)
		},
		httpserver.JsonOk,
	)
}

func (this RoleRequestRest) GetRoleRequest(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get grant request"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.RoleRequestSvc.GetRoleRequest,
		func(request GetRoleRequestRequest) it.GetRoleRequestQuery {
			return it.GetRoleRequestQuery(request)
		},
		func(data domain.RoleRequest) GetRoleRequestResponse {
			return data.GetFieldData()
		},
		httpserver.JsonOk,
	)
}

func (this RoleRequestRest) RoleRequestExists(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST grant request exists"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.RoleRequestSvc.RoleRequestExists,
		func(request RoleRequestExistsRequest) it.RoleRequestExistsQuery {
			return it.RoleRequestExistsQuery(request)
		},
		func(data dyn.ExistsResultData) RoleRequestExistsResponse {
			return RoleRequestExistsResponse(data)
		},
		httpserver.JsonOk,
	)
}

func (this RoleRequestRest) SearchRoleRequests(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search grant requests"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.RoleRequestSvc.SearchRoleRequests,
		func(request SearchRoleRequestsRequest) it.SearchRoleRequestsQuery {
			return it.SearchRoleRequestsQuery(request)
		},
		func(data it.SearchRoleRequestsResultData) SearchRoleRequestsResponse {
			return httpserver.NewSearchUsersResponseDyn(data)
		},
		httpserver.JsonOk,
		true,
	)
}

func (this RoleRequestRest) UpdateRoleRequest(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update grant request"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequestDynamic(
		echoCtx,
		this.RoleRequestSvc.UpdateRoleRequest,
		func(requestFields dmodel.DynamicFields) it.UpdateRoleRequestCommand {
			cmd := it.UpdateRoleRequestCommand{}
			cmd.SetFieldData(requestFields)
			return cmd
		},
		httpserver.NewRestMutateResponse,
		httpserver.JsonOk,
	)
}
