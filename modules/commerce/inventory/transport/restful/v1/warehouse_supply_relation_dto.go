package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/warehouse"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
)

type CreateWarehouseSupplyRelationRequest = itWarehouse.CreateWarehouseSupplyRelationCommand
type CreateWarehouseSupplyRelationResponse = httpserver.RestCreateResponse

type UpdateWarehouseSupplyRelationRequest = itWarehouse.UpdateWarehouseSupplyRelationCommand
type UpdateWarehouseSupplyRelationResponse = httpserver.RestMutateResponse

type DeleteWarehouseSupplyRelationRequest = itWarehouse.DeleteWarehouseSupplyRelationCommand
type DeleteWarehouseSupplyRelationResponse = httpserver.RestMutateResponse

type SetWarehouseSupplyRelationArchivedRequest = itWarehouse.SetWarehouseSupplyRelationArchivedCommand
type SetWarehouseSupplyRelationArchivedResponse = httpserver.RestMutateResponse

type GetWarehouseSupplyRelationRequest = itWarehouse.GetWarehouseSupplyRelationByIdQuery
type GetWarehouseSupplyRelationResponse = httpserver.RestGetOneResponse[dmodel.DynamicFields]

type WarehouseSupplyRelationExistsRequest = itWarehouse.WarehouseSupplyRelationExistsQuery
type WarehouseSupplyRelationExistsResponse = dyn.ExistsResultData

type SearchWarehouseSupplyRelationsRequest = itWarehouse.SearchWarehouseSupplyRelationsQuery
type SearchWarehouseSupplyRelationsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
