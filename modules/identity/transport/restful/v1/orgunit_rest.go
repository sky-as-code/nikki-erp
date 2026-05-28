package v1

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	domain "github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/orgunit"
)

type orgunitRestParams struct {
	dig.In

	OrgUnitSvc it.OrgUnitAppService
}

func NewOrgUnitRest(params orgunitRestParams) *OrgUnitRest {
	return &OrgUnitRest{
		OrgUnitSvc: params.OrgUnitSvc,
	}
}

type OrgUnitRest struct {
	httpserver.RestBase
	OrgUnitSvc it.OrgUnitAppService
}

func (this OrgUnitRest) CreateOrgUnit(echoCtx *echo.Context) (err error) {
	return httpserver.ServeCreate[CreateOrgUnitRequest, CreateOrgUnitResponse, domain.OrganizationalUnit](
		"create org unit",
		echoCtx,
		&it.CreateOrgUnitCommand{},
		this.OrgUnitSvc.CreateOrgUnit,
	)
}

func (this OrgUnitRest) DeleteOrgUnit(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate[DeleteOrgUnitRequest, DeleteOrgUnitResponse](
		"delete org unit",
		echoCtx,
		this.OrgUnitSvc.DeleteOrgUnit,
	)
}

func (this OrgUnitRest) GetOrgUnit(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGetOne2[GetOrgUnitRequest, GetOrgUnitResponse, domain.OrganizationalUnit](
		"get org unit",
		echoCtx,
		this.OrgUnitSvc.GetOrgUnit,
	)
}

func (this OrgUnitRest) OrgUnitExists(echoCtx *echo.Context) (err error) {
	return httpserver.ServeExists[OrgUnitExistsRequest, OrgUnitExistsResponse](
		"org unit exists",
		echoCtx,
		this.OrgUnitSvc.OrgUnitExists,
	)
}

func (this OrgUnitRest) ManageOrgUnitUsers(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate[ManageOrgUnitUsersRequest, ManageOrgUnitUsersResponse](
		"manage org unit users",
		echoCtx,
		this.OrgUnitSvc.ManageOrgUnitUsers,
	)
}

func (this OrgUnitRest) SearchOrgUnits(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch[SearchOrgUnitsRequest, SearchOrgUnitsResponse, domain.OrganizationalUnit](
		"search org units",
		echoCtx,
		this.OrgUnitSvc.SearchOrgUnits,
	)
}

func (this OrgUnitRest) UpdateOrgUnit(echoCtx *echo.Context) (err error) {
	return httpserver.ServeUpdate[UpdateOrgUnitRequest, UpdateOrgUnitResponse](
		"update org unit",
		echoCtx,
		&it.UpdateOrgUnitCommand{},
		this.OrgUnitSvc.UpdateOrgUnit,
	)
}

/*
 * Non-CRUD APIs
 */

func (this OrgUnitRest) GetModelSchema(echoCtx *echo.Context) (err error) {
	schema := dmodel.MustGetSchema(domain.OrganizationalUnitSchemaName)
	echoCtx.JSON(http.StatusOK, schema.ToSimplized())
	return nil
}
