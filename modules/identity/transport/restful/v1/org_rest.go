package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itOrg "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
	itUser "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

type organizationRestParams struct {
	dig.In

	OrgSvc itOrg.OrganizationService
}

func NewOrganizationRest(params organizationRestParams) *OrganizationRest {
	return &OrganizationRest{
		OrgSvc: params.OrgSvc,
	}
}

type OrganizationRest struct {
	httpserver.RestBase
	OrgSvc itOrg.OrganizationService
}

func (this OrganizationRest) CreateOrganization(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create organization"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.OrgSvc.CreateOrganization,
		func(request CreateOrganizationRequest) itOrg.CreateOrganizationCommand {
			return itOrg.CreateOrganizationCommand(request)
		},
		func(result itOrg.CreateOrganizationResult) CreateOrganizationResponse {
			response := CreateOrganizationResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonCreated,
	)
	return err
}

func (this OrganizationRest) UpdateOrganization(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update organization"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.OrgSvc.UpdateOrganization,
		func(request UpdateOrganizationRequest) itOrg.UpdateOrganizationCommand {
			return itOrg.UpdateOrganizationCommand(request)
		},
		func(result itOrg.UpdateOrganizationResult) UpdateOrganizationResponse {
			response := UpdateOrganizationResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this OrganizationRest) DeleteOrganization(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete organization"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.OrgSvc.DeleteOrganization,
		func(request DeleteOrganizationRequest) itOrg.DeleteOrganizationCommand {
			return itOrg.DeleteOrganizationCommand(request)
		},
		func(result itOrg.DeleteOrganizationResult) DeleteOrganizationResponse {
			response := DeleteOrganizationResponse{}
			response.FromNonEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this OrganizationRest) GetOrganizationBySlug(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get organization by slug"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.OrgSvc.GetOrganizationBySlug,
		func(request GetOrganizationBySlugRequest) itOrg.GetOrganizationBySlugQuery {
			return itOrg.GetOrganizationBySlugQuery(request)
		},
		func(result itOrg.GetOrganizationBySlugResult) GetOrganizationBySlugResponse {
			response := GetOrganizationBySlugResponse{}
			response.FromOrg(*result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this OrganizationRest) SearchOrganizations(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search organizations"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.OrgSvc.SearchOrganizations,
		func(request SearchOrganizationsRequest) itOrg.SearchOrganizationsQuery {
			return itOrg.SearchOrganizationsQuery(request)
		},
		func(result itOrg.SearchOrganizationsResult) SearchOrganizationsResponse {
			response := SearchOrganizationsResponse{}
			response.FromResult(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this OrganizationRest) ListOrgStatuses(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST list org statuses"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.OrgSvc.ListOrgStatuses,
		func(request ListOrgStatusesRequest) itOrg.ListOrgStatusesQuery {
			return itOrg.ListOrgStatusesQuery(request)
		},
		func(result itUser.ListIdentStatusesResult) ListOrgStatusesResponse {
			return *result.Data
		},
		httpserver.JsonOk,
	)
	return err
}
