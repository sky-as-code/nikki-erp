package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

type CreateStockOperationTypeRequest = itStock.CreateStockOperationTypeCommand
type CreateStockOperationTypeResponse = httpserver.RestCreateResponse

type UpdateStockOperationTypeRequest = itStock.UpdateStockOperationTypeCommand
type UpdateStockOperationTypeResponse = httpserver.RestMutateResponse

type DeleteStockOperationTypeRequest = itStock.DeleteStockOperationTypeCommand
type DeleteStockOperationTypeResponse = httpserver.RestMutateResponse

type SetStockOperationTypeArchivedRequest = itStock.SetStockOperationTypeArchivedCommand
type SetStockOperationTypeArchivedResponse = httpserver.RestMutateResponse

type GetStockOperationTypeRequest = itStock.GetStockOperationTypeByIdQuery
type GetStockOperationTypeResponse = httpserver.RestGetOneResponse[dmodel.DynamicFields]

type StockOperationTypeExistsRequest = itStock.StockOperationTypeExistsQuery
type StockOperationTypeExistsResponse = dyn.ExistsResultData

type SearchStockOperationTypesRequest = itStock.SearchStockOperationTypesQuery
type SearchStockOperationTypesResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
