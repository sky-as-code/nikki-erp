package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/warehouse"
)

type CreateStorageCategoryRequest = itWarehouse.CreateStorageCategoryCommand
type CreateStorageCategoryResponse = httpserver.RestCreateResponse

type UpdateStorageCategoryRequest = itWarehouse.UpdateStorageCategoryCommand
type UpdateStorageCategoryResponse = httpserver.RestMutateResponse

type DeleteStorageCategoryRequest = itWarehouse.DeleteStorageCategoryCommand
type DeleteStorageCategoryResponse = httpserver.RestMutateResponse

type SetStorageCategoryArchivedRequest = itWarehouse.SetStorageCategoryArchivedCommand
type SetStorageCategoryArchivedResponse = httpserver.RestMutateResponse

type GetStorageCategoryRequest = itWarehouse.GetStorageCategoryByIdQuery
type GetStorageCategoryResponse = httpserver.RestGetOneResponse[dmodel.DynamicFields]

type StorageCategoryExistsRequest = itWarehouse.StorageCategoryExistsQuery
type StorageCategoryExistsResponse = dyn.ExistsResultData

type SearchStorageCategoriesRequest = itWarehouse.SearchStorageCategoriesQuery
type SearchStorageCategoriesResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
