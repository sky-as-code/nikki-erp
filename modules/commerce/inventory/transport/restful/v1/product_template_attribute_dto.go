package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
)

type CreateProductTemplateAttributeRequest = itProduct.CreateProductTemplateAttributeCommand
type CreateProductTemplateAttributeResponse = httpserver.RestCreateResponse

type UpdateProductTemplateAttributeRequest = itProduct.UpdateProductTemplateAttributeCommand
type UpdateProductTemplateAttributeResponse = httpserver.RestMutateResponse

type DeleteProductTemplateAttributeRequest = itProduct.DeleteProductTemplateAttributeCommand
type DeleteProductTemplateAttributeResponse = httpserver.RestMutateResponse

type SetProductTemplateAttributeArchivedRequest = itProduct.SetProductTemplateAttributeArchivedCommand
type SetProductTemplateAttributeArchivedResponse = httpserver.RestMutateResponse

type GetProductTemplateAttributeRequest = itProduct.GetProductTemplateAttributeByIdQuery
type GetProductTemplateAttributeResponse = httpserver.RestGetOneResponse[dmodel.DynamicFields]

type ProductTemplateAttributeExistsRequest = itProduct.ProductTemplateAttributeExistsQuery
type ProductTemplateAttributeExistsResponse = dyn.ExistsResultData

type SearchProductTemplateAttributesRequest = itProduct.SearchProductTemplateAttributesQuery
type SearchProductTemplateAttributesResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
