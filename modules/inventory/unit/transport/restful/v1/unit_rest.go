package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/inventory/unit/interfaces"
)

type unitRestParams struct {
	dig.In

	UnitSvc it.UnitService
}

func NewUnitRest(params unitRestParams) *UnitRest {
	return &UnitRest{
		UnitSvc: params.UnitSvc,
	}
}

type UnitRest struct {
	httpserver.RestBase
	UnitSvc it.UnitService
}

func (this UnitRest) CreateUnit(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create unit"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.UnitSvc.CreateUnit,
		func(request CreateUnitRequest) it.CreateUnitCommand {
			return it.CreateUnitCommand(request)
		},
		func(result it.CreateUnitResult) CreateUnitResponse {
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
		func(request UpdateUnitRequest) it.UpdateUnitCommand {
			return it.UpdateUnitCommand(request)
		},
		func(result it.UpdateUnitResult) UpdateUnitResponse {
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
		func(request DeleteUnitRequest) it.DeleteUnitCommand {
			return it.DeleteUnitCommand(request)
		},
		func(result it.DeleteUnitResult) DeleteUnitResponse {
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
		func(request GetUnitByIdRequest) it.GetUnitByIdQuery {
			return it.GetUnitByIdQuery(request)
		},
		func(result it.GetUnitByIdResult) GetUnitByIdResponse {
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
		func(request SearchUnitsRequest) it.SearchUnitsQuery {
			return it.SearchUnitsQuery(request)
		},
		func(result it.SearchUnitsResult) SearchUnitsResponse {
			response := SearchUnitsResponse{}
			response.FromResult(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}
