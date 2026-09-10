package v1

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

type CreateStockMoveRequest = itStock.CreateStockMoveCommand
type CreateStockMoveResponse = httpserver.RestCreateResponse

type UpdateStockMoveRequest = itStock.UpdateStockMoveCommand
type UpdateStockMoveResponse = httpserver.RestMutateResponse

type DeleteStockMoveRequest = itStock.DeleteStockMoveCommand
type DeleteStockMoveResponse = httpserver.RestMutateResponse

type SetStockMoveArchivedRequest = itStock.SetStockMoveArchivedCommand
type SetStockMoveArchivedResponse = httpserver.RestMutateResponse

type GetStockMoveRequest = itStock.GetStockMoveByIdQuery
type GetStockMoveResponse = httpserver.RestGetOneResponse[dmodel.DynamicFields]

type StockMoveExistsRequest = itStock.StockMoveExistsQuery
type StockMoveExistsResponse = dyn.ExistsResultData

type SearchStockMovesRequest = itStock.SearchStockMovesQuery
type SearchStockMovesResponse = httpserver.RestSearchResponse[dmodel.DynamicFields]
