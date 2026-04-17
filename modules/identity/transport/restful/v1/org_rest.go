package v1

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
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

func (this OrganizationRest) CreateOrg(echoCtx *echo.Context) (err error) {
	return httpserver.ServeCreate(
		"create organization",
		echoCtx,
		&it.CreateOrgCommand{},
		this.OrgSvc.CreateOrg,
	)
}

func (this OrganizationRest) DeleteOrg(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"delete organization",
		echoCtx,
		this.OrgSvc.DeleteOrg,
	)
}

func (this OrganizationRest) GetOrg(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGetOne(
		"get organization",
		echoCtx,
		this.OrgSvc.GetOrg,
	)
}

func (this OrganizationRest) ManageOrgUsers(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"manage org users",
		echoCtx,
		this.OrgSvc.ManageOrgUsers,
	)
}

func (this OrganizationRest) OrgExists(echoCtx *echo.Context) (err error) {
	return httpserver.ServeExists(
		"organization exists",
		echoCtx,
		this.OrgSvc.OrgExists,
	)
}

func (this OrganizationRest) SearchOrgs(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch(
		"search organizations",
		echoCtx,
		this.OrgSvc.SearchOrgs,
	)
}

func (this OrganizationRest) SetOrgIsArchived(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"set organization is_archived",
		echoCtx,
		this.OrgSvc.SetOrgIsArchived,
	)
}

func (this OrganizationRest) UpdateOrg(echoCtx *echo.Context) (err error) {
	return httpserver.ServeUpdate(
		"update organization",
		echoCtx,
		&it.UpdateOrgCommand{},
		this.OrgSvc.UpdateOrg,
	)
}

/*
 * Non-CRUD APIs
 */

func (this OrganizationRest) GetModelSchema(echoCtx *echo.Context) (err error) {
	schema := dmodel.MustGetSchema(domain.OrganizationSchemaName)
	echoCtx.JSON(http.StatusOK, schema.ToSimplized())
	return nil
}
