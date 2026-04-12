package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
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
	return httpserver.ServeCreate(
		"create organization",
		echoCtx,
		&it.CreateOrgCommand{},
		this.OrgSvc.CreateOrg,
	)
}

func (this OrganizationRest) Delete(echoCtx echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"delete organization",
		echoCtx,
		this.OrgSvc.DeleteOrg,
	)
}

func (this OrganizationRest) GetOne(echoCtx echo.Context) (err error) {
	return httpserver.ServeGetOne(
		"get organization",
		echoCtx,
		this.OrgSvc.GetOrg,
	)
}

func (this OrganizationRest) ManageOrgUsers(echoCtx echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"manage org users",
		echoCtx,
		this.OrgSvc.ManageOrgUsers,
	)
}

func (this OrganizationRest) Exists(echoCtx echo.Context) (err error) {
	return httpserver.ServeExists(
		"organization exists",
		echoCtx,
		this.OrgSvc.OrgExists,
	)
}

func (this OrganizationRest) Search(echoCtx echo.Context) (err error) {
	return httpserver.ServeSearch(
		"search organizations",
		echoCtx,
		this.OrgSvc.SearchOrgs,
		true,
	)
}

func (this OrganizationRest) SetIsArchived(echoCtx echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"set organization is_archived",
		echoCtx,
		this.OrgSvc.SetOrgIsArchived,
	)
}

func (this OrganizationRest) Update(echoCtx echo.Context) (err error) {
	return httpserver.ServeUpdate(
		"update organization",
		echoCtx,
		&it.UpdateOrgCommand{},
		this.OrgSvc.UpdateOrg,
	)
}
