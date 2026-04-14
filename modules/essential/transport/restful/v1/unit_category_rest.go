package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/unitcategory"
)

type unitCategoryRestParams struct {
	dig.In

	UnitCategorySvc it.UnitCategoryService
}

func NewUnitCategoryRest(params unitCategoryRestParams) *UnitCategoryRest {
	return &UnitCategoryRest{
		unitCatSvc: params.UnitCategorySvc,
	}
}

type UnitCategoryRest struct {
	unitCatSvc it.UnitCategoryService
}

func (this UnitCategoryRest) Create(echoCtx echo.Context) (err error) {
	return httpserver.ServeCreate(
		"create unit category",
		echoCtx,
		&it.CreateUnitCategoryCommand{},
		this.unitCatSvc.CreateUnitCategory,
	)
}

func (this UnitCategoryRest) Update(echoCtx echo.Context) (err error) {
	return httpserver.ServeUpdate(
		"update unit category",
		echoCtx,
		&it.UpdateUnitCategoryCommand{},
		this.unitCatSvc.UpdateUnitCategory,
	)
}

func (this UnitCategoryRest) Delete(echoCtx echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"delete unit category",
		echoCtx,
		this.unitCatSvc.DeleteUnitCategory,
	)
}

func (this UnitCategoryRest) GetOne(echoCtx echo.Context) (err error) {
	return httpserver.ServeGetOne(
		"get unit category",
		echoCtx,
		this.unitCatSvc.GetUnitCategory,
	)
}

func (this UnitCategoryRest) Search(echoCtx echo.Context) (err error) {
	return httpserver.ServeSearch(
		"search unit categories",
		echoCtx,
		this.unitCatSvc.SearchUnitCategories,
		true,
	)
}

func (this UnitCategoryRest) Exists(echoCtx echo.Context) (err error) {
	return httpserver.ServeExists(
		"unit category exists",
		echoCtx,
		this.unitCatSvc.UnitCategoryExists,
	)
}
