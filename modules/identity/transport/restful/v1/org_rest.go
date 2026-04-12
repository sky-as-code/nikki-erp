package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
)

type organizationRestParams struct {
	dig.In

	OrgSvc it.OrganizationService
}

func NewOrganizationRest(params organizationRestParams) *OrganizationRest {
	return &OrganizationRest{
		OrgSvc: params.OrgSvc,
	}
}

type OrganizationRest struct {
	httpserver.RestBase
	OrgSvc it.OrganizationService
}

func (this OrganizationRest) Create(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create organization"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequestDynamic(
		echoCtx,
		this.OrgSvc.CreateOrg,
		func(requestFields dmodel.DynamicFields) it.CreateOrgCommand {
			cmd := it.CreateOrgCommand{}
			cmd.SetFieldData(requestFields)
			return cmd
		},
		func(data domain.Organization) CreateOrgResponse {
			response := httpserver.NewRestCreateResponseDyn(data.GetFieldData())
			return *response
		},
		httpserver.JsonCreated,
	)
}

func (this OrganizationRest) Delete(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete organization"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.OrgSvc.DeleteOrg,
		func(request DeleteOrgRequest) it.DeleteOrgCommand {
			return it.DeleteOrgCommand(request)
		},
		func(data dyn.MutateResultData) DeleteOrgResponse {
			return httpserver.NewRestDeleteResponse2(data)
		},
		httpserver.JsonOk,
	)
}

func (this OrganizationRest) GetOne(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get organization by slug"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.OrgSvc.GetOrg,
		func(request GetOrgRequest) it.GetOrgQuery {
			return it.GetOrgQuery(request)
		},
		func(data domain.Organization) GetOrgResponse {
			return data.GetFieldData()
		},
		httpserver.JsonOk,
	)
}

func (this OrganizationRest) ManageOrgUsers(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST manage org users"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.OrgSvc.ManageOrgUsers,
		func(request ManageOrgUsersRequest) it.ManageOrgUsersCommand {
			return it.ManageOrgUsersCommand(request)
		},
		func(data dyn.MutateResultData) ManageOrgsResponse {
			return httpserver.NewRestMutateResponse(data)
		},
		httpserver.JsonOk,
	)
}

func (this OrganizationRest) Exists(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST organization exists"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.OrgSvc.OrgExists,
		func(request OrgExistsRequest) it.OrgExistsQuery {
			return it.OrgExistsQuery(request)
		},
		func(data dyn.ExistsResultData) OrgExistsResponse {
			return OrgExistsResponse(data)
		},
		httpserver.JsonOk,
	)
}

func (this OrganizationRest) Search(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search organizations"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.OrgSvc.SearchOrgs,
		func(request SearchOrgsRequest) it.SearchOrgsQuery {
			return it.SearchOrgsQuery(request)
		},
		func(data it.SearchOrgsResultData) SearchOrgsResponse {
			return httpserver.NewSearchResponseDyn(data)
		},
		httpserver.JsonOk,
		true,
	)
}

func (this OrganizationRest) SetIsArchived(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST set organization is_archived"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.OrgSvc.SetOrgIsArchived,
		func(request SetOrgIsArchivedRequest) it.SetOrgIsArchivedCommand {
			return request
		},
		httpserver.NewRestMutateResponse,
		httpserver.JsonOk,
	)
}

func (this OrganizationRest) Update(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update organization"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequestDynamic(
		echoCtx,
		this.OrgSvc.UpdateOrg,
		func(requestFields dmodel.DynamicFields) it.UpdateOrgCommand {
			cmd := it.UpdateOrgCommand{}
			cmd.SetFieldData(requestFields)
			return cmd
		},
		httpserver.NewRestMutateResponse,
		httpserver.JsonOk,
	)
}
