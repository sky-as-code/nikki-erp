package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

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

func (this ProductRest) Create(echoCtx echo.Context) (err error) {
	return httpserver.ServeCreate(
		"create product",
		echoCtx,
		&itProduct.CreateProductCommand{},
		this.ProductSvc.CreateProduct,
	)
}

func (this ProductRest) Update(echoCtx echo.Context) (err error) {
	return httpserver.ServeUpdate(
		"update product",
		echoCtx,
		&itProduct.UpdateProductCommand{},
		this.ProductSvc.UpdateProduct,
	)
}

func (this ProductRest) Delete(echoCtx echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"delete product",
		echoCtx,
		this.ProductSvc.DeleteProduct,
	)
}

func (this ProductRest) GetOne(echoCtx echo.Context) (err error) {
	return httpserver.ServeGetOne(
		"get product",
		echoCtx,
		this.ProductSvc.GetProduct,
	)
}

func (this ProductRest) Exists(echoCtx echo.Context) (err error) {
	return httpserver.ServeExists(
		"product exists",
		echoCtx,
		this.ProductSvc.ProductExists,
	)
}

func (this ProductRest) SetIsArchived(echoCtx echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"set product is_archived",
		echoCtx,
		this.ProductSvc.SetProductIsArchived,
	)
}

func (this ProductRest) Search(echoCtx echo.Context) (err error) {
	return httpserver.ServeSearch(
		"search products",
		echoCtx,
		this.ProductSvc.SearchProducts,
		true,
	)
}
