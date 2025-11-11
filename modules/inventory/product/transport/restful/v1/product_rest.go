package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces"
)

type productRestParams struct {
	dig.In

	ProductSvc it.ProductService
}

func NewProductRest(params productRestParams) *ProductRest {
	return &ProductRest{
		ProductSvc: params.ProductSvc,
	}
}

type ProductRest struct {
	httpserver.RestBase
	ProductSvc it.ProductService
}

func (this ProductRest) CreateProduct(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create product"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.ProductSvc.CreateProduct,
		func(request CreateProductRequest) it.CreateProductCommand {
			return it.CreateProductCommand(request)
		},
		func(result it.CreateProductResult) CreateProductResponse {
			response := CreateProductResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonCreated,
	)
	return err
}

func (this ProductRest) UpdateProduct(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update product"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.ProductSvc.UpdateProduct,
		func(request UpdateProductRequest) it.UpdateProductCommand {
			return it.UpdateProductCommand(request)
		},
		func(result it.UpdateProductResult) UpdateProductResponse {
			response := UpdateProductResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this ProductRest) DeleteProduct(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete product"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.ProductSvc.DeleteProduct,
		func(request DeleteProductRequest) it.DeleteProductCommand {
			return it.DeleteProductCommand(request)
		},
		func(result it.DeleteProductResult) DeleteProductResponse {
			response := DeleteProductResponse{}
			response.FromNonEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this ProductRest) GetProductById(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get product by id"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.ProductSvc.GetProductById,
		func(request GetProductByIdRequest) it.GetProductByIdQuery {
			return it.GetProductByIdQuery(request)
		},
		func(result it.GetProductByIdResult) GetProductByIdResponse {
			response := GetProductByIdResponse{}
			response.FromProduct(*result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this ProductRest) SearchProducts(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search products"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.ProductSvc.SearchProducts,
		func(request SearchProductsRequest) it.SearchProductsQuery {
			return it.SearchProductsQuery(request)
		},
		func(result it.SearchProductsResult) SearchProductsResponse {
			response := SearchProductsResponse{}
			response.FromResult(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}
