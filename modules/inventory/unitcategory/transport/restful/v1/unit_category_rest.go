package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/inventory/unitcategory/interfaces"
)

type unitCategoryRestParams struct {
	dig.In

	UnitCategorySvc it.UnitCategoryService
}

func NewUnitCategoryRest(params unitCategoryRestParams) *UnitCategoryRest {
	return &UnitCategoryRest{
		UnitCategorySvc: params.UnitCategorySvc,
	}
}

type UnitCategoryRest struct {
	httpserver.RestBase
	UnitCategorySvc it.UnitCategoryService
}

func (r UnitCategoryRest) CreateUnitCategory(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create unit category"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, r.UnitCategorySvc.CreateUnitCategory,
		func(request CreateUnitCategoryRequest) it.CreateUnitCategoryCommand {
			return it.CreateUnitCategoryCommand(request)
		},
		func(result it.CreateUnitCategoryResult) CreateUnitCategoryResponse {
			response := CreateUnitCategoryResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonCreated,
	)
	return err
}

func (r UnitCategoryRest) UpdateUnitCategory(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update unit category"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, r.UnitCategorySvc.UpdateUnitCategory,
		func(request UpdateUnitCategoryRequest) it.UpdateUnitCategoryCommand {
			return it.UpdateUnitCategoryCommand(request)
		},
		func(result it.UpdateUnitCategoryResult) UpdateUnitCategoryResponse {
			response := UpdateUnitCategoryResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (r UnitCategoryRest) DeleteUnitCategory(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete unit category"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, r.UnitCategorySvc.DeleteUnitCategory,
		func(request DeleteUnitCategoryRequest) it.DeleteUnitCategoryCommand {
			return it.DeleteUnitCategoryCommand(request)
		},
		func(result it.DeleteUnitCategoryResult) DeleteUnitCategoryResponse {
			response := DeleteUnitCategoryResponse{}
			response.FromNonEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (r UnitCategoryRest) GetUnitCategoryById(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get unit category by id"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, r.UnitCategorySvc.GetUnitCategoryById,
		func(request GetUnitCategoryByIdRequest) it.GetUnitCategoryByIdQuery {
			return it.GetUnitCategoryByIdQuery(request)
		},
		func(result it.GetUnitCategoryByIdResult) GetUnitCategoryByIdResponse {
			response := GetUnitCategoryByIdResponse{}
			response.FromUnitCategory(*result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (r UnitCategoryRest) SearchUnitCategories(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search unit categories"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, r.UnitCategorySvc.SearchUnitCategories,
		func(request SearchUnitCategoriesRequest) it.SearchUnitCategoriesQuery {
			return it.SearchUnitCategoriesQuery(request)
		},
		func(result it.SearchUnitCategoriesResult) SearchUnitCategoriesResponse {
			response := SearchUnitCategoriesResponse{}
			response.FromResult(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}
