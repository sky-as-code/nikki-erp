package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

type CreateProductCategoryRequest = itProduct.CreateProductCategoryCommand
type CreateProductCategoryResponse = httpserver.RestCreateResponse

type UpdateProductCategoryRequest = itProduct.UpdateProductCategoryCommand
type UpdateProductCategoryResponse = httpserver.RestMutateResponse

type DeleteProductCategoryRequest = itProduct.DeleteProductCategoryCommand
type DeleteProductCategoryResponse = httpserver.RestMutateResponse

type SetProductCategoryArchivedRequest = itProduct.SetProductCategoryArchivedCommand
type SetProductCategoryArchivedResponse = httpserver.RestMutateResponse

type GetProductCategoryRequest = itProduct.GetProductCategoryByIdQuery
type GetProductCategoryResponse = httpserver.RestGetOneResponse[dmodel.DynamicFields]

type ProductCategoryExistsRequest = itProduct.ProductCategoryExistsQuery
type ProductCategoryExistsResponse = dyn.ExistsResultData

type SearchProductCategoriesRequest = itProduct.SearchProductCategoriesQuery
type SearchProductCategoriesResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
