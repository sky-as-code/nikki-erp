package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/warehouse"
)

type CreateWarehouseRequest = itWarehouse.CreateWarehouseRowCommand
type CreateWarehouseResponse = httpserver.RestCreateResponse

type UpdateWarehouseRequest = itWarehouse.UpdateWarehouseCommand
type UpdateWarehouseResponse = httpserver.RestMutateResponse

type DeleteWarehouseRequest = itWarehouse.DeleteWarehouseCommand
type DeleteWarehouseResponse = httpserver.RestMutateResponse

type SetWarehouseArchivedRequest = itWarehouse.SetWarehouseArchivedCommand
type SetWarehouseArchivedResponse = httpserver.RestMutateResponse

type GetWarehouseRequest = itWarehouse.GetWarehouseByIdQuery
type GetWarehouseResponse = httpserver.RestGetOneResponse[dmodel.DynamicFields]

type WarehouseExistsRequest = itWarehouse.WarehouseExistsQuery
type WarehouseExistsResponse = dyn.ExistsResultData

type SearchWarehousesRequest = itWarehouse.SearchWarehousesQuery
type SearchWarehousesResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
