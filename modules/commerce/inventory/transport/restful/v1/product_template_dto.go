package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
)

type CreateProductTemplateRequest = itProduct.CreateProductTemplateCommand
type CreateProductTemplateResponse = httpserver.RestCreateResponse

type UpdateProductTemplateRequest = itProduct.UpdateProductTemplateCommand
type UpdateProductTemplateResponse = httpserver.RestMutateResponse

type DeleteProductTemplateRequest = itProduct.DeleteProductTemplateCommand
type DeleteProductTemplateResponse = httpserver.RestMutateResponse

type SetProductTemplateArchivedRequest = itProduct.SetProductTemplateArchivedCommand
type SetProductTemplateArchivedResponse = httpserver.RestMutateResponse

type GetProductTemplateRequest = itProduct.GetProductTemplateByIdQuery
type GetProductTemplateResponse = httpserver.RestGetOneResponse[dmodel.DynamicFields]

type ProductTemplateExistsRequest = itProduct.ProductTemplateExistsQuery
type ProductTemplateExistsResponse = dyn.ExistsResultData

type SearchProductTemplatesRequest = itProduct.SearchProductTemplatesQuery
type SearchProductTemplatesResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
