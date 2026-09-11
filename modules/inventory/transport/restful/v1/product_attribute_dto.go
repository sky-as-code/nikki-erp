package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

type CreateProductAttributeRequest = itProduct.CreateProductAttributeCommand
type CreateProductAttributeResponse = httpserver.RestCreateResponse

type UpdateProductAttributeRequest = itProduct.UpdateProductAttributeCommand
type UpdateProductAttributeResponse = httpserver.RestMutateResponse

type DeleteProductAttributeRequest = itProduct.DeleteProductAttributeCommand
type DeleteProductAttributeResponse = httpserver.RestMutateResponse

type SetProductAttributeArchivedRequest = itProduct.SetProductAttributeArchivedCommand
type SetProductAttributeArchivedResponse = httpserver.RestMutateResponse

type GetProductAttributeRequest = itProduct.GetProductAttributeByIdQuery
type GetProductAttributeResponse = httpserver.RestGetOneResponse[dmodel.DynamicFields]

type ProductAttributeExistsRequest = itProduct.ProductAttributeExistsQuery
type ProductAttributeExistsResponse = dyn.ExistsResultData

type SearchProductAttributesRequest = itProduct.SearchProductAttributesQuery
type SearchProductAttributesResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
