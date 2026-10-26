package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/warehouse"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
)

type CreateInventoryLocationRequest = itWarehouse.CreateInventoryLocationCommand
type CreateInventoryLocationResponse = httpserver.RestCreateResponse

type UpdateInventoryLocationRequest = itWarehouse.UpdateInventoryLocationCommand
type UpdateInventoryLocationResponse = httpserver.RestMutateResponse

type DeleteInventoryLocationRequest = itWarehouse.DeleteInventoryLocationCommand
type DeleteInventoryLocationResponse = httpserver.RestMutateResponse

type SetInventoryLocationArchivedRequest = itWarehouse.SetInventoryLocationArchivedCommand
type SetInventoryLocationArchivedResponse = httpserver.RestMutateResponse

type GetInventoryLocationRequest = itWarehouse.GetInventoryLocationByIdQuery
type GetInventoryLocationResponse = httpserver.RestGetOneResponse[dmodel.DynamicFields]

type InventoryLocationExistsRequest = itWarehouse.InventoryLocationExistsQuery
type InventoryLocationExistsResponse = dyn.ExistsResultData

type SearchInventoryLocationsRequest = itWarehouse.SearchInventoryLocationsQuery
type SearchInventoryLocationsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
