package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

type CreateBrandRequest = itProduct.CreateBrandCommand
type CreateBrandResponse = httpserver.RestCreateResponse

type UpdateBrandRequest = itProduct.UpdateBrandCommand
type UpdateBrandResponse = httpserver.RestMutateResponse

type DeleteBrandRequest = itProduct.DeleteBrandCommand
type DeleteBrandResponse = httpserver.RestMutateResponse

type SetBrandArchivedRequest = itProduct.SetBrandArchivedCommand
type SetBrandArchivedResponse = httpserver.RestMutateResponse

type GetBrandRequest = itProduct.GetBrandByIdQuery
type GetBrandResponse = httpserver.RestGetOneResponse[dmodel.DynamicFields]

type BrandExistsRequest = itProduct.BrandExistsQuery
type BrandExistsResponse = dyn.ExistsResultData

type SearchBrandsRequest = itProduct.SearchBrandsQuery
type SearchBrandsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
