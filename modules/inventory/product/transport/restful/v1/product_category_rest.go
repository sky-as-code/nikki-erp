package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itProductCategory "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/productcategory"
)

type productCategoryRestParams struct {
	dig.In

	ProductCategorySvc itProductCategory.ProductCategoryService
}

func NewProductCategoryRest(params productCategoryRestParams) *ProductCategoryRest {
	return &ProductCategoryRest{
		ProductCategorySvc: params.ProductCategorySvc,
	}
}

type ProductCategoryRest struct {
	httpserver.RestBase
	ProductCategorySvc itProductCategory.ProductCategoryService
}

func (this ProductCategoryRest) Create(echoCtx *echo.Context) (err error) {
	return httpserver.ServeCreate(
		"create product category",
		echoCtx,
		&itProductCategory.CreateProductCategoryCommand{},
		this.ProductCategorySvc.CreateProductCategory,
	)
}

func (this ProductCategoryRest) Delete(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"delete product category",
		echoCtx,
		this.ProductCategorySvc.DeleteProductCategory,
	)
}

func (this ProductCategoryRest) Exists(echoCtx *echo.Context) (err error) {
	return httpserver.ServeExists(
		"product category exists",
		echoCtx,
		this.ProductCategorySvc.ProductCategoryExists,
	)
}

func (this ProductCategoryRest) GetOne(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGetOne(
		"get product category",
		echoCtx,
		this.ProductCategorySvc.GetProductCategory,
	)
}

func (this ProductCategoryRest) Search(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch(
		"search product categories",
		echoCtx,
		this.ProductCategorySvc.SearchProductCategories,
		true,
	)
}

func (this ProductCategoryRest) Update(echoCtx *echo.Context) (err error) {
	return httpserver.ServeUpdate(
		"update product category",
		echoCtx,
		&itProductCategory.UpdateProductCategoryCommand{},
		this.ProductCategorySvc.UpdateProductCategory,
	)
}
