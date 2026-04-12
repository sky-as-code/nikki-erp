package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/orgunit"
)

type orgunitRestParams struct {
	dig.In

	OrgUnitSvc it.OrgUnitService
}

func NewOrgUnitRest(params orgunitRestParams) *OrgUnitRest {
	return &OrgUnitRest{
		OrgUnitSvc: params.OrgUnitSvc,
	}
}

type OrgUnitRest struct {
	httpserver.RestBase
	OrgUnitSvc it.OrgUnitService
}

func (this OrgUnitRest) CreateOrgUnit(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create org unit"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequestDynamic(
		echoCtx,
		this.OrgUnitSvc.CreateOrgUnit,
		func(requestFields dmodel.DynamicFields) it.CreateOrgUnitCommand {
			cmd := it.CreateOrgUnitCommand{}
			cmd.SetFieldData(requestFields)
			return cmd
		},
		func(data domain.OrganizationalUnit) CreateOrgUnitResponse {
			response := httpserver.NewRestCreateResponseDyn(data.GetFieldData())
			return *response
		},
		httpserver.JsonCreated,
	)
}

func (this OrgUnitRest) DeleteOrgUnit(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete org unit"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.OrgUnitSvc.DeleteOrgUnit,
		func(request DeleteOrgUnitRequest) it.DeleteOrgUnitCommand {
			return it.DeleteOrgUnitCommand(request)
		},
		func(data dyn.MutateResultData) DeleteOrgUnitResponse {
			return httpserver.NewRestDeleteResponse2(data)
		},
		httpserver.JsonOk,
	)
}

func (this OrgUnitRest) GetOrgUnit(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get org unit"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.OrgUnitSvc.GetOrgUnit,
		func(request GetOrgUnitRequest) it.GetOrgUnitQuery {
			return it.GetOrgUnitQuery(request)
		},
		func(data domain.OrganizationalUnit) dmodel.DynamicFields {
			return data.GetFieldData()
		},
		httpserver.JsonOk,
	)
}

func (this OrgUnitRest) OrgUnitExists(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST org unit exists"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.OrgUnitSvc.OrgUnitExists,
		func(request OrgUnitExistsRequest) it.OrgUnitExistsQuery {
			return it.OrgUnitExistsQuery(request)
		},
		func(data dyn.ExistsResultData) OrgUnitExistsResponse {
			return OrgUnitExistsResponse(data)
		},
		httpserver.JsonOk,
	)
}

func (this OrgUnitRest) ManageOrgUnitUsers(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST manage org unit users"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.OrgUnitSvc.ManageOrgUnitUsers,
		func(request ManageOrgUnitUsersRequest) it.ManageOrgUnitUsersCommand {
			return it.ManageOrgUnitUsersCommand(request)
		},
		func(data dyn.MutateResultData) ManageOrgUnitUsersResponse {
			return httpserver.NewRestMutateResponse(data)
		},
		httpserver.JsonOk,
	)
}

func (this OrgUnitRest) SearchOrgUnits(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search org units"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.OrgUnitSvc.SearchOrgUnits,
		func(request SearchOrgUnitsRequest) it.SearchOrgUnitsQuery {
			return it.SearchOrgUnitsQuery(request)
		},
		func(data it.SearchOrgUnitsResultData) SearchOrgUnitsResponse {
			return httpserver.NewSearchResponseDyn(data)
		},
		httpserver.JsonOk,
		true,
	)
}

func (this OrgUnitRest) UpdateOrgUnit(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update org unit"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequestDynamic(
		echoCtx,
		this.OrgUnitSvc.UpdateOrgUnit,
		func(requestFields dmodel.DynamicFields) it.UpdateOrgUnitCommand {
			cmd := it.UpdateOrgUnitCommand{}
			cmd.SetFieldData(requestFields)
			return cmd
		},
		httpserver.NewRestMutateResponse,
		httpserver.JsonOk,
	)
}
