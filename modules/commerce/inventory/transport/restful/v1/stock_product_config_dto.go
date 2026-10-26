package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

type CreateStockProductConfigRequest = itStock.CreateStockProductConfigCommand
type CreateStockProductConfigResponse = httpserver.RestCreateResponse

type UpdateStockProductConfigRequest = itStock.UpdateStockProductConfigCommand
type UpdateStockProductConfigResponse = httpserver.RestMutateResponse

type DeleteStockProductConfigRequest = itStock.DeleteStockProductConfigCommand
type DeleteStockProductConfigResponse = httpserver.RestMutateResponse

type SetStockProductConfigArchivedRequest = itStock.SetStockProductConfigArchivedCommand
type SetStockProductConfigArchivedResponse = httpserver.RestMutateResponse

type GetStockProductConfigRequest = itStock.GetStockProductConfigByIdQuery
type GetStockProductConfigResponse = httpserver.RestGetOneResponse[dmodel.DynamicFields]

type StockProductConfigExistsRequest = itStock.StockProductConfigExistsQuery
type StockProductConfigExistsResponse = dyn.ExistsResultData

type SearchStockProductConfigsRequest = itStock.SearchStockProductConfigsQuery
type SearchStockProductConfigsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
