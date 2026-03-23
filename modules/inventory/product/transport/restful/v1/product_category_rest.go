package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
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

func (this ProductCategoryRest) CreateProductCategory(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create product category"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.ProductCategorySvc.CreateProductCategory,
		func(request CreateProductCategoryRequest) itProductCategory.CreateProductCategoryCommand {
			return itProductCategory.CreateProductCategoryCommand(request)
		},
		func(result itProductCategory.CreateProductCategoryResult) CreateProductCategoryResponse {
			response := CreateProductCategoryResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonCreated,
	)
	return err
}

func (this ProductCategoryRest) UpdateProductCategory(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update product category"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.ProductCategorySvc.UpdateProductCategory,
		func(request UpdateProductCategoryRequest) itProductCategory.UpdateProductCategoryCommand {
			return itProductCategory.UpdateProductCategoryCommand(request)
		},
		func(result itProductCategory.UpdateProductCategoryResult) UpdateProductCategoryResponse {
			response := UpdateProductCategoryResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this ProductCategoryRest) DeleteProductCategory(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete product category"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.ProductCategorySvc.DeleteProductCategory,
		func(request DeleteProductCategoryRequest) itProductCategory.DeleteProductCategoryCommand {
			return itProductCategory.DeleteProductCategoryCommand(request)
		},
		func(result itProductCategory.DeleteProductCategoryResult) DeleteProductCategoryResponse {
			response := DeleteProductCategoryResponse{}
			response.FromNonEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this ProductCategoryRest) GetProductCategoryById(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get product category by id"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.ProductCategorySvc.GetProductCategoryById,
		func(request GetProductCategoryByIdRequest) itProductCategory.GetProductCategoryByIdQuery {
			return itProductCategory.GetProductCategoryByIdQuery(request)
		},
		func(result itProductCategory.GetProductCategoryByIdResult) GetProductCategoryByIdResponse {
			response := GetProductCategoryByIdResponse{}
			response.FromProductCategory(*result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this ProductCategoryRest) SearchProductCategories(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search product categories"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.ProductCategorySvc.SearchProductCategories,
		func(request SearchProductCategoriesRequest) itProductCategory.SearchProductCategoriesQuery {
			return itProductCategory.SearchProductCategoriesQuery(request)
		},
		func(result itProductCategory.SearchProductCategoriesResult) SearchProductCategoriesResponse {
			response := SearchProductCategoriesResponse{}
			response.FromResult(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}
