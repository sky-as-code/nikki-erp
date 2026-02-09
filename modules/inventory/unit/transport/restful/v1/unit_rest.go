package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
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

func (this UnitRest) CreateUnit(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create unit"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.UnitSvc.CreateUnit,
		func(request CreateUnitRequest) itUnit.CreateUnitCommand {
			return itUnit.CreateUnitCommand(request)
		},
		func(result itUnit.CreateUnitResult) CreateUnitResponse {
			response := CreateUnitResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonCreated,
	)
	return err
}

func (this UnitRest) UpdateUnit(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update unit"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.UnitSvc.UpdateUnit,
		func(request UpdateUnitRequest) itUnit.UpdateUnitCommand {
			return itUnit.UpdateUnitCommand(request)
		},
		func(result itUnit.UpdateUnitResult) UpdateUnitResponse {
			response := UpdateUnitResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this UnitRest) DeleteUnit(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete unit"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.UnitSvc.DeleteUnit,
		func(request DeleteUnitRequest) itUnit.DeleteUnitCommand {
			return itUnit.DeleteUnitCommand(request)
		},
		func(result itUnit.DeleteUnitResult) DeleteUnitResponse {
			response := DeleteUnitResponse{}
			response.FromNonEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this UnitRest) GetUnitById(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get unit by id"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.UnitSvc.GetUnitById,
		func(request GetUnitByIdRequest) itUnit.GetUnitByIdQuery {
			return itUnit.GetUnitByIdQuery(request)
		},
		func(result itUnit.GetUnitByIdResult) GetUnitByIdResponse {
			response := GetUnitByIdResponse{}
			response.FromUnit(*result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this UnitRest) SearchUnits(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search units"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.UnitSvc.SearchUnits,
		func(request SearchUnitsRequest) itUnit.SearchUnitsQuery {
			return itUnit.SearchUnitsQuery(request)
		},
		func(result itUnit.SearchUnitsResult) SearchUnitsResponse {
			response := SearchUnitsResponse{}
			response.FromResult(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}
