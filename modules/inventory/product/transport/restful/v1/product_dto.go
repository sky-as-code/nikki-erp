package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/product"
)

type CreateProductRequest = itProduct.CreateProductCommand
type CreateProductResponse = httpserver.RestCreateResponse

type UpdateProductRequest = itProduct.UpdateProductCommand
type UpdateProductResponse = httpserver.RestMutateResponse

type DeleteProductRequest = itProduct.DeleteProductCommand
type DeleteProductResponse = httpserver.RestDeleteResponse2

type GetProductRequest = itProduct.GetProductQuery
type GetProductResponse = dmodel.DynamicFields

type ProductExistsRequest = itProduct.ProductExistsQuery
type ProductExistsResponse = dyn.ExistsResultData

type SetProductIsArchivedRequest = itProduct.SetProductIsArchivedCommand
type SetProductIsArchivedResponse = httpserver.RestMutateResponse

type SearchProductsRequest = itProduct.SearchProductsQuery
type SearchProductsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
