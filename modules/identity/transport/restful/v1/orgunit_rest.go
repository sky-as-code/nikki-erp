package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
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

func (this OrgUnitRest) Create(echoCtx echo.Context) (err error) {
	return httpserver.ServeCreate(
		"create org unit",
		echoCtx,
		&it.CreateOrgUnitCommand{},
		this.OrgUnitSvc.CreateOrgUnit,
	)
}

func (this OrgUnitRest) Delete(echoCtx echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"delete org unit",
		echoCtx,
		this.OrgUnitSvc.DeleteOrgUnit,
	)
}

func (this OrgUnitRest) GetOne(echoCtx echo.Context) (err error) {
	return httpserver.ServeGetOne(
		"get org unit",
		echoCtx,
		this.OrgUnitSvc.GetOrgUnit,
	)
}

func (this OrgUnitRest) Exists(echoCtx echo.Context) (err error) {
	return httpserver.ServeExists(
		"org unit exists",
		echoCtx,
		this.OrgUnitSvc.OrgUnitExists,
	)
}

func (this OrgUnitRest) ManageOrgUnitUsers(echoCtx echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"manage org unit users",
		echoCtx,
		this.OrgUnitSvc.ManageOrgUnitUsers,
	)
}

func (this OrgUnitRest) Search(echoCtx echo.Context) (err error) {
	return httpserver.ServeSearch(
		"search org units",
		echoCtx,
		this.OrgUnitSvc.SearchOrgUnits,
		true,
	)
}

func (this OrgUnitRest) Update(echoCtx echo.Context) (err error) {
	return httpserver.ServeUpdate(
		"update org unit",
		echoCtx,
		&it.UpdateOrgUnitCommand{},
		this.OrgUnitSvc.UpdateOrgUnit,
	)
}
