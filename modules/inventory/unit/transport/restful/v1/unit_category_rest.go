package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

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

func (this UnitCategoryRest) Create(echoCtx echo.Context) (err error) {
	return httpserver.ServeCreate(
		"create unit category",
		echoCtx,
		&itUnitCategory.CreateUnitCategoryCommand{},
		this.UnitCategorySvc.CreateUnitCategory,
	)
}

func (this UnitCategoryRest) Update(echoCtx echo.Context) (err error) {
	return httpserver.ServeUpdate(
		"update unit category",
		echoCtx,
		&itUnitCategory.UpdateUnitCategoryCommand{},
		this.UnitCategorySvc.UpdateUnitCategory,
	)
}

func (this UnitCategoryRest) Delete(echoCtx echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"delete unit category",
		echoCtx,
		this.UnitCategorySvc.DeleteUnitCategory,
	)
}

func (this UnitCategoryRest) GetOne(echoCtx echo.Context) (err error) {
	return httpserver.ServeGetOne(
		"get unit category",
		echoCtx,
		this.UnitCategorySvc.GetUnitCategory,
	)
}

func (this UnitCategoryRest) Search(echoCtx echo.Context) (err error) {
	return httpserver.ServeSearch(
		"search unit categories",
		echoCtx,
		this.UnitCategorySvc.SearchUnitCategories,
		true,
	)
}

func (this UnitCategoryRest) Exists(echoCtx echo.Context) (err error) {
	return httpserver.ServeExists(
		"unit category exists",
		echoCtx,
		this.UnitCategorySvc.UnitCategoryExists,
	)
}
