package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/entitlement"
)

type entitlementRestParams struct {
	dig.In

	EntitlementSvc it.EntitlementService
}

func NewEntitlementRest(params entitlementRestParams) *EntitlementRest {
	return &EntitlementRest{EntitlementSvc: params.EntitlementSvc}
}

type EntitlementRest struct {
	httpserver.RestBase
	EntitlementSvc it.EntitlementService
}

func (this EntitlementRest) CreateEntitlement(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create entitlement"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequestDynamic(
		echoCtx,
		this.EntitlementSvc.CreateEntitlement,
		func(requestFields dmodel.DynamicFields) it.CreateEntitlementCommand {
			cmd := it.CreateEntitlementCommand{}
			cmd.SetFieldData(requestFields)
			return cmd
		},
		func(data domain.Entitlement) CreateEntitlementResponse {
			response := httpserver.NewRestCreateResponseDyn(data.GetFieldData())
			return *response
		},
		httpserver.JsonCreated,
	)
}

func (this EntitlementRest) DeleteEntitlement(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete entitlement"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.EntitlementSvc.DeleteEntitlement,
		func(request DeleteEntitlementRequest) it.DeleteEntitlementCommand {
			return it.DeleteEntitlementCommand(request)
		},
		func(data dyn.MutateResultData) DeleteEntitlementResponse {
			return httpserver.NewRestDeleteResponse2(data)
		},
		httpserver.JsonOk,
	)
}

func (this EntitlementRest) GetEntitlement(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get entitlement"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.EntitlementSvc.GetEntitlement,
		func(request GetEntitlementRequest) it.GetEntitlementQuery {
			return it.GetEntitlementQuery(request)
		},
		func(data domain.Entitlement) GetEntitlementResponse {
			return data.GetFieldData()
		},
		httpserver.JsonOk,
	)
}

func (this EntitlementRest) EntitlementExists(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST entitlement exists"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.EntitlementSvc.EntitlementExists,
		func(request EntitlementExistsRequest) it.EntitlementExistsQuery {
			return it.EntitlementExistsQuery(request)
		},
		func(data dyn.ExistsResultData) EntitlementExistsResponse {
			return EntitlementExistsResponse(data)
		},
		httpserver.JsonOk,
	)
}

func (this EntitlementRest) ManageEntitlementRoles(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST manage entitlement roles"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.EntitlementSvc.ManageEntitlementRoles,
		func(request ManageEntitlementRolesRequest) it.ManageEntitlementRolesCommand {
			return it.ManageEntitlementRolesCommand(request)
		},
		func(data dyn.MutateResultData) ManageEntitlementRolesResponse {
			return httpserver.NewRestMutateResponse(data)
		},
		httpserver.JsonOk,
	)
}

func (this EntitlementRest) SearchEntitlements(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search entitlements"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.EntitlementSvc.SearchEntitlements,
		func(request SearchEntitlementsRequest) it.SearchEntitlementsQuery {
			return it.SearchEntitlementsQuery(request)
		},
		func(data it.SearchEntitlementsResultData) SearchEntitlementsResponse {
			return httpserver.NewSearchUsersResponseDyn(data)
		},
		httpserver.JsonOk,
		true,
	)
}

func (this EntitlementRest) SetEntitlementIsArchived(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST set entitlement is_archived"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.EntitlementSvc.SetEntitlementIsArchived,
		func(request SetEntitlementIsArchivedRequest) it.SetEntitlementIsArchivedCommand {
			return request
		},
		func(data dyn.MutateResultData) SetEntitlementIsArchivedResponse {
			return httpserver.NewRestUpdateResponse2(data)
		},
		httpserver.JsonOk,
	)
}

func (this EntitlementRest) UpdateEntitlement(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update entitlement"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequestDynamic(
		echoCtx,
		this.EntitlementSvc.UpdateEntitlement,
		func(requestFields dmodel.DynamicFields) it.UpdateEntitlementCommand {
			cmd := it.UpdateEntitlementCommand{}
			cmd.SetFieldData(requestFields)
			return cmd
		},
		func(data dyn.MutateResultData) UpdateEntitlementResponse {
			return httpserver.NewRestUpdateResponse2(data)
		},
		httpserver.JsonOk,
	)
}
