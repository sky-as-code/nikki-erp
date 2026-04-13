package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itProductCategory "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/productcategory"
)

type CreateProductCategoryRequest = itProductCategory.CreateProductCategoryCommand
type CreateProductCategoryResponse = httpserver.RestCreateResponse

type UpdateProductCategoryRequest = itProductCategory.UpdateProductCategoryCommand
type UpdateProductCategoryResponse = httpserver.RestMutateResponse

type DeleteProductCategoryRequest = itProductCategory.DeleteProductCategoryCommand
type DeleteProductCategoryResponse = httpserver.RestDeleteResponse2

type GetProductCategoryRequest = itProductCategory.GetProductCategoryQuery
type GetProductCategoryResponse = dmodel.DynamicFields

type SearchProductCategoriesRequest = itProductCategory.SearchProductCategoriesQuery
type SearchProductCategoriesResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
