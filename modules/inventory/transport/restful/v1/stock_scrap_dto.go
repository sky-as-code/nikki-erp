package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

type CreateStockScrapRequest = itStock.CreateStockScrapCommand
type CreateStockScrapResponse = httpserver.RestCreateResponse

type UpdateStockScrapRequest = itStock.UpdateStockScrapCommand
type UpdateStockScrapResponse = httpserver.RestMutateResponse

type DeleteStockScrapRequest = itStock.DeleteStockScrapCommand
type DeleteStockScrapResponse = httpserver.RestMutateResponse

type SetStockScrapArchivedRequest = itStock.SetStockScrapArchivedCommand
type SetStockScrapArchivedResponse = httpserver.RestMutateResponse

type GetStockScrapRequest = itStock.GetStockScrapByIdQuery
type GetStockScrapResponse = httpserver.RestGetOneResponse[dmodel.DynamicFields]

type StockScrapExistsRequest = itStock.StockScrapExistsQuery
type StockScrapExistsResponse = dyn.ExistsResultData

type SearchStockScrapsRequest = itStock.SearchStockScrapsQuery
type SearchStockScrapsResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
