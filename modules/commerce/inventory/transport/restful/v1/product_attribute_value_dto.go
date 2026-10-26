package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
)

type CreateProductAttributeValueRequest = itProduct.CreateProductAttributeValueCommand
type CreateProductAttributeValueResponse = httpserver.RestCreateResponse

type UpdateProductAttributeValueRequest = itProduct.UpdateProductAttributeValueCommand
type UpdateProductAttributeValueResponse = httpserver.RestMutateResponse

type DeleteProductAttributeValueRequest = itProduct.DeleteProductAttributeValueCommand
type DeleteProductAttributeValueResponse = httpserver.RestMutateResponse

type SetProductAttributeValueArchivedRequest = itProduct.SetProductAttributeValueArchivedCommand
type SetProductAttributeValueArchivedResponse = httpserver.RestMutateResponse

type GetProductAttributeValueRequest = itProduct.GetProductAttributeValueByIdQuery
type GetProductAttributeValueResponse = httpserver.RestGetOneResponse[dmodel.DynamicFields]

type ProductAttributeValueExistsRequest = itProduct.ProductAttributeValueExistsQuery
type ProductAttributeValueExistsResponse = dyn.ExistsResultData

type SearchProductAttributeValuesRequest = itProduct.SearchProductAttributeValuesQuery
type SearchProductAttributeValuesResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
