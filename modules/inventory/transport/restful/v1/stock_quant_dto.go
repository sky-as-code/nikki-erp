package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

type CreateStockQuantRequest = itStock.CreateStockQuantCommand
type CreateStockQuantResponse = httpserver.RestCreateResponse

type UpdateStockQuantRequest = itStock.UpdateStockQuantCommand
type UpdateStockQuantResponse = httpserver.RestMutateResponse

type DeleteStockQuantRequest = itStock.DeleteStockQuantCommand
type DeleteStockQuantResponse = httpserver.RestMutateResponse

type SetStockQuantArchivedRequest = itStock.SetStockQuantArchivedCommand
type SetStockQuantArchivedResponse = httpserver.RestMutateResponse

type GetStockQuantRequest = itStock.GetStockQuantByIdQuery
type GetStockQuantResponse = httpserver.RestGetOneResponse[dmodel.DynamicFields]

type StockQuantExistsRequest = itStock.StockQuantExistsQuery
type StockQuantExistsResponse = dyn.ExistsResultData

type SearchStockQuantsRequest = itStock.SearchStockQuantsQuery
type SearchStockQuantsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
