package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itUnit "github.com/sky-as-code/nikki-erp/modules/inventory/unit/interfaces/unit"
)

type unitRestParams struct {
	dig.In

	UnitSvc itUnit.UnitService
}

func NewUnitRest(params unitRestParams) *UnitRest {
	return &UnitRest{
		UnitSvc: params.UnitSvc,
	}
}

type UnitRest struct {
	httpserver.RestBase
	UnitSvc itUnit.UnitService
}

func (this UnitRest) Create(echoCtx echo.Context) (err error) {
	return httpserver.ServeCreate(
		"create unit",
		echoCtx,
		&itUnit.CreateUnitCommand{},
		this.UnitSvc.CreateUnit,
	)
}

func (this UnitRest) Delete(echoCtx echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"delete unit",
		echoCtx,
		this.UnitSvc.DeleteUnit,
	)
}

func (this UnitRest) Exists(echoCtx echo.Context) (err error) {
	return httpserver.ServeExists(
		"unit exists",
		echoCtx,
		this.UnitSvc.UnitExists,
	)
}

func (this UnitRest) GetOne(echoCtx echo.Context) (err error) {
	return httpserver.ServeGetOne(
		"get unit",
		echoCtx,
		this.UnitSvc.GetUnit,
	)
}

func (this UnitRest) Search(echoCtx echo.Context) (err error) {
	return httpserver.ServeSearch(
		"search units",
		echoCtx,
		this.UnitSvc.SearchUnits,
		true,
	)
}

func (this UnitRest) Update(echoCtx echo.Context) (err error) {
	return httpserver.ServeUpdate(
		"update unit",
		echoCtx,
		&itUnit.UpdateUnitCommand{},
		this.UnitSvc.UpdateUnit,
	)
}
