package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

type CreateProductVariantRequest = itProduct.CreateProductVariantCommand
type CreateProductVariantResponse = httpserver.RestCreateResponse

type UpdateProductVariantRequest = itProduct.UpdateProductVariantCommand
type UpdateProductVariantResponse = httpserver.RestMutateResponse

type DeleteProductVariantRequest = itProduct.DeleteProductVariantCommand
type DeleteProductVariantResponse = httpserver.RestMutateResponse

type SetProductVariantArchivedRequest = itProduct.SetProductVariantArchivedCommand
type SetProductVariantArchivedResponse = httpserver.RestMutateResponse

type GetProductVariantRequest = itProduct.GetProductVariantByIdQuery
type GetProductVariantResponse = httpserver.RestGetOneResponse[dmodel.DynamicFields]

type ProductVariantExistsRequest = itProduct.ProductVariantExistsQuery
type ProductVariantExistsResponse = dyn.ExistsResultData

type SearchProductVariantsRequest = itProduct.SearchProductVariantRowsQuery
type SearchProductVariantsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
