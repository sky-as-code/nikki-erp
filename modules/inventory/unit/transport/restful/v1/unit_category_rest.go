package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itUnitCategory "github.com/sky-as-code/nikki-erp/modules/inventory/unit/interfaces/unitcategory"
)

type unitCategoryRestParams struct {
	dig.In

	UnitCategorySvc itUnitCategory.UnitCategoryService
}

func NewUnitCategoryRest(params unitCategoryRestParams) *UnitCategoryRest {
	return &UnitCategoryRest{
		UnitCategorySvc: params.UnitCategorySvc,
	}
}

type UnitCategoryRest struct {
	httpserver.RestBase
	UnitCategorySvc itUnitCategory.UnitCategoryService
}

func (r UnitCategoryRest) CreateUnitCategory(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create unit category"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, r.UnitCategorySvc.CreateUnitCategory,
		func(request CreateUnitCategoryRequest) itUnitCategory.CreateUnitCategoryCommand {
			return itUnitCategory.CreateUnitCategoryCommand(request)
		},
		func(result itUnitCategory.CreateUnitCategoryResult) CreateUnitCategoryResponse {
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
		func(request UpdateUnitCategoryRequest) itUnitCategory.UpdateUnitCategoryCommand {
			return itUnitCategory.UpdateUnitCategoryCommand(request)
		},
		func(result itUnitCategory.UpdateUnitCategoryResult) UpdateUnitCategoryResponse {
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
		func(request DeleteUnitCategoryRequest) itUnitCategory.DeleteUnitCategoryCommand {
			return itUnitCategory.DeleteUnitCategoryCommand(request)
		},
		func(result itUnitCategory.DeleteUnitCategoryResult) DeleteUnitCategoryResponse {
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
		func(request GetUnitCategoryByIdRequest) itUnitCategory.GetUnitCategoryByIdQuery {
			return itUnitCategory.GetUnitCategoryByIdQuery(request)
		},
		func(result itUnitCategory.GetUnitCategoryByIdResult) GetUnitCategoryByIdResponse {
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
		func(request SearchUnitCategoriesRequest) itUnitCategory.SearchUnitCategoriesQuery {
			return itUnitCategory.SearchUnitCategoriesQuery(request)
		},
		func(result itUnitCategory.SearchUnitCategoriesResult) SearchUnitCategoriesResponse {
			response := SearchUnitCategoriesResponse{}
			response.FromResult(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}
