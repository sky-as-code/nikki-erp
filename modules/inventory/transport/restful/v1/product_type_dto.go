package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

type CreateProductTypeRequest = itProduct.CreateProductTypeCommand
type CreateProductTypeResponse = httpserver.RestCreateResponse

type UpdateProductTypeRequest = itProduct.UpdateProductTypeCommand
type UpdateProductTypeResponse = httpserver.RestMutateResponse

type DeleteProductTypeRequest = itProduct.DeleteProductTypeCommand
type DeleteProductTypeResponse = httpserver.RestMutateResponse

type SetProductTypeArchivedRequest = itProduct.SetProductTypeArchivedCommand
type SetProductTypeArchivedResponse = httpserver.RestMutateResponse

type GetProductTypeRequest = itProduct.GetProductTypeByIdQuery
type GetProductTypeResponse = httpserver.RestGetOneResponse[dmodel.DynamicFields]

type ProductTypeExistsRequest = itProduct.ProductTypeExistsQuery
type ProductTypeExistsResponse = dyn.ExistsResultData

type SearchProductTypesRequest = itProduct.SearchProductTypesQuery
type SearchProductTypesResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
