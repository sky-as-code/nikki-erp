package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/product"
)

type productRestParams struct {
	dig.In

	ProductSvc itProduct.ProductService
}

func NewProductRest(params productRestParams) *ProductRest {
	return &ProductRest{
		ProductSvc: params.ProductSvc,
	}
}

type ProductRest struct {
	httpserver.RestBase
	ProductSvc itProduct.ProductService
}

func (this ProductRest) CreateProduct(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create product"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.ProductSvc.CreateProduct,
		func(request CreateProductRequest) itProduct.CreateProductCommand {
			return itProduct.CreateProductCommand(request)
		},
		func(result itProduct.CreateProductResult) CreateProductResponse {
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
		func(request UpdateProductRequest) itProduct.UpdateProductCommand {
			return itProduct.UpdateProductCommand(request)
		},
		func(result itProduct.UpdateProductResult) UpdateProductResponse {
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
		func(request DeleteProductRequest) itProduct.DeleteProductCommand {
			return itProduct.DeleteProductCommand(request)
		},
		func(result itProduct.DeleteProductResult) DeleteProductResponse {
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
		func(request GetProductByIdRequest) itProduct.GetProductByIdQuery {
			return itProduct.GetProductByIdQuery(request)
		},
		func(result itProduct.GetProductByIdResult) GetProductByIdResponse {
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
		func(request SearchProductsRequest) itProduct.SearchProductsQuery {
			return itProduct.SearchProductsQuery(request)
		},
		func(result itProduct.SearchProductsResult) SearchProductsResponse {
			response := SearchProductsResponse{}
			response.FromResult(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}
