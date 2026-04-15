package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/unit"
)

type unitRestParams struct {
	dig.In

	UnitSvc it.UnitService
}

func NewUnitRest(params unitRestParams) *UnitRest {
	return &UnitRest{
		unitSvc: params.UnitSvc,
	}
}

type UnitRest struct {
	unitSvc it.UnitService
}

func (this UnitRest) Create(echoCtx *echo.Context) (err error) {
	return httpserver.ServeCreate(
		"create unit",
		echoCtx,
		&it.CreateUnitCommand{},
		this.unitSvc.CreateUnit,
	)
}

func (this UnitRest) Delete(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"delete unit",
		echoCtx,
		this.unitSvc.DeleteUnit,
	)
}

func (this UnitRest) Exists(echoCtx *echo.Context) (err error) {
	return httpserver.ServeExists(
		"unit exists",
		echoCtx,
		this.unitSvc.UnitExists,
	)
}

func (this UnitRest) GetOne(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGetOne(
		"get unit",
		echoCtx,
		this.unitSvc.GetUnit,
	)
}

func (this UnitRest) Search(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch(
		"search units",
		echoCtx,
		this.unitSvc.SearchUnits,
		true,
	)
}

func (this UnitRest) Update(echoCtx *echo.Context) (err error) {
	return httpserver.ServeUpdate(
		"update unit",
		echoCtx,
		&it.UpdateUnitCommand{},
		this.unitSvc.UpdateUnit,
	)
}
